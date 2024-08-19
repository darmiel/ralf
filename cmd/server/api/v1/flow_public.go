package v1

import (
	"fmt"
	"github.com/darmiel/ralf/cmd/server/constraints"
	"github.com/darmiel/ralf/cmd/server/service"
	"github.com/darmiel/ralf/pkg/model"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

type FlowRoutes struct {
	storageService service.StorageService
	publicRoutes   *PublicRoutes
}

// NewFlowRoutes creates a new FlowRoutes with the necessary dependencies.
func NewFlowRoutes(storageService service.StorageService, processRoutes *PublicRoutes) *FlowRoutes {
	return &FlowRoutes{
		storageService: storageService,
		publicRoutes:   processRoutes,
	}
}

// getFlowICSHandler is the handler for GET /:flow.ics.
func (f *FlowRoutes) getFlowICSHandler(ctx *fiber.Ctx) error {
	flowID := ctx.Params("flow")
	verbose := ctx.QueryBool("verbose", false)
	debug := ctx.QueryBool("debug", true)

	flow, err := f.storageService.GetFlow(ctx.Context(), flowID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "cannot find flow: "+err.Error())
	}

	profile := model.Profile{
		Name:          flow.Name,
		Source:        flow.Source,
		CacheDuration: constraints.ClampCacheModelDuration(flow.CacheDuration),
		Flows:         flow.Flows,
	}

	cal, debugMessages, err := f.publicRoutes.executeFlow(ctx.Context(), &profile, debug, verbose)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Append debug messages as header
	ctx.Append("X-Debug-Message-Count", strconv.Itoa(len(debugMessages)))
	for i, v := range debugMessages {
		ctx.Append(fmt.Sprintf("X-Debug-Message-%d", i+1), v)
	}

	// Append content-type and return calendar
	ctx.Set("Content-Type", "text/calendar")
	return ctx.Status(200).SendString(cal.Serialize())
}
