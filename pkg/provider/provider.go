package provider

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kcloutie/knot/pkg/template"
	"go.uber.org/zap"
)

func HasRequiredProperties(properties map[string]string, requiredProperties []string) (bool, error) {
	missingProperties := []string{}
	hasMissing := false
	for _, propName := range requiredProperties {
		_, exists := properties[propName]
		if !exists {
			missingProperties = append(missingProperties, propName)
			hasMissing = true
		}
	}

	if hasMissing {
		return hasMissing, fmt.Errorf("missing the following required properties: %s", strings.Join(missingProperties, ", "))
	}
	return hasMissing, nil
}

func SetGoTemplateOptionValues(log *zap.Logger, config *template.RenderTemplateOptions, properties map[string]string) {
	propVal, exists := properties[template.LeftDelimPropertyName]
	if exists {
		config.LeftDelim = propVal
	}

	propVal, exists = properties[template.RightDelimPropertyName]
	if exists {
		config.RightDelim = propVal
	}

	propVal, exists = properties[template.IgnoreTemplateErrors]
	if exists {
		boolValue, err := strconv.ParseBool(propVal)
		if err != nil {
			log.Warn(fmt.Sprintf("invalid value of '%s' specified for the %s property. Expected true or false", propVal, template.IgnoreTemplateErrors))
			return
		}
		config.IgnoreTemplateErrors = boolValue
	}
}
