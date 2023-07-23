package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Illustration struct {
	Id    string `json:"_id"`
	Title string `json:"title"`
	Image string `json:"image"`
	Slug  string `json:"slug"`
}

type ApiResponse struct {
	Illustrations []Illustration `json:"illos"`
	HasMore       bool           `json:"hasMore"`
	NextPage      int            `json:"nextPage"`
}

func getIllustrations() (illustrations []Illustration) {
	baseUrl := "https://undraw.co/api/illustrations?page="

	for inx := 0; ; inx++ {
		log.Printf("Downloading page %d\n", inx)

		// Get the response
		response, err := http.Get(baseUrl + strconv.Itoa(inx))
		if err != nil {
			log.Fatalln(err)
		}

		// Read the response body
		data, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatalln(err)
		}

		// Unmarshal the data
		var apiResponse ApiResponse
		err = json.Unmarshal(data, &apiResponse)
		if err != nil {
			log.Fatalln(err)
		}

		illustrations = append(illustrations, apiResponse.Illustrations...)

		if apiResponse.HasMore == false {
			break
		}
	}

	for inx, illustration := range illustrations {
		illustrations[inx].Id = kebabCase(illustration.Title)
	}

	sort.Slice(illustrations, func(i, j int) bool {
		return illustrations[i].Id < illustrations[j].Id
	})

	return
}

func kebabCase(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "-", "_")

	startNum := regexp.MustCompile(`^\d+`)
	value = startNum.ReplaceAllStringFunc(value, func(match string) string {
		return "_" + match
	})
	value = strings.Replace(value, "void", "void_", 1)
	return value
}

func downloadIllustrations(illustrations []Illustration) (downloads []string) {

	illustrationDir := "illustrations"
	err := os.MkdirAll(illustrationDir, 0755)
	if err != nil {
		log.Fatalln(err)
	}

	for _, illustration := range illustrations {

		name := kebabCase(illustration.Title)

		log.Printf("Downloading illustration %s\n", name)

		// Get the response
		response, err := http.Get(illustration.Image)
		if err != nil {
			log.Fatalln(err)
		}

		// Read the response body
		data, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatalln(err)
		}

		// Save the file
		filename := illustrationDir + "/" + name + ".svg"
		err = os.WriteFile(filename, data, 0644)
		if err != nil {
			log.Fatalln(err)
		}

		downloads = append(downloads, filename)
	}

	return downloads
}

func updateLib(illustrations []Illustration, downloads []string) {
	dart := `// ignore_for_file: unused_field
			/// Enums to help locate the correct illustration
			enum UnDrawIllustration {`

	for _, illustration := range illustrations {
		dart += "/// Title: " + illustration.Title + "\n"
		dart += "/// <br/>\n"
		dart += "/// <img src=\"" + illustration.Image + "\" alt=\"" + illustration.Title + "\" width=\"200\"/>\n"
		dart += illustration.Id + ",\n"
	}

	dart += `}`

	dart += `/// Map of illustrations with url to download
			const Map<UnDrawIllustration, String> unDrawIllustrations = {`

	for inx, filename := range downloads {
		illustration := illustrations[inx]
		dart += "UnDrawIllustration." + illustration.Id + ": " + `"` + filename + `",`
	}

	dart += `};`

	err := os.WriteFile("lib/illustrations.g.dart", []byte(dart), 0644)
	if err != nil {
		log.Fatalln(err)
	}

	// Execute dart format on the file
	err = exec.Command("dart", "format", "lib/illustrations.g.dart").Run()
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	illustrations := getIllustrations()
	log.Printf("Collected %d illustrations\n", len(illustrations))

	downloads := downloadIllustrations(illustrations)
	updateLib(illustrations, downloads)
}
