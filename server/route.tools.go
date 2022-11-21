package server

import (
	"encoding/base64"
	"github.com/gofiber/fiber/v2"
	"gopkg.in/yaml.v3"
)

func (d *DemoServer) routeMinifyYaml(ctx *fiber.Ctx) error {
	var cnt []*yaml.Node
	if err := yaml.Unmarshal(ctx.Body(), &cnt); err != nil {
		// TODO: close body?
		return err
	}
	for _, c := range cnt {
		// remove all comments
		c.FootComment = ""
		c.LineComment = ""
		c.HeadComment = ""
	}
	data, err := yaml.Marshal(cnt)
	if err != nil {
		return err
	}

	var res string
	if q := ctx.Query("base64", "false"); q == "true" {
		res = base64.StdEncoding.EncodeToString(data)
	} else {
		res = string(data)
	}

	return ctx.SendString(res)
}
