package actions

import (
	"errors"
	"fmt"
	ics "github.com/darmiel/golang-ical"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrNoReplacerSpecified          = errors.New("specify replacers with either `match` and `replace` or `map`")
	ErrInvalidReplacementMapNotList = errors.New("invalid replacement-map, should be a list")
	ErrInvalidReplacementMapNotMap  = errors.New("invalid replacement-map, should contain at least 1 element")
	ErrInvalidReplacementMapNotKey  = errors.New("invalid replacement-map, should contain `match` and `replace`")
)

const CaseSensitiveKey = "case-sensitive"

type RegexReplaceAction struct{}

func (rra *RegexReplaceAction) Identifier() string {
	return "actions/regex-replace"
}

///

type mapReplacers struct {
	// Match matches a string based on a regular expression
	Match string
	// Replacement is the replacement (can contain group-placeholders like $1, $2, ...)
	Replacement string
	// Pattern is the compiled Match-Pattern and filled later
	Pattern *regexp.Regexp
}

// Do replaces all occurrences of the pattern in the orig string.
// The replacement can contain ${i} to replace with sub-match.
// TODO: Fix this, it doesn't work correctly.
// Hello! with match = (?i)(?:H(e)llo)|(?:Wo(r)ld), repl = [$1] should output [e]
// but it doesn't.
func (m *mapReplacers) Do(orig string) string {
	// plain replace
	repl := m.Pattern.ReplaceAllString(orig, m.Replacement)
	// advanced replace with $1, $2, ...
	if strings.Contains(m.Replacement, "$") {
		sub := m.Pattern.FindAllStringSubmatch(orig, 10)
		if len(sub) > 0 {
			for i, v := range sub[0] {
				repl = strings.ReplaceAll(repl, "$"+strconv.Itoa(i+1), v)
			}
		}
	}
	return repl
}

// wrap ICalParameters (map[string][]string) to ics.PropertyParameter
func mapToKV(m map[string][]string) []ics.PropertyParameter {
	res := make([]ics.PropertyParameter, len(m))
	i := 0
	for k, v := range m {
		res[i] = &ics.KeyValues{
			Key:   k,
			Value: v,
		}
		i += 1
	}
	return res
}

func (rra *RegexReplaceAction) Execute(event *ics.VEvent, with map[string]interface{}, verbose bool) (ActionMessage, error) {
	// replace map
	var replacers []*mapReplacers

	globalCaseSensitive, err := optional[bool](with, CaseSensitiveKey, true)
	if err != nil {
		return nil, err
	}

	if has(with, "match") && has(with, "replace") {
		match, err := required[string](with, "match")
		if err != nil {
			return nil, err
		}
		repl, err := required[string](with, "replace")
		if err != nil {
			return nil, err
		}
		// if pattern was marked globally as not-case-sensitive
		// append (?i) flag to pattern to sub-patterns
		if !globalCaseSensitive {
			match = "(?i)" + match
		}
		// single replacer
		replacers = append(replacers, &mapReplacers{
			Match:       match,
			Replacement: repl,
		})
	} else if mapTypeRaw, ok := with["map"]; ok {
		// m should be of type map[string]interface{}
		mapRaw, ok := mapTypeRaw.([]interface{})
		if !ok {
			return nil, ErrInvalidReplacementMapNotList
		}
		for _, v := range mapRaw {
			// should be a map
			m, ok := v.(map[string]interface{})
			if !ok {
				return nil, ErrInvalidReplacementMapNotMap
			}
			if !has(m, "match") || !has(m, "replace") {
				return nil, ErrInvalidReplacementMapNotKey
			}
			match, err := required[string](m, "match")
			if err != nil {
				return nil, err
			}
			repl, err := required[string](m, "replace")
			if err != nil {
				return nil, err
			}
			localCaseSensitive, err := optional[bool](m, CaseSensitiveKey, true)
			if err != nil {
				return nil, err
			}
			if !localCaseSensitive || (!globalCaseSensitive && !has(m, CaseSensitiveKey)) {
				match = "(?i)" + match
			}
			replacers = append(replacers, &mapReplacers{
				Match:       match,
				Replacement: repl,
			})
		}
	} else {
		return nil, ErrNoReplacerSpecified
	}

	// if no "in" was specified, by default, replace in description
	// some ics-events contain specific parameters, for example CN for ORGANIZER
	// to target these parameters, you can use the / syntax:
	// "ORGANIZER/CN", to target multiple parameters, separate by comma:
	// "ORGANIZER/CN,MAIL,STATUS". This, however, ONLY addresses the parameters.
	// If you also want to filter for the ORGANIZER-value itself, append a plus sign at the property name:
	// "ORGANIZER+/CN" which will check ORGANIZER and ORGANIZER/CN
	in, err := strArray(with, "in", []interface{}{"description"})
	if err != nil {
		return nil, err
	}

	// compile RegEx patterns
	for _, v := range replacers {
		if v.Pattern, err = regexp.Compile(v.Match); err != nil {
			return nil, fmt.Errorf("cannot compile expression '%s': %v",
				v.Match, err)
		}
	}

	// execute replace
	for _, s := range in {
		var (
			// specific targeted ics-parameters
			parameters []string
			// also include the value for replacement?
			// ORGANIZER/CN only checks and replaces the CN parameter.
			// set this to true, to also include the ORGANIZER parameter.
			// this behaviour changes when appending the '+' sign to the property name, like
			// ORGANIZER+/CN
			parametersIncludeSelf bool
			// marks the event to have changes and to update the property
			save bool
		)
		if strings.Contains(s, "/") {
			spl := strings.SplitN(s, "/", 2)
			s, parameters = spl[0], strings.Split(spl[1], ",")
		}

		// this check could be moved to the /-check, but to avoid confusion when forgetting the plus sign,
		// just remove it in case of no parameters found.
		if strings.HasSuffix(s, "+") {
			s = s[:len(s)-1]
			parametersIncludeSelf = true
		}

		// check if the property even exists
		prop := ics.ComponentProperty(strings.ToUpper(s))
		val := event.GetProperty(prop)
		if val == nil {
			continue
		}

		// read (and replace) parameters
	ppp:
		for _, param := range parameters {
			values, ok := val.ICalParameters[param]
			if !ok {
				continue ppp
			}
			for i, v := range values {
				upd := v
				for _, r := range replacers {
					upd = r.Do(upd)
				}
				if upd != v {
					if verbose {
						fmt.Printf("[actions/regex-replace] ~Param[%s] '%s' changed to '%s'\n",
							param, v, upd)
					}
					values[i] = upd
					save = true
				}
			}
		}

		upd := val.Value

		// read (and replace) property value itself
		// if either no parameter was specified or the '+' flag was used.
		if len(parameters) <= 0 || parametersIncludeSelf {
			for _, r := range replacers {
				upd = r.Do(upd)
			}

			// only print if something changed
			if val.Value != upd {
				if verbose {
					fmt.Printf("[actions/regex-replace] ~Param+(%s) '%s' changed to '%s'\n",
						strings.ToUpper(s), val.Value, upd)
				}
				save = true
			}
		}

		if save {
			// apply changes
			kv := mapToKV(val.ICalParameters)
			event.SetProperty(prop, upd, kv...)
		}
	}

	return nil, nil
}
