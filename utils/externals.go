package utils

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/viper"
)

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (s *State) GetStringFromViperByOrgOrDefault(keyName string) string {
	if viper.IsSet(s.OrgName + "." + keyName) {
		return viper.GetString(s.OrgName + "." + keyName)
	} else {
		return viper.GetString("defaults." + keyName)
	}
}

func (s *State) GetObjectFromViperByOrgOrDefault(keyName string) map[string]any {
	if viper.IsSet(s.OrgName + "." + keyName) {
		return viper.GetStringMap(s.OrgName + "." + keyName)
	} else {
		return viper.GetStringMap("defaults." + keyName)
	}
}

func (s *State) SetupBackendConfig() map[string]interface{} {
	var stateS3Path string
	if !viper.IsSet(s.OrgName + ".backend") {
		stateS3Path = stateS3Path + "org_" + s.OrgName + "/"
	}

	for _, dimension := range s.UnitManifest.Dimensions {
		stateS3Path = stateS3Path + dimension + "_" + s.ParsedDimensions[dimension] + "/"
	}
	s.StateS3Path = stateS3Path + s.UnitName + ".tfstate"

	backendConfig := s.GetObjectFromViperByOrgOrDefault("backend")
	if len(backendConfig) == 0 {
		log.Println("no backend config provied!")
	}

	var backendConfigMap = make(map[string]interface{}, len(backendConfig))
	for param, value := range backendConfig {
		backendConfigMap[param] = strings.Replace(value.(string), "$iacconsole_state_path", s.StateS3Path, 1)
	}

	return backendConfigMap
}

func (s *State) GetDimData(dimensionKey string, dimensionValue string, skipOnNotFound bool) (map[string]interface{}, error) {
	var dimensionJsonMap map[string]interface{}

	if s.IacconsoleApiUrl == "" {
		inventroyJsonPath := s.InventoryPath + "/" + dimensionKey + "/" + dimensionValue + ".json"
		dimensionJsonBytes, err := os.ReadFile(inventroyJsonPath)
		if err != nil {
			if os.IsNotExist(err) && skipOnNotFound {
				log.Println("inventory files: Optional dimension " + s.OrgName + "/" + dimensionKey + "/" + dimensionValue + " not found, skipping")
				return dimensionJsonMap, nil
			}
			return nil, err
		}
		err = json.Unmarshal(dimensionJsonBytes, &dimensionJsonMap)
		if err != nil {
			return nil, err
		}
	} else {
		resp, err := http.Get(s.IacconsoleApiUrl + "/v1/dimension/" + s.OrgName + "/" + dimensionKey + "/" + dimensionValue + "?workspace=" + s.Workspace + "&fallbacktomaster=true")
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 404 {
			resp.Body.Close()
			if skipOnNotFound {
				log.Println("optional dimension " + s.OrgName + "/" + dimensionKey + "/" + dimensionValue + " not found, skipping")
				return dimensionJsonMap, nil
			}
			log.Println("requested dimension not found response 404:" + resp.Request.URL.String())
			return nil, fmt.Errorf("dimension %s/%s/%s not found", s.OrgName, dimensionKey, dimensionValue)
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			return nil, fmt.Errorf("request %s/%s/%s?workspace=%s failed with response: %v", s.OrgName, dimensionKey, dimensionValue, s.Workspace, resp.StatusCode)
		}
		defer resp.Body.Close()

		dimensionJsonBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("reading body response failed: %s", err)
		}

		var IacConsoleDBResponse IaCConsoleDBResponse
		err = json.Unmarshal(dimensionJsonBytes, &IacConsoleDBResponse)
		if err != nil {
			return nil, fmt.Errorf("error during unmarshal json response: %v", err)
		}

		if len(IacConsoleDBResponse.Dimensions) != 1 {
			return nil, fmt.Errorf("should be only one dimension in response")
		}
		if IacConsoleDBResponse.Error != "" {
			log.Println(IacConsoleDBResponse.Error)
		}
		dimensionJsonMap = IacConsoleDBResponse.Dimensions[0].DimData

	}

	return dimensionJsonMap, nil
}
