package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

func (s *State) GenerateVarsByDims() error {
	for dimKey, dimValue := range s.ParsedDimensions {
		dimensionJsonMap, err := s.GetDimData(dimKey, dimValue, false)
		if err != nil {
			return err
		}

		targetAutoTfvarMap := map[string]interface{}{
			"iacconsole_" + dimKey + "_data": dimensionJsonMap,
			"iacconsole_" + dimKey + "_name": dimValue,
		}

		if err := writeTfvarsMaps(targetAutoTfvarMap, dimKey, s.CmdWorkTempDir); err != nil {
			return err
		}
		log.Println("attached dimension in var.iacconsole_" + dimKey + "_data and var.iacconsole_" + dimKey + "_name")

	}
	return nil
}

func (s *State) GenerateVarsByDimOptional(optionType string) error {
	for dimKey := range s.ParsedDimensions {
		dimensionJsonMap, err := s.GetDimData(dimKey, "dim_"+optionType, true)
		if err != nil {
			return err
		}
		if len(dimensionJsonMap) > 0 {
			targetAutoTfvarMap := map[string]interface{}{
				"iacconsole_" + dimKey + "_" + optionType: dimensionJsonMap,
			}

			if err := writeTfvarsMaps(targetAutoTfvarMap, dimKey+"_"+optionType, s.CmdWorkTempDir); err != nil {
				return err
			}
			log.Println("attached " + optionType + " in var.iacconsole_" + dimKey + "_" + optionType)
		}
	}
	return nil
}

func (s *State) GenerateVarsByDimAndData(optionType string, dimKey string, dimensionJsonMap map[string]interface{}) error {
	targetAutoTfvarMap := map[string]interface{}{
		"iacconsole_" + dimKey + "_" + optionType: dimensionJsonMap,
	}
	if err := writeTfvarsMaps(targetAutoTfvarMap, dimKey+"_"+optionType, s.CmdWorkTempDir); err != nil {
		return err
	}
	log.Println("attached " + optionType + " in var.iacconsole_" + dimKey + "_" + optionType)
	return nil
}

func (s *State) GenerateVarsByEnvVars() error {
	targetAutoTfvarMap := make(map[string]interface{})

	for _, envVar := range os.Environ() {
		if strings.HasPrefix(envVar, "iacconsole_envvar_") {
			envVarList := strings.SplitN(envVar, "=", 2)
			targetAutoTfvarMap[envVarList[0]] = envVarList[1]
			log.Println("attached env variable in var." + envVarList[0])
		}
	}

	if len(targetAutoTfvarMap) > 0 {
		if err := writeTfvarsMaps(targetAutoTfvarMap, "envivars", s.CmdWorkTempDir); err != nil {
			return err
		}
	}
	return nil
}

func writeTfvarsMaps(targetAutoTfvarMap map[string]interface{}, fileName string, cmdWorkTempDir string) error {
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

	if err := marshalJsonAndWrite(targetVarsTfMapFull, targetVarsTfPath); err != nil {
		return err
	}
	if err := marshalJsonAndWrite(targetAutoTfvarMap, targetAutoTfvarsPath); err != nil {
		return err
	}
	return nil
}

func marshalJsonAndWrite(jsonMap map[string]interface{}, jsonPath string) error {
	targetAutoTfvarMapBytes, err := json.Marshal(jsonMap)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %v", err)
	}
	err = os.WriteFile(jsonPath, targetAutoTfvarMapBytes, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}
	return nil
}
