package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"text/template"
)

const (
	registryPlaceholder = `// new presets go here
`
	pytestPlaceholder = `### gnomock-generator
`
	startPresetPlaceholder = `### /start/preset
`
	requestBodyPlaceholder = `### preset-request
`
	readmePlaceholder = `<!-- new presets go here -->
`
	ciTestPlaceholder = `### preset tests go here
`
	circleciJobsPlaceholder = `### circleci jobs go here
`
)

var fMap = template.FuncMap{
	"lower": strings.ToLower,
}

type presetParams struct {
	Name        string
	DefaultPort int
	Image       string
	Public      bool
}

func main() {
	if err := generate(); err != nil {
		log.Fatalln(err)
	}

	log.Println("done")
}

func generate() error {
	var pp presetParams

	flag.StringVar(&pp.Name, "name", "", `new preset name, e.g "Redis", "Postgres", etc.`)
	flag.StringVar(&pp.Image, "image", "", "full docker image name")
	flag.IntVar(&pp.DefaultPort, "default-port", 0, "default TCP Port to use")
	flag.BoolVar(&pp.Public, "public", false, "prepare this preset for public use")
	flag.Parse()

	if err := presetPkg(pp); err != nil {
		return err
	}

	if pp.Public {
		if err := gnomockdPkg(pp); err != nil {
			return err
		}

		if err := registry(pp); err != nil {
			return err
		}

		if err := sdktestPkg(pp); err != nil {
			return err
		}

		if err := swagger(pp); err != nil {
			return err
		}

		if err := readme(pp); err != nil {
			return err
		}

		if err := github(pp); err != nil {
			return err
		}

		if err := circleci(pp); err != nil {
			return err
		}
	}

	return nil
}

// presetPkg generates a minimal working version of a preset. It also creates a
// README.md file that needs to be manually edited when the preset is ready.
func presetPkg(params presetParams) error {
	dir := path.Join("preset", strings.ToLower(params.Name))
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("can't create preset folder: %w", err)
	}

	files := []string{"preset.go", "preset_test.go", "options.go", "README.md"}
	for _, file := range files {
		if err := presetFile(dir, file, params); err != nil {
			return fmt.Errorf("can't generate preset file: %w", err)
		}
	}

	return nil
}

func presetFile(dir, file string, params presetParams) error {
	fName := fmt.Sprintf("%s.template", file)

	tmpl, err := template.New(fName).Funcs(fMap).ParseFiles(path.Join("cmd/generator/templates/preset", fName))
	if err != nil {
		return fmt.Errorf("can't parse template: %w", err)
	}

	presetFile := path.Join(dir, file)

	f, err := os.Create(presetFile)
	if err != nil {
		return fmt.Errorf("can't create new file: %w", err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Println("can't close file:", err)
		}
	}()

	if err := tmpl.Execute(f, params); err != nil {
		return fmt.Errorf("can't execute template: %w", err)
	}

	return nil
}

// gnomockdPkg adds the preset tests to gnomockd package.
func gnomockdPkg(params presetParams) error {
	gnomockdPath := path.Join("internal", "gnomockd")
	testdataPath := path.Join(gnomockdPath, "testdata")

	dir := path.Join(testdataPath, strings.ToLower(params.Name))
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("can't create testdata dir: %s", err)
	}

	presetFileName := fmt.Sprintf("%s.json", strings.ToLower(params.Name))

	presetFile, err := os.Create(path.Join(testdataPath, presetFileName))
	if err != nil {
		return fmt.Errorf("can't create preset file: %w", err)
	}

	defer func() {
		if err := presetFile.Close(); err != nil {
			log.Println("can't close file:", err)
		}
	}()

	if err := json.NewEncoder(presetFile).Encode(map[string]interface{}{
		"preset": map[string]interface{}{
			"version": "latest",
		},
		"options": map[string]interface{}{
			"debug": true,
		},
	}); err != nil {
		return fmt.Errorf("can't write into preset file: %w", err)
	}

	presetTestTemplate := "cmd/generator/templates/gnomockd/preset_test.go.template"

	tmpl, err := template.New("preset_test.go.template").Funcs(fMap).ParseFiles(presetTestTemplate)
	if err != nil {
		return fmt.Errorf("can't parse template: %w", err)
	}

	testFile := path.Join(gnomockdPath, fmt.Sprintf("%s_test.go", strings.ToLower(params.Name)))

	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("can't create new file: %w", err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Println("can't close file:", err)
		}
	}()

	if err := tmpl.Execute(f, params); err != nil {
		return fmt.Errorf("can't execute template: %w", err)
	}

	return nil
}

