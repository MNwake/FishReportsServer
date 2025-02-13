package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"fishreports/model"
)

// ✅ Load Counties from JSON File
var counties []model.County

// LoadCounties reads the counties JSON file and returns the slice of counties.
func LoadCounties(filePath string) ([]model.County, error) {
	var counties []model.County
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read county data file: %w", err)
	}
	err = json.Unmarshal(file, &counties)
	if err != nil {
		return nil, fmt.Errorf("failed to parse county JSON: %w", err)
	}
	return counties, nil
}

// ✅ Load Fish Survey Data
func LoadFishData(m *model.FishSurveyModel, syncDir string) error {
	m.FishDataByCounty = make(map[string][]model.FishData)
	fileChan := make(chan string, 100)
	var wg sync.WaitGroup

	worker := func() {
		for path := range fileChan {
			if surveys, err := processFile(m, path); err != nil {
				log.Printf("❌ Error processing file: %s - %v", path, err)
			} else {
				log.Printf("✅ Processed %d surveys from %s", surveys, path)
			}
			wg.Done()
		}
	}

	numWorkers := 8
	for i := 0; i < numWorkers; i++ {
		go worker()
	}

	err := filepath.WalkDir(syncDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Printf("❌ Error accessing %s: %v", path, err)
			return nil
		}
		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}
		wg.Add(1)
		fileChan <- path
		return nil
	})

	close(fileChan)
	wg.Wait()
	return err
}

func processFile(m *model.FishSurveyModel, path string) (int, error) {
	m.Mutex.Lock()  // ✅ Lock before modifying shared data
	defer m.Mutex.Unlock()  // ✅ Unlock after modification

	fileData, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	// Step 1: Read raw JSON into map
	var rawData map[string]interface{}
	if err := json.Unmarshal(fileData, &rawData); err != nil {
		return 0, err
	}

	// Step 2: Transform fishCount safely
	TransformFishCount(rawData, m)

	// Step 3: Convert transformed JSON back into FishData struct
	transformedJSON, _ := json.Marshal(rawData)
	var fishData model.FishData
	if err := json.Unmarshal(transformedJSON, &fishData); err != nil {
		return 0, err
	}

	// Step 4: Safely store data in the map
	m.FishDataByCounty[fishData.Result.CountyName] = append(m.FishDataByCounty[fishData.Result.CountyName], fishData)

	// ✅ Return number of surveys processed
	return len(fishData.Result.Surveys), nil
}


func TransformFishCount(data map[string]interface{}, m *model.FishSurveyModel) {

	for key, value := range data {
		if key == "fishCount" {
			if list, ok := value.([]interface{}); ok {
				transformed := []map[string]int{}
				for _, pair := range list {
					if pairList, ok := pair.([]interface{}); ok && len(pairList) == 2 {
						transformed = append(transformed, map[string]int{
							"length":   int(pairList[0].(float64)), // Convert float64 to int
							"quantity": int(pairList[1].(float64)),
						})
					}
				}
				data[key] = transformed
				log.Printf("✅ Transformed fishCount: %+v\n", transformed)
			}
		} else if nested, ok := value.(map[string]interface{}); ok {
			TransformFishCount(nested, m) // 🔄 Recursively transform nested objects
		} else if nestedList, ok := value.([]interface{}); ok {
			for _, item := range nestedList {
				if itemMap, ok := item.(map[string]interface{}); ok {
					TransformFishCount(itemMap, m)
				}
			}
		}
	}
}

// ✅ Load Species Mapping
func LoadSpeciesMap(m *model.FishSurveyModel, speciesFile string) error {
	file, err := os.ReadFile(speciesFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(file, &m.SpeciesMap)
	if err != nil {
		return err
	}

	// Capitalize common names after loading
	for code, species := range m.SpeciesMap {
		species.CommonName = capitalizeFirst(species.CommonName)
		m.SpeciesMap[code] = species
	}

	log.Printf("✅ Loaded %d species from %s", len(m.SpeciesMap), speciesFile)
	return nil
}

// Helper function to capitalize first letter of each word
func capitalizeFirst(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[0:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}

// ✅ Get All Counties
func GetAllCounties() []model.County {
	return counties
}