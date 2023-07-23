package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
)

func removeDuplicates(elements []string) []string {
	encountered := map[string]bool{}
	result := []string{}

	for _, element := range elements {
		if encountered[element] == true {
			continue
		}

		encountered[element] = true
		result = append(result, element)
	}

	return result
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	dir := "illustrations"
	// Read all xml in directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Panicln(err)
	}

	//colorScheme := make(map[string][]string)
	colorScheme := make(map[string]int)
	colorsPerIllustration := make(map[string]int)

	for _, entry := range entries {
		log.Printf("Reading %s\n", entry.Name())

		// Read the file as xml
		byt, err := os.ReadFile(dir + "/" + entry.Name())
		if err != nil {
			log.Panicln(err)
		}

		svgText := string(byt)

		doubles := make(map[string]bool)

		// Find all fill attributes in the svg
		reg := regexp.MustCompile(`"(#[a-f0-9]{3,8}|none)"`)
		matches := reg.FindAllStringSubmatch(svgText, -1)
		for _, match := range matches {
			color := match[1]

			if doubles[color] {
				continue
			}

			//if _, ok := colorScheme[color]; !ok {
			//	colorScheme[color] = []string{}
			//}

			//colorScheme[color] = append(colorScheme[color], entry.Name())
			colorScheme[color] += 1
			doubles[color] = true
			colorsPerIllustration[entry.Name()] += 1
		}

		//break
	}

	//// Remove duplicates
	//for k, v := range colorScheme {
	//	colorScheme[k] = removeDuplicates(v)
	//}

	byt, _ := json.MarshalIndent(colorScheme, "", "  ")
	fmt.Println(string(byt))

	err = os.WriteFile("colors.json", byt, 0644)
	if err != nil {
		log.Panicln(err)
	}

	byt, _ = json.MarshalIndent(colorsPerIllustration, "", "  ")
	fmt.Println(string(byt))

	err = os.WriteFile("colors_illustration.json", byt, 0644)
	if err != nil {
		log.Panicln(err)
	}
}
