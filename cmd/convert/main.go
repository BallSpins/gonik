package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

type TargetData struct {
	Type       string
	PostalCode string
	Name       string
}

func main() {
	jsonPath := "./data/wilayah.json"
	outputPath := "./data/wilayah.bin"

	fileData, err := os.ReadFile(jsonPath)
	if err != nil {
		fmt.Printf("Error: failed to read file %s\n", jsonPath)
		return
	}

	var rawJson map[string]map[string]string
	if err := json.Unmarshal(fileData, &rawJson); err != nil {
		fmt.Println("Error: invalid JSON format")
		return
	}

	rawData := make(map[string]TargetData)

	// 1. Provinsi
	if provinsi, ok := rawJson["provinsi"]; ok {
		for k, v := range provinsi {
			rawData[k+"----"] = TargetData{Type: "P", PostalCode: "     ", Name: v}
		}
	}

	// 2. Kabkot
	if kabkot, ok := rawJson["kabkot"]; ok {
		for k, v := range kabkot {
			rawData[k+"--"] = TargetData{Type: "K", PostalCode: "     ", Name: v}
		}
	}

	// 3. Kecamatan
	if kecamatan, ok := rawJson["kecamatan"]; ok {
		for k, v := range kecamatan {
			parts := strings.Split(v, " -- ")
			name := strings.TrimSpace(parts[0])
			postal := "     "
			if len(parts) > 1 {
				// Ambil kode pos pertama jika berupa range
				rawPostal := strings.TrimSpace(strings.Split(parts[1], "-")[0])
				if isNumeric(rawPostal) {
					postal = fmt.Sprintf("%-5s", rawPostal[:min(5, len(rawPostal))])
				}
			}
			rawData[k] = TargetData{Type: "C", PostalCode: postal, Name: name}
		}
	}

	// Sortir Key secara alfabetis
	keys := make([]string, 0, len(rawData))
	for k := range rawData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	outFunc, err := os.Create(outputPath)
	if err != nil {
		fmt.Println("Error: failed to create wilayah.bin")
		return
	}
	defer outFunc.Close()

	var totalBytes int
	for _, code := range keys {
		info := rawData[code]

		nameClean := strings.TrimSpace(info.Name)
		postalClean := strings.TrimSpace(info.PostalCode)

		row := fmt.Sprintf("%s|%s|%s|%s\n", code, info.Type, postalClean, nameClean)

		n, err := outFunc.WriteString(row)
		if err != nil {
			fmt.Println("Error: failed to write row")
			return
		}
		totalBytes += n
	}

	fmt.Printf("Conversion complete! %s ditulis (%d bytes)\n", outputPath, len(rawData)*50)
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
