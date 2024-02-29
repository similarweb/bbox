package params

import (
	"regexp"
)

// IsValidBuildID checks if the buildID is valid and returns a boolean. The buildID can contain letters, numbers, hyphens, and underscores.
func IsValidBuildID(buildID string) bool {
	matched, _ := regexp.MatchString("^[a-zA-Z0-9-_]+$", buildID)
	return matched
}

// IsValidBranchName checks if the branchName is valid and returns a boolean.
func IsValidBranchName(branchName string) bool {
	matched, _ := regexp.MatchString("^[^ ~^:\\\\?*[\\]@{}/][^ ~^:\\\\?*[\\]@{}]*$", branchName)
	return matched
}

// ValidateParamKey checks if the parameter is valid key and returns a boolean.
func ValidateParamKey(key string) bool {
	matched, _ := regexp.MatchString(`^\w+[a-zA-Z0-9\\;,*/_.-]*`, key)
	return matched
}

// ValidateParamValue checks if the parameter is valid value and returns a boolean.
func ValidateParamValue(value string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9\\;,*/@:_.-]*$`, value)
	return matched
}
