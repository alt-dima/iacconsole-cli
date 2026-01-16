package utils

import (
	"encoding/json"
	"log"
	"os"
)

func (s *State) ParseUnitManifest(unitManifestFileName string) {
	unitManifestPath := s.UnitPath + "/" + unitManifestFileName
	// Let's first read the `config.json` file
	content, err := os.ReadFile(unitManifestPath)
	if err != nil {
		log.Fatal("iacconsole error when opening file: ", err)
	}

	// Now let's unmarshall the data into `payload`
	var unitManifest unitManifestStruct
	err = json.Unmarshal(content, &unitManifest)
	if err != nil {
		log.Fatal("iacconsole error during Unmarshal(): ", err)
	}

	s.UnitManifest = unitManifest
	log.Println("iacconsole loaded unit manifest: " + unitManifestPath)
}
