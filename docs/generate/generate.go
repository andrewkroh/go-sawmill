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
	"embed"
	_ "embed"
	"flag"
	"io/ioutil"
	"log"
	"sort"
	"text/template"

	"github.com/andrewkroh/go-event-pipeline/internal/proctemplate"
)

//go:embed assets/*.gotmpl
var templatesFS embed.FS

type TemplateData struct {
	Processors []Processor
}

type Processor struct {
	Name string
	proctemplate.Processor
}

var templates = template.Must(template.New("").
	Funcs(proctemplate.TemplateFuncs).
	Option("missingkey=error").
	ParseFS(templatesFS, "assets/*.gotmpl"))

// Flags
var processorsYmlFile string
var outputFile string

func init() {
	flag.StringVar(&processorsYmlFile, "p", "processors.yml", "processors.yml file to use as the source")
	flag.StringVar(&outputFile, "o", "README.md", "output file")
}

func main() {
	flag.Parse()

	p, err := proctemplate.ReadProcessorsYAMLFile(processorsYmlFile)
	if err != nil {
		log.Fatal(err)
	}

	templateData := TemplateData{}
	for _, v := range p.Processors {
		for name, proc := range v {
			sort.Slice(proc.Configuration, func(i, j int) bool {
				return proc.Configuration[i].Name < proc.Configuration[j].Name
			})
			for i := range proc.Configuration {
				proc.Configuration[i].Type = goTypeToYAMLType(proc.Configuration[i].Type)
			}
			templateData.Processors = append(templateData.Processors, Processor{
				Name:      name,
				Processor: proc,
			})
			break
		}
	}
	sort.Slice(templateData.Processors, func(i, j int) bool {
		return templateData.Processors[i].Name < templateData.Processors[j].Name
	})

	// Render template.
	buf := new(bytes.Buffer)
	if templates.ExecuteTemplate(buf, "README.md.gotmpl", templateData); err != nil {
		log.Fatal(err)
	}

	if err = ioutil.WriteFile(outputFile, buf.Bytes(), 0o644); err != nil {
		log.Fatal(err)
	}
}

func goTypeToYAMLType(in string) string {
	switch in {
	case "github.com/andrewkroh/go-event-pipeline/pkg/config.EventValue":
		return "any"
	default:
		return in
	}
}
