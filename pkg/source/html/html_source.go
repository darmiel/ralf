package htmlsource

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	ics "github.com/darmiel/golang-ical"
	"github.com/ralf-life/engine/internal/util"
	httpsource "github.com/ralf-life/engine/pkg/source/http"
	"golang.org/x/net/html/charset"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Error declarations to handle various validation failures.
var (
	ErrSelectorsRequired   = errors.New("selectors are required")
	ErrParentRequired      = errors.New("parent is required")
	ErrStartRequired       = errors.New("start is required")
	ErrStartFormatRequired = errors.New("start format is required")
	ErrEndRequired         = errors.New("end is required")
	ErrEndFormatRequired   = errors.New("end format is required")
	ErrTitleRequired       = errors.New("summary is required")
)

// Error declarations to handle various parsing failures.
var (
	ErrReadingResponse = errors.New("unable to read the response body")
	ErrEmptyParent     = errors.New("empty parent")
	ErrEmptySelection  = errors.New("empty selection")
	ErrSkip            = errors.New("skip")
)

// Selector defines the fields required to extract specific elements from HTML.
type Selector struct {
	// Parent is the parent selector
	Parent string `json:"parent" yaml:"parent" bson:"parent"`

	// Start is the start date selector
	Start string `json:"start" yaml:"start" bson:"start"`

	// StartFormat is the start date format
	StartFormat string `json:"start_format" yaml:"start_format" bson:"start_format"`

	// End is the end date selector
	End string `json:"end" yaml:"end" bson:"end"`

	// EndFormat is the end date format
	EndFormat string `json:"end_format" yaml:"end_format" bson:"end_format"`

	// All is a flag to select all elements, this defaults to false
	// (optional)
	All bool `json:"all" yaml:"all" bson:"all"`

	// Soft is a flag to ignore errors, this defaults to false
	// (optional)
	Soft bool `json:"soft" yaml:"soft" bson:"soft"`

	// Summary is the title selector
	// (optional)
	Summary string `json:"summary" yaml:"summary" bson:"summary"`

	// Description is the description selector
	// (optional)
	Description string `json:"description" yaml:"description" bson:"description"`

	// Location is the location selector
	// (optional)
	Location string `json:"location" yaml:"location" bson:"location"`

	// URL is the URL selector
	// (optional)
	URL string `json:"url" yaml:"url" bson:"url"`

	// Organizer is the organizer selector
	// (optional)
	Organizer string `json:"organizer" yaml:"organizer" bson:"organizer"`

	// Status is the status selector
	// (optional)
	Status string `json:"status" yaml:"status" bson:"status"`
}

// Validate ensures that mandatory fields in Selector are properly set.
func (s *Selector) Validate() error {
	if s.Parent == "" {
		return ErrParentRequired
	}
	if s.Start == "" {
		return ErrStartRequired
	}
	if s.StartFormat == "" {
		return ErrStartFormatRequired
	}
	if s.End == "" {
		return ErrEndRequired
	}
	if s.EndFormat == "" {
		return ErrEndFormatRequired
	}
	if s.Summary == "" {
		return ErrTitleRequired
	}
	return nil
}

// Options defines the configuration for HTML source fetching and parsing.
type Options struct {
	httpsource.Options `json:",inline" yaml:",inline" bson:",inline"`

	// Name is the name of the output calendar
	Name string `json:"name" yaml:"name" bson:"name"`
	// Description is the description of the output calendar
	Description string `json:"description" yaml:"description" bson:"description"`
	// TZID is the timezone ID to use
	TZID string `json:"tzid" yaml:"tzid" bson:"tzid"`
	// Selectors are the HTML selectors to use
	Selectors []Selector `json:"selectors" yaml:"selectors" bson:"selectors"`
}

// KeyIdentifier returns the key identifier for the source
func (o *Options) KeyIdentifier() string {
	return "html"
}

