package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"log"
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
	SubSteps []SubStep `json:"sub_steps,omitempty"`
}

type Response struct {
	Token    string `json:"token"`
	StepList []Step `json:"steps"`
}

func main() {
	godotenv.Load()
	var dirPath string = os.Getenv("FOLDER_PATH")

	var step_list []Step

	files, err := os.ReadDir(dirPath)

	if err != nil {
		log.Println("Erreur lors de l'ouverture du dossier:", err)
		return
	}

	for _, file := range files {
		file_path := filepath.Join(dirPath, file.Name())
		if !file.IsDir() && filepath.Ext(file.Name()) == ".xml" {
			var sub_steps []SubStep
			sub_passed := true

			content, err_file := os.ReadFile(file_path)
			if err_file != nil {
				log.Println("Erreur lors de la lecture du fichier:", err_file)
				continue
			}

			doc, err_parsing := xmlquery.Parse(strings.NewReader(string(content)))
			if err_parsing != nil {
				panic(err_parsing)
			}
			step_, _ := xmlquery.Query(doc, "//testcase")
			step_name := step_.SelectAttr("classname")

			testcase_list, _ := xmlquery.QueryAll(doc, "//testcase")

			for _, testcase := range testcase_list {
				var failure_message string = ""
				var error_message string = ""

				failure := testcase.SelectElement("failure")
				if failure != nil {
					sub_passed = false
					error_message = failure.SelectAttr("message")
					failure_message = failure.InnerText()
				} else {
					sub_passed = true
				}

				var name string = testcase.SelectAttr("name")
				failure_message = html.UnescapeString(failure_message)

				var sub_step SubStep = SubStep{Name: name, Passed: sub_passed, Message: failure_message, Error: error_message}
				sub_steps = append(sub_steps, sub_step)
			}
			step := Step{Name: step_name, SubSteps: sub_steps}
			step_list = append(step_list, step)
		}
	}
	response := Response{Token: "", StepList: step_list}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	_ = enc.Encode(response)

	responseJSON := buf.String()
	fmt.Print(responseJSON)

	/*
		data := bytes.NewBuffer([]byte(responseJSON))
		url := os.Getenv("CRACKITO_URL") + "/api/v1/endpoint/ci-result/" + os.Getenv("CALLBACK_TOKEN")
		req, _ := http.NewRequest("POST", url, data)
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println("Erreur lors de l'envoi de la requête :", err)
			return
		}
		defer resp.Body.Close()
	*/

}
