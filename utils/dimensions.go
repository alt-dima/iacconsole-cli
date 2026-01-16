package utils

import (
	"log"
	"strings"
)

func (s *State) ParseDimensions() {
	parsedDimArgs := parseDimArgs(s.DimensionsFlags)

	for _, dimension := range s.UnitManifest.Dimensions {
		if _, ok := parsedDimArgs[dimension]; !ok {
			log.Fatalln("dimension " + dimension + " not passed with -d arg")
		}
	}

	s.ParsedDimensions = parsedDimArgs
}

func parseDimArgs(dimensionsArgs []string) map[string]string {
	parsedDimArgs := make(map[string]string)
	for _, dimension := range dimensionsArgs {
		dimensionSlice := strings.SplitN(dimension, ":", 2)
		if len(dimensionSlice) != 2 {
			log.Fatalln("Invalid dimension format: " + dimension + ". Expected format: key:value")
		}
		if strings.HasPrefix(dimensionSlice[1], "dim_") {
			log.Fatalln("dimension " + dimension + " with dim_ prefix can't be passed with -d arg")
		}
		parsedDimArgs[dimensionSlice[0]] = dimensionSlice[1]
	}
	return parsedDimArgs
}
