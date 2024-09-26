package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"
)

//go:embed template/main.gotmpl
var mainFile []byte

//go:embed template/config.gotmpl
var configFile []byte

//go:embed template/service.gotmpl
var serviceFile []byte

//go:embed template/config.yaml
var configYamlFile []byte

type TemplateData struct {
	Name               string
	ServiceName        string
	ServiceType        string
	ServiceConstructor string
	ServiceConfigType  string
	ServiceConfigName  string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Missing project name")
		os.Exit(1)
	}
	projectName := os.Args[1]

	// check proper name
	onlyAlphaRegex := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9\-]+$`)
	if !onlyAlphaRegex.MatchString(projectName) {
		fmt.Println("Invalid project name, only alphanumeric characters and dashes allowed, first character must be a letter.")
		os.Exit(1)
	}

	// pascalize name
	splitWords := regexp.MustCompile(`([^\-][a-z0-9]*)`)
	results := splitWords.FindAllStringSubmatch(projectName, -1)
	words := strings.Builder{}
	for i := range results {
		if results[i][0][0] >= 97 && results[i][0][0] <= 122 {
			results[i][0] = fmt.Sprintf("%s%s", string(results[i][0][0]-32), results[i][0][1:])
		}

		words.WriteString(results[i][0])
	}
	pascalProjectName := words.String()

	// env
	data := TemplateData{
		Name:               projectName,
		ServiceName:        fmt.Sprintf("%sService", pascalProjectName),
		ServiceType:        fmt.Sprintf("*%sService", pascalProjectName),
		ServiceConstructor: fmt.Sprintf("&%sService{}", pascalProjectName),
		ServiceConfigType:  fmt.Sprintf("*%sConfig", pascalProjectName),
		ServiceConfigName:  fmt.Sprintf("%sConfig", pascalProjectName),
	}

	fmt.Printf("Creating project %s...\n", projectName)
	err := os.Mkdir(projectName, 0755)
	if err != nil {
		log.Fatal(fmt.Errorf("error creating folder ./%s: %v", projectName, err))
	}

	writeOrFail(projectName, "main.go", mainFile, data)
	writeOrFail(projectName, "config.go", configFile, data)
	writeOrFail(projectName, "service.go", serviceFile, data)
	writeOrFail(projectName, "config.yaml", configYamlFile, data)

	fmt.Println("Done! Happy coding!")
}

func writeOrFail(projectName, fileName string, file []byte, data TemplateData) {
	err := os.WriteFile(fmt.Sprintf("%s/%s", projectName, fileName), renderTemplate(file, data), 0644)
	if err != nil {
		log.Fatal(fmt.Errorf("could not write %s: %w", fileName, err))
	}
}

func renderTemplate(tmp []byte, data TemplateData) []byte {
	t, err := template.New("tmp").Parse(string(tmp))
	if err != nil {
		panic(errors.Wrap(err, "error parsing template").Error())
	}

	out := bytes.NewBuffer([]byte{})
	if t.Execute(out, data) != nil {
		panic(errors.Wrap(err, "error executing template").Error())
	}

	return out.Bytes()
}
