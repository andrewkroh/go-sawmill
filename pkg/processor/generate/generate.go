// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

import (
	"bytes"
	_ "embed"
	"flag"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"text/template"

	"github.com/andrewkroh/go-event-pipeline/internal/proctemplate"
)

var (
	//go:embed assets/processor.go.gotmpl
	processorTemplate string

	//go:embed assets/license-header.txt
	licenseHeader string
)

type TemplateData struct {
	License            string
	Name               string
	IncludeProcessFunc bool
	proctemplate.Processor
}

var goFileTemplate = template.Must(template.New("processor").
	Funcs(proctemplate.TemplateFuncs).
	Parse(processorTemplate))

// Flags
var processorsYmlFile string

func init() {
	flag.StringVar(&processorsYmlFile, "p", "processors.yml", "processors.yml file to use as the source")
}

func main() {
	flag.Parse()

	p, err := proctemplate.ReadProcessorsYAMLFile(processorsYmlFile)
	if err != nil {
		log.Fatal(err)
	}

	// Output generated data to the directory holding the processors.yml file.
	outputDir, err := filepath.Abs(filepath.Dir(processorsYmlFile))
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range p.Processors {
		for name, data := range p {
			// Sort config options by name.
			sort.Slice(data.Configuration, func(i, j int) bool {
				return data.Configuration[i].Name < data.Configuration[j].Name
			})

			templateData := TemplateData{
				License:            licenseHeader,
				Name:               name,
				Processor:          data,
				IncludeProcessFunc: true,
			}

			outputGoFile := filepath.Join(outputDir, name, name+".go")

			// Check if the existing file has been modified and keep that
			// modification.
			hasProcess, err := hasProcessFunc(outputGoFile)
			if err == nil && !hasProcess {
				templateData.IncludeProcessFunc = false
			}

			// Render template.
			buf := new(bytes.Buffer)
			if err := goFileTemplate.Execute(buf, templateData); err != nil {
				log.Fatal(err)
			}

			// gofmt the output.
			srcBytes, err := format.Source(buf.Bytes())
			if err != nil {
				log.Fatal(err)
			}

			if err = os.MkdirAll(filepath.Dir(outputGoFile), 0o755); err != nil {
				log.Fatal(err)
			}

			if err = ioutil.WriteFile(outputGoFile, srcBytes, 0o644); err != nil {
				log.Fatal(err)
			}
		}
	}
}

// hasProcessFunc returns true if the file contains a Process function.
func hasProcessFunc(goFile string) (bool, error) {
	fileSet := token.NewFileSet()

	f, err := parser.ParseFile(fileSet, goFile, nil, parser.ParseComments)
	if err != nil {
		return false, err
	}

	var found bool
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Name.Name == "Process" {
				found = true
				return false
			}
		}
		return true
	})

	return found, nil
}
