package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/redis/go-redis/v9"
)

type DemoServer struct {
	app *fiber.App
	red *redis.Client
}

func (d *DemoServer) Start() error {
	return d.app.Listen(":1887")
}

type info struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

func New(rc *redis.Client, version, commit, date string) *DemoServer {
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
	app.Get("/icanhasralf", func(ctx *fiber.Ctx) error {
		return ctx.JSON(&info{
			Version: version,
			Commit:  commit,
			Date:    date,
		})
	})

	return d
}
