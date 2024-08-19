package v1

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/darmiel/ralf/cmd/server/constraints"
	"github.com/darmiel/ralf/cmd/server/service"
	"strconv"
	"strings"
	"time"

	"github.com/darmiel/golang-ical"
	"github.com/darmiel/ralf/pkg/engine"
	"github.com/darmiel/ralf/pkg/model"
	"github.com/gofiber/fiber/v2"
	"gopkg.in/yaml.v3"
)

// PublicRoutes holds the dependencies for the /process routes.
type PublicRoutes struct {
	cacheService service.CacheService
}

// NewPublicRoutes creates a new PublicRoutes with the necessary dependencies.
func NewPublicRoutes(cacheService service.CacheService) *PublicRoutes {
	return &PublicRoutes{
		cacheService: cacheService,
	}
}

// processGetHandler is the handler for the GET /process route.
func (p *PublicRoutes) processGetHandler(c *fiber.Ctx) error {
	q := c.Query("template")
	if q == "" {
		return fiber.NewError(fiber.StatusBadRequest, "`template` (base64-encoded) parameter missing.")
	}
	content, err := base64.StdEncoding.DecodeString(q)
	if err != nil {
		return fiber.NewError(fiber.StatusExpectationFailed, "invalid base64 ("+err.Error()+")")
	}
	return p.processContent(content, c)
}

// processPostHandler is the handler for the POST /process route.
func (p *PublicRoutes) processPostHandler(c *fiber.Ctx) error {
	return p.processContent(c.Body(), c)
}

// processContent processes the content and returns the resulting calendar.
func (p *PublicRoutes) processContent(content []byte, ctx *fiber.Ctx) error {
	debug := ctx.QueryBool("debug", true)
	verbose := ctx.QueryBool("verbose", false)

	// Parse the profile
	var profile model.Profile
	dec := yaml.NewDecoder(bytes.NewReader(content))
	dec.KnownFields(true)
	if err := dec.Decode(&profile); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid RALF-SPEC ("+err.Error()+")")
	}

	// validate profile
	if len(profile.Source) != 1 {
		return fiber.NewError(fiber.StatusBadRequest, "only one source is allowed")
	}
	profile.CacheDuration = constraints.ClampCacheModelDuration(profile.CacheDuration)

	cal, debugMessages, err := p.executeFlow(ctx.Context(), &profile, debug, verbose)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Append debug messages as header
	ctx.Append("X-Debug-Message-Count", strconv.Itoa(len(debugMessages)))
	for i, v := range debugMessages {
		ctx.Append(fmt.Sprintf("X-Debug-Message-%d", i+1), fmt.Sprintf("%+v", v))
	}

	// Append content-type and return calendar
	ctx.Set("Content-Type", "text/calendar")
	return ctx.Status(201).SendString(cal.Serialize())
}

// executeFlow executes the flow for the given profile and returns the resulting calendar.
func (p *PublicRoutes) executeFlow(
	ctx context.Context,
	profile *model.Profile,
	debug, verbose bool,
) (*ics.Calendar, []string, error) {
	source := profile.Source[0]
	if err := source.Validate(); err != nil {
		return nil, nil, fmt.Errorf("source could not be validated: %w", err)
	}

	cal, err := p.getSource(ctx, source, time.Duration(profile.CacheDuration))
	if err != nil {
		return nil, nil, fmt.Errorf("cannot request source: %w", err)
	}

	// Create context and run flow
	cp := &engine.ContextFlow{
		Profile:     profile,
		Context:     make(map[string]interface{}),
		EnableDebug: debug,
		Verbose:     verbose,
	}

	debugMessages := make([]string, len(cp.Debugs))
	for i, v := range cp.Debugs {
		debugMessages[i] = fmt.Sprintf("%+v", v)
	}

	if err = engine.ModifyCalendar(cp, profile.Flows, cal); err != nil {
		return nil, debugMessages, fmt.Errorf("failed to run flow: %w", err)
	}

	return cal, debugMessages, nil
}

// getSource retrieves the source from the cache or requests it from the source.
func (p *PublicRoutes) getSource(
	ctx context.Context,
	source model.Source,
	cacheDuration time.Duration,
) (*ics.Calendar, error) {
	cacheKey, err := source.CacheKey()
	if err != nil {
		return nil, err
	}

	body, err := p.cacheService.Get(ctx, cacheKey)
	if err != nil || body == "" {
		return p.getSourceWithRequest(ctx, source, cacheDuration, cacheKey)
	}

	return ics.ParseCalendar(strings.NewReader(body))
}

// getSourceWithRequest requests the source and caches it.
func (p *PublicRoutes) getSourceWithRequest(
	ctx context.Context,
	source model.Source,
	cacheDuration time.Duration,
	cacheKey string,
) (*ics.Calendar, error) {
	cal, err := source.Run()
	if err != nil {
		return nil, err
	}
	if err = p.cacheService.Set(ctx, cacheKey, cal.Serialize(), int64(cacheDuration.Seconds())); err != nil {
		return nil, err
	}
	return cal, nil
}

func (p *PublicRoutes) routeMinifyYAML(ctx *fiber.Ctx) error {
	var content []*yaml.Node
	if err := yaml.Unmarshal(ctx.Body(), &content); err != nil {
		return err
	}
	for _, c := range content {
		c.FootComment = ""
		c.LineComment = ""
		c.HeadComment = ""
	}
	data, err := yaml.Marshal(content)
	if err != nil {
		return err
	}
	var res string
	if ctx.QueryBool("base64", false) {
		res = base64.StdEncoding.EncodeToString(data)
	} else {
		res = string(data)
	}
	return ctx.SendString(res)
}