// registry adds the new preset to gnomockd preset registry so that it becomes
// available over HTTP.
func registry(params presetParams) error {
	if err := replacePlaceholder(
		path.Join("cmd", "server", "presets.go"),
		"cmd/generator/templates/cmd/server/presets.go.template",
		registryPlaceholder,
		params,
	); err != nil {
		return fmt.Errorf("can't generate registry code: %w", err)
	}

	return nil
}

// sdktestPkg generates code required for testing the generated SDK. It creates
// a `testdata` folder for a new preset and adds a test stub to `test_sdk.py`.
//
// Test generator is the most basic possible: it will generate wrong names for
// any preset that doesn't have a single word, simple case name like "Redis" or
// "Kubernetes": names like RabbitMQ will break.
func sdktestPkg(params presetParams) error {
	testPath := path.Join("sdktest", "python", "test")
	dir := path.Join(testPath, "testdata", strings.ToLower(params.Name))

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("can't create testdata dir: %w", err)
	}

	pytestFileName := path.Join(testPath, "test_sdk.py")

	if err := replacePlaceholder(
		pytestFileName,
		"cmd/generator/templates/sdktest/python/test/test_sdk.py.template",
		pytestPlaceholder,
		params,
	); err != nil {
		return fmt.Errorf("can't generate python tests: %w", err)
	}

	return nil
}

// swagger generates new definitions in swagger.yaml file. These definitions
// should be extended with options supported by a new preset.
func swagger(params presetParams) error {
	swaggerFile := path.Join("swagger", "swagger.yaml")

	if err := replacePlaceholder(
		swaggerFile,
		"cmd/generator/swagger/start.template",
		startPresetPlaceholder,
		params,
	); err != nil {
		return fmt.Errorf("can't generate swagger spec: %w", err)
	}

	if err := replacePlaceholder(
		swaggerFile,
		"cmd/generator/swagger/body.template",
		requestBodyPlaceholder,
		params,
	); err != nil {
		return fmt.Errorf("can't generate swagger spec: %w", err)
	}

	return nil
}

func readme(params presetParams) error {
	return replacePlaceholder(
		"README.md",
		"cmd/generator/templates/README.md.template",
		readmePlaceholder,
		params,
	)
}

func github(params presetParams) error {
	return replacePlaceholder(
		path.Join(".github", "workflows", "test.yaml"),
		"cmd/generator/templates/.github/workflows/test.yaml.template",
		ciTestPlaceholder,
		params,
	)
}

func circleci(params presetParams) error {
	circleciPath := path.Join(".circleci", "config.yml")

	if err := replacePlaceholder(
		circleciPath,
		"cmd/generator/templates/.circleci/config.yml.template",
		ciTestPlaceholder,
		params,
	); err != nil {
		return fmt.Errorf("can't create circleci job: %w", err)
	}

	if err := replacePlaceholder(
		circleciPath,
		"cmd/generator/templates/.circleci/jobs.template",
		circleciJobsPlaceholder,
		params,
	); err != nil {
		return fmt.Errorf("can't add circleci job to config.yml: %w", err)
	}

	return nil
}

// replacePlaceholder replaces `placeholder` in `targetFile` with the result of
// `tmplFile` template execution using `params` values.
//
// nolint:gosec
func replacePlaceholder(targetFile, tmplFile, placeholder string, params presetParams) error {
	targetBs, err := ioutil.ReadFile(targetFile)
	if err != nil {
		return fmt.Errorf("can't read %s: %w", targetFile, err)
	}

	tmplName := path.Base(tmplFile)

	tmpl, err := template.New(tmplName).Funcs(fMap).ParseFiles(tmplFile)
	if err != nil {
		return fmt.Errorf("can't read %s: %w", tmplName, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return fmt.Errorf("can't execute template: %w", err)
	}

	targetStr := strings.ReplaceAll(string(targetBs), placeholder, buf.String())

	if err := ioutil.WriteFile(targetFile, []byte(targetStr), 0644); err != nil {
		return fmt.Errorf("can't write %s: %w", targetFile, err)
	}

	return nil
}
