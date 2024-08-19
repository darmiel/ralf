package v1

import (
	"github.com/darmiel/ralf/cmd/server/constraints"
	"github.com/darmiel/ralf/cmd/server/logging"
	"github.com/darmiel/ralf/cmd/server/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"time"
)

var (
	ErrFlowNotFound            = fiber.NewError(fiber.StatusNotFound, "flow not found")
	ErrFlowDoesNotBelongToUser = fiber.NewError(fiber.StatusUnauthorized, "flow does not belong to user")
)

type AuthorizedFlowRoutes struct {
	storageService service.StorageService
	processRoutes  *PublicRoutes
	authService    service.AuthService
}

func NewAuthorizedFlowRoutes(
	storageService service.StorageService,
	processRoutes *PublicRoutes,
	authService service.AuthService,
) *AuthorizedFlowRoutes {
	return &AuthorizedFlowRoutes{
		storageService: storageService,
		processRoutes:  processRoutes,
		authService:    authService,
	}
}

// getUserFlowsHandler handles GET /flows.
func (f *AuthorizedFlowRoutes) getUserFlowsHandler(c *fiber.Ctx) error {
	u := c.Locals("user").(*service.AuthUser)

	flows, err := f.storageService.GetFlowsByUser(c.Context(), u.UserID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "cannot retrieve flows: "+err.Error())
	}

	return c.Status(200).JSON(flows)
}

// saveFlowHandler handles POST /flows.
func (f *AuthorizedFlowRoutes) saveFlowHandler(c *fiber.Ctx) error {
	u := c.Locals("user").(*service.AuthUser)

	var flow service.SavedFlow
	if err := c.BodyParser(&flow); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request payload: "+err.Error())
	}

	var action string

	// if a flow ID is provided, make sure it belongs to the user
	if flow.FlowID != "" {
		action = "update"

		// get the flow
		oldFlow, err := f.storageService.GetFlow(c.Context(), flow.FlowID)
		if err != nil {
			return ErrFlowNotFound
		}
		// make sure the flow belongs to the user
		if oldFlow.UserID != u.UserID {
			return ErrFlowDoesNotBelongToUser
		}
	} else {
		action = "create"

		// if no flow ID is provided (create new flow), find new free ID (max tries: 10)
		for i := 0; i < 10; i++ {
			flow.FlowID = uuid.New().String()
			if _, err := f.storageService.GetFlow(c.Context(), flow.FlowID); err != nil {
				break
			}
		}
	}

	flow.UserID = u.UserID
	flow.CacheDuration = constraints.ClampCacheModelDuration(flow.CacheDuration)

	if err := f.storageService.SaveFlow(c.Context(), &flow); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "cannot save flow: "+err.Error())
	}

	h := service.History{
		FlowID:    flow.FlowID,
		Address:   c.IP(),
		Timestamp: time.Now(),
		Success:   true,
		Action:    action,
	}
	if err := f.storageService.SaveHistory(c.Context(), h); err != nil {
		logger := logging.GetLoggerFromContext(c)
		logger.Warn("cannot save history", zap.String("action", action), zap.Error(err))
	}

	return c.Status(200).SendString("SavedFlow saved")
}

func (f *AuthorizedFlowRoutes) flowLocalAccessCheckMiddleware(c *fiber.Ctx) error {
	u := c.Locals("user").(*service.AuthUser)
	flowID := c.Params("flow")

	flow, err := f.storageService.GetFlow(c.Context(), flowID)
	if err != nil {
		return ErrFlowNotFound
	}
	if flow.UserID != u.UserID {
		return ErrFlowDoesNotBelongToUser
	}

	return c.Next()
}

// deleteFlowHandler handles DELETE /:flow.
func (f *AuthorizedFlowRoutes) deleteFlowHandler(c *fiber.Ctx) error {
	flow := c.Locals("flow").(*service.SavedFlow)

	if err := f.storageService.DeleteFlow(c.Context(), flow.FlowID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "cannot delete flow: "+err.Error())
	}

	h := service.History{
		FlowID:    flow.FlowID,
		Address:   c.IP(),
		Timestamp: time.Now(),
		Success:   true,
		Action:    "delete",
	}
	if err := f.storageService.SaveHistory(c.Context(), h); err != nil {
		logger := logging.GetLoggerFromContext(c)
		logger.Warn("cannot save [delete] history", zap.Error(err))
	}

	return c.Status(200).SendString("SavedFlow deleted")
}

// getFlowJSONHandler is the handler for GET /:flow.json.
func (f *AuthorizedFlowRoutes) getFlowJSONHandler(c *fiber.Ctx) error {
	flow := c.Locals("flow").(*service.SavedFlow)
	return c.Status(200).JSON(flow)
}

// getFlowHistoryHandler is the handler for GET /:flow/history.
func (f *AuthorizedFlowRoutes) getFlowHistoryHandler(c *fiber.Ctx) error {
	flow := c.Locals("flow").(*service.SavedFlow)

	limit := constraints.ClampHistoryLimit(c.QueryInt("limit", 100))

	history, err := f.storageService.GetFlowHistory(c.Context(), flow.FlowID, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "cannot retrieve history: "+err.Error())
	}

	if len(history) > limit {
		history = history[:limit]
	}

	return c.Status(200).JSON(history)
}