// Validate validates the source options
func (o *Options) Validate() error {
	if err := o.Options.Validate(); err != nil {
		return err
	}
	if len(o.Selectors) == 0 {
		return ErrSelectorsRequired
	}
	for _, s := range o.Selectors {
		if err := s.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// CacheKey returns the cache key for the source
func (o *Options) CacheKey() (string, error) {
	return util.CreateCacheKey(o)
}

// Run executes the source
func (o *Options) Run() (*ics.Calendar, error) {
	resp, err := o.Options.MakeRequest()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseResponse(resp, o)
}

func (o *Options) String() string {
	return fmt.Sprintf("HTML Source: %s", o.URL)
}

// parseResponse reads and parses the HTML document from the response.
func parseResponse(resp *http.Response, o *Options) (*ics.Calendar, error) {
	utf8Body, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(utf8Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrReadingResponse, err)
	}
	return buildCalendar(doc, o)
}

// buildCalendar constructs an iCalendar object from the parsed HTML.
func buildCalendar(doc *goquery.Document, o *Options) (*ics.Calendar, error) {
	calendar := ics.NewCalendar()
	setCalendarProperties(calendar, o)

	var count uint = 0
	for _, selector := range o.Selectors {
		if err := parseEvents(doc, calendar, selector, &count); err != nil {
			return nil, err
		}
	}

	return calendar, nil
}

// setCalendarProperties sets the properties of the calendar from Options.
func setCalendarProperties(calendar *ics.Calendar, o *Options) {
	if o.Name != "" {
		calendar.SetName(o.Name)
	}
	if o.Description != "" {
		calendar.SetDescription(o.Description)
	}
	if o.TZID != "" {
		calendar.SetTzid(o.TZID)
	}
}

// parseEvents creates and adds events to the calendar based on the selector configuration.
func parseEvents(doc *goquery.Document, calendar *ics.Calendar, selector Selector, count *uint) error {
	args := strings.Split(selector.Parent, ">")
	for i, arg := range args {
		args[i] = strings.TrimSpace(arg)
	}

	parents := doc.Find(selector.Parent)
	if len(parents.Nodes) == 0 {
		return fmt.Errorf("%w: %s", ErrEmptyParent, selector.Parent)
	}

	var err error
	parents.EachWithBreak(func(_ int, selection *goquery.Selection) bool {
		*count++

		event := ics.NewEvent(strconv.Itoa(int(*count)))
		if err = assignEventDetails(event, selection, selector); err != nil {
			if errors.Is(err, ErrSkip) {
				err = nil
				return selector.All
			}
			return false
		}

		calendar.AddVEvent(event)
		return selector.All // continue with other elements if All is set
	})

	return err
}

func assignOptional(
	path string,
	setter func(s string, props ...ics.PropertyParameter),
	root *goquery.Selection,
	selector *Selector,
) error {
	if path == "" {
		return nil
	}
	text := root.Find(path).Text()
	if text == "" {
		if !selector.Soft {
			return fmt.Errorf("%w: %s", ErrEmptySelection, path)
		}
		return nil
	}
	setter(text)
	return nil
}

// assignEventDetails assigns details from HTML to the event fields.
func assignEventDetails(event *ics.VEvent, root *goquery.Selection, selector Selector) error {
	startText := root.Find(selector.Start).Text()
	if startText == "" {
		if !selector.Soft {
			return fmt.Errorf("%w: %s", ErrEmptySelection, selector.Start)
		}
		return ErrSkip
	}
	startDate, err := time.Parse(selector.StartFormat, startText)
	if err != nil {
		return err
	}

	endText := root.Find(selector.End).Text()
	if endText == "" {
		if !selector.Soft {
			return fmt.Errorf("%w: %s", ErrEmptySelection, selector.End)
		}
		return ErrSkip
	}
	if startText == endText {
		// if the start and end dates are the same, we assume it's a 1-full day event
		event.SetAllDayStartAt(startDate)
		event.SetAllDayEndAt(startDate)
	} else {
		var endDate time.Time
		if endDate, err = time.Parse(selector.EndFormat, endText); err != nil {
			return err
		}

		event.SetStartAt(startDate)
		event.SetEndAt(endDate)
	}

	if err = assignOptional(selector.Summary, event.SetSummary, root, &selector); err != nil {
		return err
	}
	if err = assignOptional(selector.Description, event.SetDescription, root, &selector); err != nil {
		return err
	}
	if err = assignOptional(selector.Location, event.SetLocation, root, &selector); err != nil {
		return err
	}
	if err = assignOptional(selector.URL, event.SetURL, root, &selector); err != nil {
		return err
	}
	if err = assignOptional(selector.Organizer, event.SetOrganizer, root, &selector); err != nil {
		return err
	}
	if err = assignOptional(selector.Status, func(s string, props ...ics.PropertyParameter) {
		event.SetStatus(ics.ObjectStatus(s))
	}, root, &selector); err != nil {
		return err
	}

	return nil
}
