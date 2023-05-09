package server

import (
	"github.com/go-redis/redis/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type DemoServer struct {
	app *fiber.App
	red *redis.Client
}

func (d *DemoServer) Start() error {
	return d.app.Listen(":1887")
}

func New(rc *redis.Client) *DemoServer {
	app := fiber.New()
	d := &DemoServer{
		app: app,
		red: rc,
	}

	app.Post("/tools/minify", d.routeMinifyYaml)

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))
	app.Post("/process", d.routeProcessPost)
	app.Get("/process", d.routeProcessGet)

	return d
}
