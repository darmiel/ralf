package v1

import (
	"github.com/darmiel/ralf/cmd/server/service"
	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(
	router fiber.Router,
	storageService service.StorageService,
	authService service.AuthService,
	cacheService service.CacheService,
) {
	// register the routes for the /process endpoint.
	// these routes can be used without authentication.
	processRoutes := NewProcessRoutes(cacheService)
	router.Get("/process", processRoutes.processGetHandler)
	router.Post("/process", processRoutes.processPostHandler)

	// for the next routes, the storage service is required
	if storageService == nil {
		return
	}

	flowRoutes := NewFlowRoutes(storageService, processRoutes)
	router.Get("/:flow.ics", flowRoutes.getFlowICSHandler)
	router.Get("/:flow.json", flowRoutes.getFlowJSONHandler)
	router.Get("/:flow/history", flowRoutes.getFlowHistoryHandler)

	// for the next routes, also the auth service is required
	if authService == nil {
		return
	}

	router.Use(authService.Middleware())
	authFlowRoutes := NewAuthorizedFlowRoutes(storageService, processRoutes, authService)
	router.Post("/flows", authFlowRoutes.saveFlowHandler)
	router.Delete("/:flow", authFlowRoutes.deleteFlowHandler)
	router.Get("/flows", authFlowRoutes.getUserFlowsHandler)
}
