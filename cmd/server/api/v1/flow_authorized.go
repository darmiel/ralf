package v1

import (
	"github.com/darmiel/ralf/cmd/server/constraints"
	"github.com/darmiel/ralf/cmd/server/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"time"
)

type AuthorizedFlowRoutes struct {
	storageService service.StorageService
	processRoutes  *ProcessRoutes
	authService    service.AuthService
}

func NewAuthorizedFlowRoutes(
	storageService service.StorageService,
	processRoutes *ProcessRoutes,
	authService service.AuthService,
) *AuthorizedFlowRoutes {
	return &AuthorizedFlowRoutes{
		storageService: storageService,
		processRoutes:  processRoutes,
		authService:    authService,
	}
}

// deleteFlowHandler handles DELETE /:flow.
func (f *AuthorizedFlowRoutes) deleteFlowHandler(c *fiber.Ctx) error {
	u := c.Locals("user").(*service.AuthUser)

	flowID := c.Params("flow")

	// make sure the flow exists and belongs to the user
	flow, err := f.storageService.GetFlow(c.Context(), flowID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "cannot find flow: "+err.Error())
	}
	if flow.UserID != u.UserID {
		return fiber.NewError(fiber.StatusUnauthorized, "flow does not belong to user")
	}

	if err = f.storageService.DeleteFlow(c.Context(), flow.FlowID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "cannot delete flow: "+err.Error())
	}

	h := service.History{
		FlowID:    flowID,
		Address:   c.IP(),
		Timestamp: time.Now(),
		Success:   true,
		Action:    "delete",
	}
	if err = f.storageService.SaveHistory(c.Context(), h); err != nil {
		log.Warnf("cannot save DELETE history: %s", err.Error())
	}

	return c.Status(200).SendString("SavedFlow deleted")
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

	flow.UserID = u.UserID
	if flow.FlowID == "" {
		flow.FlowID = uuid.New().String()
	}
	flow.CacheDuration = constraints.ClampCacheModelDuration(flow.CacheDuration)

	if err := f.storageService.SaveFlow(c.Context(), &flow); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "cannot save flow: "+err.Error())
	}

	h := service.History{
		FlowID:    flow.FlowID,
		Address:   c.IP(),
		Timestamp: time.Now(),
		Success:   true,
		Action:    "update",
	}
	if err := f.storageService.SaveHistory(c.Context(), h); err != nil {
		log.Warnf("cannot save UPDATE history: %s", err.Error())
	}

	return c.Status(200).SendString("SavedFlow saved")
}
