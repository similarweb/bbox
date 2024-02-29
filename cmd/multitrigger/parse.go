package multitrigger

import (
	"bbox/pkg/params"
	"bbox/pkg/types"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

const combinationPartsNumber = 4

// parseCombinations parses the combinations from the command line and returns a slice of BuildParameters.
func parseCombinations(combinations []string) ([]types.BuildParameters, error) {
	parsed := make([]types.BuildParameters, 0, len(combinations))

	for _, combo := range combinations {
		parts := strings.Split(combo, ";")
		if len(parts) != combinationPartsNumber {
			log.Errorf("invalid combination format: %s. expected: 'buildTypeID;branchName;downloadArtifactsBool;key1=value1&key2=value2'", combo)
			return nil, fmt.Errorf("invalid combination format: %s", combo)
		}

		if !params.IsValidBuildID(parts[0]) {
			return nil, fmt.Errorf("invalid buildTypeID: %s", parts[0])
		}

		if !params.IsValidBranchName(parts[1]) {
			return nil, fmt.Errorf("invalid branchName: %s", parts[1])
		}

		downloadArtifacts, valid := strconv.ParseBool(parts[2])

		if valid != nil {
			return nil, fmt.Errorf("invalid downloadArtifacts boolean: %s", parts[2])
		}

		properties, err := parseProperties(parts[3])
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse properties: %s", parts[3])
		}

		parsed = append(parsed, types.BuildParameters{
			BuildTypeID:       parts[0],
			BranchName:        parts[1],
			DownloadArtifacts: downloadArtifacts,
			PropertiesFlag:    properties,
		})
	}

	return parsed, nil
}

// parseProperties parses the properties from the command line and returns a map of string to string.
func parseProperties(properties string) (map[string]string, error) {
	propertiesMap := make(map[string]string)

	for _, prop := range strings.Split(properties, "&") {
		kv := strings.SplitN(prop, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid property format: %s", prop)
		}

		key, value := kv[0], kv[1]

		if !params.ValidateParamKey(key) {
			return nil, fmt.Errorf("invalid property key: %s", key)
		}

		if !params.ValidateParamValue(value) {
			return nil, fmt.Errorf("invalid property value: %s", value)
		}

		propertiesMap[key] = value
	}

	return propertiesMap, nil
}
