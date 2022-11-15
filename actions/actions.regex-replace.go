package actions

import (
	"errors"
	"fmt"
	ics "github.com/arran4/golang-ical"
	"regexp"
	"strconv"
	"strings"
)

type RegexReplaceAction struct{}

func (rra *RegexReplaceAction) Identifier() string {
	return "actions/regex-replace"
}

///

// patternReplace replaces all occurrences of the pattern in the orig string.
// The replacement can contain ${i} to replace with sub-match.
// TODO: Fix this, it doesn't work correctly.
// Hello! with match = (?i)(?:H(e)llo)|(?:Wo(r)ld), repl = [$1] should output [e]
// but it doesn't.
func patternReplace(orig, replacement string, pattern *regexp.Regexp) string {
	// plain replace
	repl := pattern.ReplaceAllString(orig, replacement)
	// advanced replace with $1, $2, ...
	if strings.Contains(replacement, "$") {
		sub := pattern.FindAllStringSubmatch(orig, 10)
		if len(sub) > 0 {
			for i, v := range sub[0] {
				repl = strings.ReplaceAll(repl, "$"+strconv.Itoa(i+1), v)
			}
		}
	}
	return repl
}

func strArray(with map[string]interface{}, key string, def []interface{}) ([]string, error) {
	in, err := optional[[]interface{}](with, key, def)
	if err != nil {
		return nil, err
	}
	var res = make([]string, len(in))
	for i, v := range in {
		str, ok := v.(string)
		if !ok {
			return nil, errors.New(fmt.Sprintf("'%v' is not a string", v))
		}
		res[i] = str
	}
	return res, nil
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

func (rra *RegexReplaceAction) Execute(event *ics.VEvent, with map[string]interface{}) (ActionMessage, error) {
	match, err := required[string](with, "match")
	if err != nil {
		return nil, err
	}

	// replacement (can contain group-placeholders like $1, $2, ...)
	repl, err := required[string](with, "replace")
	if err != nil {
		return nil, err
	}

	// if pattern was marked as not-case-sensitive
	// append (?i) flag to pattern
	cs, err := optional[bool](with, "case-sensitive", false)
	if err != nil {
		return nil, err
	}
	if !cs {
		match = "(?i)" + match
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

	// compile given RegEx pattern
	pattern, err := regexp.Compile(match)
	if err != nil {
		return nil, err
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
				upd := patternReplace(v, repl, pattern)
				if upd != v {
					fmt.Printf("[actions/regex-replace] ~Parameter[%s] '%s' --> '%s'\n",
						param, v, upd)
					values[i] = upd
					save = true
				}
			}
		}

		upd := val.Value

		// read (and replace) property value itself
		// if either no parameter was specified or the '+' flag was used.
		if len(parameters) <= 0 || parametersIncludeSelf {
			upd = patternReplace(val.Value, repl, pattern)

			// only print if something changed
			if val.Value != upd {
				fmt.Printf("[actions/regex-replace] ~Parameter(%s) '%s' --> '%s'\n",
					strings.ToUpper(s), val.Value, upd)
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
