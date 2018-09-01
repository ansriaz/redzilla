package k8s

import (
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	regex         = regexp.MustCompile("{{\\s*(?P<variable>\\w+)\\s*}}")
	variableRegex = regexp.MustCompile("^\\$([a-zA-Z_]\\w*)$")
)

func ParseTemplate(template string, substitutions map[string]interface{}) (string, error) {
	res := regex.ReplaceAllStringFunc(template, func(match string) string {

		template_variable_name := strings.ToLower(regex.FindStringSubmatch(match)[1])

		subst := substitutions[template_variable_name] //FIXME if there are spaces this won't work

		// log.Debugf("matched %s, replace: %s", match, subst)
		if subst == "" {
			log.Warnf("No substitute candidate in template for %s", match)
			// TODO change this to multiple returns of errors in the main function
		}
		return fmt.Sprintf("%v", subst)
	})

	return res, nil
}

func refineMap(substitutions map[string]interface{}, variables map[string]interface{}) (map[string]interface{}, error) {
	res := make(map[string]interface{})

	for field, value := range substitutions {
		match := variableRegex.FindStringSubmatch(fmt.Sprintf("%v", value))

		if len(match) > 1 {
			if v := variables[match[1]]; v != nil {
				// log.Debugf("Sust %s matches variable format (%s) %s", field, value, v)
				res[strings.ToLower(field)] = v
			}
		}

		if res[strings.ToLower(field)] == nil {
			res[strings.ToLower(field)] = value
		}
	}

	return res, nil
}
