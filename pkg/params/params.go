package params

import (
	"bbox/pkg/types"
	"fmt"
	"github.com/pkg/errors"
	"regexp"
	"strings"
)

// ParseCombinations parses the combinations from the command line and returns a slice of BuildParameters
func ParseCombinations(combinations []string) ([]types.BuildParameters, error) {
	parsed := make([]types.BuildParameters, 0, len(combinations))
	for _, combo := range combinations {
		parts := strings.Split(combo, ";")
		if len(parts) != 4 {
			return nil, fmt.Errorf("invalid combination format: %s", combo)
		}

		if !isValidBuildID(parts[0]) {
			return nil, fmt.Errorf("invalid buildTypeID: %s", parts[0])
		}

		if !isValidBranchName(parts[1]) {
			return nil, fmt.Errorf("invalid branchName: %s", parts[1])
		}

		if parts[2] != "true" && parts[2] != "false" {
			return nil, fmt.Errorf("invalid downloadArtifacts boolean: %s", parts[2])
		}

		downloadArtifacts, valid := isValidDownloadArtifacts(parts[2])

		if !valid {
			return nil, fmt.Errorf("invalid downloadArtifacts boolean: %s", parts[2])
		}

		properties, err := parseProperties(parts[3])

		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse properties: %s", parts[3])
		}

		parsed = append(parsed, types.BuildParameters{
			BuildTypeId:       parts[0],
			BranchName:        parts[1],
			DownloadArtifacts: downloadArtifacts,
			PropertiesFlag:    properties,
		})
	}

	return parsed, nil
}

func isValidBuildID(buildID string) bool {
	matched, _ := regexp.MatchString("^[a-zA-Z0-9-_]+$", buildID)
	return matched
}

func isValidBranchName(branchName string) bool {
	matched, _ := regexp.MatchString("^[^ ~^:\\\\?*[\\]@{}/][^ ~^:\\\\?*[\\]@{}]*$", branchName)
	return matched
}

// isValidDownloadArtifacts checks if the downloadArtifacts string is either "true" or "false", case-insensitively.
// Returns a boolean indicating if the value is "true" or "false" and a second boolean validity flag.
func isValidDownloadArtifacts(downloadArtifacts string) (bool, bool) {
	normalized := strings.ToLower(downloadArtifacts)
	if normalized == "true" {
		return true, true
	} else if normalized == "false" {
		return false, true
	}

	return false, false
}

// parseProperties parses the properties from the command line and returns a map of string to string
func parseProperties(properties string) (map[string]string, error) {
	propertiesMap := make(map[string]string)
	for _, prop := range strings.Split(properties, "&") {
		kv := strings.SplitN(prop, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid property format: %s", prop)
		}
		key, value := kv[0], kv[1]
		if !validateParamKey(key) {
			return nil, fmt.Errorf("invalid property key: %s", key)
		}
		if !validateParamValue(value) {
			return nil, fmt.Errorf("invalid property value: %s", value)
		}
		propertiesMap[key] = value
	}
	return propertiesMap, nil
}

// validateParamKey checks if the parameter is valid key and returns a boolean
func validateParamKey(key string) bool {
	matched, _ := regexp.MatchString(`^\w+[a-zA-Z0-9\\;,*/_.-]*`, key)
	return matched
}

// validateParamValue checks if the parameter is valid value and returns a boolean
func validateParamValue(value string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9\\;,*/@:_.-]*$`, value)
	return matched
}
