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
	publicRoutes := NewPublicRoutes(cacheService)
	router.Get("/process", publicRoutes.processGetHandler)
	router.Post("/process", publicRoutes.processPostHandler)
	router.Post("/minify", publicRoutes.routeMinifyYAML)

	// for the next routes, the storage service is required
	if storageService == nil {
		return
	}

	flowRoutes := NewFlowRoutes(storageService, publicRoutes)
	router.Get("/:flow.ics", flowRoutes.getFlowICSHandler)

	// for the next routes, also the auth service is required
	if authService == nil {
		return
	}

	router.Use(authService.Middleware())
	authFlowRoutes := NewAuthorizedFlowRoutes(storageService, publicRoutes, authService)
	router.Get("/flows", authFlowRoutes.getUserFlowsHandler)
	router.Post("/flows", authFlowRoutes.saveFlowHandler)

	router.Use("/:flow", authFlowRoutes.flowLocalAccessCheckMiddleware)
	router.Delete("/:flow", authFlowRoutes.deleteFlowHandler)
	router.Get("/:flow/raw", authFlowRoutes.getFlowJSONHandler)
	router.Get("/:flow/history", authFlowRoutes.getFlowHistoryHandler)
}
