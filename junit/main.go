package main

import (
	"bytes"
	"encoding/json"
	"html"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/joho/godotenv"
)

type SubStep struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type Step struct {
	Name     string    `json:"name"`
	SubSteps []SubStep `json:"tests,omitempty"`
}

type Response struct {
	Token string `json:"token"`
	Steps []Step `json:"steps"`
}

func parseXML(path string) *xmlquery.Node {
	content, err_file := os.ReadFile(path)
	if err_file != nil {
		log.Println("Erreur lors de la lecture du fichier:", err_file)
		return nil
	}

	doc, err_parsing := xmlquery.Parse(strings.NewReader(string(content)))
	if err_parsing != nil {
		panic(err_parsing)
	}

	return doc
}

func createSubStep(testcase *xmlquery.Node) SubStep {
	var passed bool = true
	var failureMessage string = ""
	var errorMessage string = ""

	failure := testcase.SelectElement("failure")
	if failure != nil {
		passed = false
		errorMessage = failure.SelectAttr("message")
		failureMessage = failure.InnerText()
	}

	var name string = testcase.SelectAttr("name")
	failureMessage = html.UnescapeString(failureMessage)

	var subStep SubStep = SubStep{Name: name, Passed: passed, Message: failureMessage, Error: errorMessage}

	return subStep
}

func main() {
	godotenv.Load()
	var dirPath string = os.Getenv("FOLDER_PATH")

	var steps []Step

	files, err := os.ReadDir(dirPath)

	if err != nil {
		log.Println("Erreur lors de l'ouverture du dossier:", err)
		return
	}

	for _, file := range files {
		filePath := filepath.Join(dirPath, file.Name())
		if !file.IsDir() && filepath.Ext(file.Name()) == ".xml" {
			var subSteps []SubStep

			doc := parseXML(filePath)

			step_, _ := xmlquery.Query(doc, "//testcase")
			name := step_.SelectAttr("classname")

			testcases, _ := xmlquery.QueryAll(doc, "//testcase")

			for _, testcase := range testcases {

				subStep := createSubStep(testcase)

				subSteps = append(subSteps, subStep)
			}
			step := Step{Name: name, SubSteps: subSteps}
			steps = append(steps, step)
		}
	}

	response := Response{Token: os.Getenv("CALLBACK_TOKEN"), Steps: steps}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	_ = enc.Encode(response)

	responseJSON := buf.String()

	data := bytes.NewBuffer([]byte(responseJSON))
	url := os.Getenv("WEBHOOK_URL")
	req, _ := http.NewRequest("POST", url, data)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Erreur lors de l'envoi de la requête :", err)
		return
	}
	defer resp.Body.Close()

}
