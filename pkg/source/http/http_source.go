package httpsource

import (
	"errors"
	"fmt"
	ics "github.com/darmiel/golang-ical"
	"github.com/ralf-life/engine/internal/util"
	"io"
	"net/http"
	"strings"
	"time"
)

const MaxContentLength = 8 * 1000 * 1000

var (
	ErrURLRequired = errors.New("URL is required")
)

const (
	DefaultMethod = "GET"
)

type Options struct {
	// URL to fetch
	URL string `json:"url" yaml:"url" bson:"url"`
	// Method to use
	Method string `json:"method" yaml:"method" bson:"method"`
	// Headers to send
	Headers map[string]string `json:"headers" yaml:"headers" bson:"headers"`
	// Body to send
	Body string `json:"body" yaml:"body" bson:"body"`
	// Timeout in seconds
	Timeout int `json:"timeout" yaml:"timeout" bson:"timeout"`
}

// KeyIdentifier returns the key identifier for the source
func (o *Options) KeyIdentifier() string {
	return "http"
}

// Validate validates the source options
func (o *Options) Validate() error {
	if o.URL == "" {
		return ErrURLRequired
	}
	return nil
}

// CacheKey returns the cache key for the source
func (o *Options) CacheKey() (string, error) {
	return util.CreateCacheKey(o)
}

// MakeRequest makes the HTTP request
func (o *Options) MakeRequest() (*http.Response, error) {
	method := o.Method
	if method == "" {
		method = DefaultMethod
	}
	var body io.Reader
	if o.Body != "" {
		body = strings.NewReader(o.Body)
	}
	request, err := http.NewRequest(method, o.URL, body)
	if err != nil {
		return nil, err
	}
	if o.Headers != nil {
		for k, v := range o.Headers {
			request.Header.Add(k, v)
		}
	}
	client := &http.Client{}
	if o.Timeout > 0 {
		client.Timeout = time.Duration(o.Timeout) * time.Second
	}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	resp.Body = http.MaxBytesReader(nil, resp.Body, MaxContentLength)
	return resp, err
}

// Run executes the source
func (o *Options) Run() (*ics.Calendar, error) {
	resp, err := o.MakeRequest()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ics.ParseCalendar(resp.Body)
}

func (o *Options) String() string {
	method := o.Method
	if method == "" {
		method = DefaultMethod
	}
	return fmt.Sprintf("%s %s", method, o.URL)
}
