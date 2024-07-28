package main

import (
	"fmt"
	htmlsource "github.com/ralf-life/engine/pkg/source/html"
	httpsource "github.com/ralf-life/engine/pkg/source/http"
)

func main() {
	options := htmlsource.Options{
		Options: httpsource.Options{
			URL: "https://www.boulderwelt-karlsruhe.de/routenbau/",
		},
		Selectors: []htmlsource.Selector{
			{
				Parent: "#tablepress-20 > tbody > tr",
				All:    true,
				Soft:   true,

				Start:       ".column-1",
				StartFormat: "02.01.2006",
				End:         ".column-1",
				EndFormat:   "02.01.2006",
				Summary:     ".column-2",
			},
		},
	}
	fmt.Println(options.CacheKey())
	if err := options.Validate(); err != nil {
		panic(err)
	}
	fmt.Println(options.Run())
}
