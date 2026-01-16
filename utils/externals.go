package utils

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
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

func (s *State) GetDimData(dimensionKey string, dimensionValue string, skipOnNotFound bool) map[string]interface{} {
	var dimensionJsonMap map[string]interface{}

	if s.IacconsoleApiUrl == "" {
		inventroyJsonPath := s.InventoryPath + "/" + dimensionKey + "/" + dimensionValue + ".json"
		dimensionJsonBytes, err := os.ReadFile(inventroyJsonPath)
		if err != nil {
			if os.IsNotExist(err) && skipOnNotFound {
				log.Println("inventory files: Optional dimension " + s.OrgName + "/" + dimensionKey + "/" + dimensionValue + " not found, skipping")
				return dimensionJsonMap
			}
			log.Fatal("inventory files: error when opening dim file: ", err.Error())
		}
		err = json.Unmarshal(dimensionJsonBytes, &dimensionJsonMap)
		if err != nil {
			log.Fatal("error during Unmarshal(): ", err)
		}
	} else {
		resp, err := http.Get(s.IacconsoleApiUrl + "/v1/dimension/" + s.OrgName + "/" + dimensionKey + "/" + dimensionValue + "?workspace=" + s.Workspace + "&fallbacktomaster=true")
		if err != nil {
			log.Fatalf("request failed: %s", err)
		} else if resp.StatusCode == 404 {
			resp.Body.Close()
			if skipOnNotFound {
				log.Println("optional dimension " + s.OrgName + "/" + dimensionKey + "/" + dimensionValue + " not found, skipping")
				return dimensionJsonMap
			} else {
				log.Println("requested dimension not found response 404:" + resp.Request.URL.String())
				log.Fatalln("dimension " + s.OrgName + "/" + dimensionKey + "/" + dimensionValue + " not found")
			}
		} else if resp.StatusCode != 200 {
			resp.Body.Close()
			log.Fatalf("request "+s.OrgName+"/"+dimensionKey+"/"+dimensionValue+"?workspace="+s.Workspace+" failed with response: %v", resp.StatusCode)
		}
		defer resp.Body.Close()

		dimensionJsonBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("reading body response failed: %s", err)
		}

		var IacConsoleDBResponse IaCConsoleDBResponse
		err = json.Unmarshal(dimensionJsonBytes, &IacConsoleDBResponse)
		if err != nil {
			log.Fatal("error during unmarshal json response: ", err)
		}

		if len(IacConsoleDBResponse.Dimensions) != 1 {
			log.Fatalf("should be only one dimension in response")
		}
		if IacConsoleDBResponse.Error != "" {
			log.Println(IacConsoleDBResponse.Error)
		}
		dimensionJsonMap = IacConsoleDBResponse.Dimensions[0].DimData

	}

	return dimensionJsonMap
}
