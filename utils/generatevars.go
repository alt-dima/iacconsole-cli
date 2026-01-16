package utils

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

func (s *State) GenerateVarsByDims() {
	for dimKey, dimValue := range s.ParsedDimensions {
		dimensionJsonMap := s.GetDimData(dimKey, dimValue, false)

		targetAutoTfvarMap := map[string]interface{}{
			"iacconsole_" + dimKey + "_data": dimensionJsonMap,
			"iacconsole_" + dimKey + "_name": dimValue,
		}

		writeTfvarsMaps(targetAutoTfvarMap, dimKey, s.CmdWorkTempDir)
		log.Println("attached dimension in var.iacconsole_" + dimKey + "_data and var.iacconsole_" + dimKey + "_name")

	}
}

func (s *State) GenerateVarsByDimOptional(optionType string) {
	for dimKey := range s.ParsedDimensions {
		dimensionJsonMap := s.GetDimData(dimKey, "dim_"+optionType, true)
		if len(dimensionJsonMap) > 0 {
			targetAutoTfvarMap := map[string]interface{}{
				"iacconsole_" + dimKey + "_" + optionType: dimensionJsonMap,
			}

			writeTfvarsMaps(targetAutoTfvarMap, dimKey+"_"+optionType, s.CmdWorkTempDir)
			log.Println("attached " + optionType + " in var.iacconsole_" + dimKey + "_" + optionType)
		}
	}
}

func (s *State) GenerateVarsByDimAndData(optionType string, dimKey string, dimensionJsonMap map[string]interface{}) {
	targetAutoTfvarMap := map[string]interface{}{
		"iacconsole_" + dimKey + "_" + optionType: dimensionJsonMap,
	}
	writeTfvarsMaps(targetAutoTfvarMap, dimKey+"_"+optionType, s.CmdWorkTempDir)
	log.Println("attached " + optionType + " in var.iacconsole_" + dimKey + "_" + optionType)
}

func (s *State) GenerateVarsByEnvVars() {
	targetAutoTfvarMap := make(map[string]interface{})

	for _, envVar := range os.Environ() {
		if strings.HasPrefix(envVar, "iacconsole_envvar_") {
			envVarList := strings.SplitN(envVar, "=", 2)
			targetAutoTfvarMap[envVarList[0]] = envVarList[1]
			log.Println("attached env variable in var." + envVarList[0])
		}
	}

	if len(targetAutoTfvarMap) > 0 {
		writeTfvarsMaps(targetAutoTfvarMap, "envivars", s.CmdWorkTempDir)
	}
}

func writeTfvarsMaps(targetAutoTfvarMap map[string]interface{}, fileName string, cmdWorkTempDir string) {
	targetVarsTfPath := cmdWorkTempDir + "/iacconsole_" + fileName + "_vars.tf.json"
	targetAutoTfvarsPath := cmdWorkTempDir + "/iacconsole_" + fileName + ".auto.tfvars.json"

	targetVarsTfMap := make(map[string]interface{})

	for key, value := range targetAutoTfvarMap {
		switch value.(type) {
		case string:
			targetVarsTfMap[key] = map[string]string{"type": "string"}
		default:
			targetVarsTfMap[key] = map[string]interface{}{}
		}

	}

	targetVarsTfMapFull := map[string]interface{}{
		"variable": targetVarsTfMap,
	}

	marshalJsonAndWrite(targetVarsTfMapFull, targetVarsTfPath)
	marshalJsonAndWrite(targetAutoTfvarMap, targetAutoTfvarsPath)
}

func marshalJsonAndWrite(jsonMap map[string]interface{}, jsonPath string) {
	targetAutoTfvarMapBytes, err := json.Marshal(jsonMap)
	if err != nil {
		log.Fatal("failed to marshal json: ", err)
	}
	err = os.WriteFile(jsonPath, targetAutoTfvarMapBytes, os.ModePerm)
	if err != nil {
		log.Fatal("error writing file: ", err)
	}
}
