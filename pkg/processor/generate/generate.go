package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	wordwrap "github.com/mitchellh/go-wordwrap"
	"gopkg.in/yaml.v3"
)

func main() {
	f, err := os.Open("../processors.yml")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	dec := yaml.NewDecoder(bufio.NewReader(f))
	//dec.KnownFields(true)
	//var i map[interface{}]interface{}
	var p Processors
	if err := dec.Decode(&p); err != nil {
		log.Fatal(err)
	}

	//m := cleanMapInterface(i)
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(data))

	for _, p := range p.Processors {
		for name, data := range p {
			sort.Slice(data.Configuration, func(i, j int) bool {
				return data.Configuration[i].Name < data.Configuration[j].Name
			})

			buf := new(bytes.Buffer)
			err := goFileTemplate.Execute(buf, ProcessorTemplateVar{
				License:   "// Foo License",
				Name:      name,
				Processor: data,
			})
			if err != nil {
				log.Fatal(err)
			}

			path := filepath.Join("..", name, name+".go")
			if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				log.Fatal(err)
			}

			if err = ioutil.WriteFile(path, buf.Bytes(), 0644); err != nil {
				log.Fatal(err)
			}
		}
	}

}

type Processors struct {
	Processors []map[string]Processor
}

type Processor struct {
	Description   string
	Configuration []ConfigurationOption
}

type ConfigurationOption struct {
	Name        string
	Type        string
	Required    bool
	Optional    bool
	Default     interface{}
	Description string
}

func cleanSlice(in []interface{}) []interface{} {
	result := make([]interface{}, len(in))
	for i, v := range in {
		result[i] = cleanValue(v)
	}
	return result
}

func cleanMapInterface(in map[interface{}]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range in {
		key := k.(string)
		result[key] = cleanValue(v)
	}
	return result
}

func cleanMapString(in map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range in {
		result[k] = cleanValue(v)
	}
	return result
}

func cleanValue(v interface{}) interface{} {
	switch v := v.(type) {
	case []interface{}:
		return cleanSlice(v)
	case map[interface{}]interface{}:
		return cleanMapInterface(v)
	case map[string]interface{}:
		return cleanMapString(v)
	default:
		return v
	}
}

// descriptionToComment builds a comment string that is wrapped at 80 chars.
func descriptionToComment(indent, desc string) string {
	textLength := 80 - len(strings.Replace(indent, "\t", "    ", 4)+" // ")
	lines := strings.Split(wordwrap.WrapString(desc, uint(textLength)), "\n")
	if len(lines) > 0 {
		// Remove empty first line.
		if strings.TrimSpace(lines[0]) == "" {
			lines = lines[1:]
		}
	}
	if len(lines) > 0 {
		// Remove empty last line.
		if strings.TrimSpace(lines[len(lines)-1]) == "" {
			lines = lines[:len(lines)-1]
		}
	}
	for i := 0; i < len(lines); i++ {

	}
	return trimTrailingWhitespace(strings.Join(lines, "\n"+indent+"// "))
}

func trimTrailingWhitespace(text string) string {
	var lines [][]byte
	s := bufio.NewScanner(bytes.NewBufferString(text))
	for s.Scan() {
		lines = append(lines, bytes.TrimRightFunc(s.Bytes(), unicode.IsSpace))
	}
	if err := s.Err(); err != nil {
		log.Fatal(err)
	}
	return string(bytes.Join(lines, []byte("\n")))
}

// goDataType returns the Go type to use for Elasticsearch mapping data type.
func goDataType(fieldName, elasticsearchDataType string) string {
	// Special cases.
	switch {
	case fieldName == "duration" && elasticsearchDataType == "long":
		return "time.Duration"
	case fieldName == "args" && elasticsearchDataType == "keyword":
		return "[]string"
	}

	switch elasticsearchDataType {
	case "keyword", "text", "ip", "geo_point":
		return "string"
	case "long":
		return "int64"
	case "integer":
		return "int32"
	case "float":
		return "float64"
	case "date":
		return "time.Time"
	case "boolean":
		return "bool"
	case "object":
		return "map[string]interface{}"
	default:
		log.Fatalf("no translation for %v (field %s)", elasticsearchDataType, fieldName)
		return ""
	}
}

// abbreviations capitalizes common abbreviations.
func abbreviations(abv string) string {
	switch strings.ToLower(abv) {
	case "id", "ppid", "pid", "pgid", "mac", "ip", "iana", "uid", "ecs", "as", "icmp":
		return strings.ToUpper(abv)
	default:
		return abv
	}
}

// goTypeName removes special characters ('_', '.', '@') and returns a
// camel-cased name.
func goTypeName(name string) string {
	var b strings.Builder
	for _, w := range strings.FieldsFunc(name, isSeparator) {
		b.WriteString(strings.Title(abbreviations(w)))
	}
	return b.String()
}

// isSeparate returns true if the character is a field name separator. This is
// used to detect the separators in fields like ephemeral_id or instance.name.
func isSeparator(c rune) bool {
	switch c {
	case '.', '_':
		return true
	case '@':
		// This effectively filters @ from field names.
		return true
	default:
		return false
	}
}

var (
	goFileTemplate = template.Must(template.New("type").Funcs(templateFuncs).Parse(
		strings.Replace(typeTmpl[1:], `\u0060`, "`", -1)))
)

const typeTmpl = `
{{.License}}
// Code generated by processor/generate.go - DO NOT EDIT.
package {{ .Name | to_lower }}

import (
	"github.com/andrewkroh/go-event-pipeline/pkg/processor"
	"github.com/andrewkroh/go-event-pipeline/pkg/processor/registry"
)

func init() {
	registry.MustRegister(processorName, New)
}

const (
	processorName = "{{ .Name }}"
)

// Config contains the configuration options for the {{ .Name }} processor.
type Config struct {
{{- range $field := .Configuration}}
	// {{ description "\t" $field.Description}}
	{{$field.Name | to_exported_go_type}} {{$field.Type}} \u0060config:"{{$field.Name}}"{{ if $field.Required }} validate:"required"{{ end }}\u0060
{{ end -}}
}

// InitDefaults initializes the configuration options to their default values.
func (c *Config) InitDefaults() {
{{- range $field := .Configuration | select_defaults }}
	c.{{$field.Name | to_exported_go_type}} = {{$field.Default | quote_strings }}{{ end }}
}

// {{ description "" .Description }}
type {{.Name | to_exported_go_type }} struct {
	config Config
}

// New returns a new {{.Name | to_exported_go_type}} processor.
func New(config Config) (*{{.Name | to_exported_go_type}}, error) {
	return &{{.Name | to_exported_go_type}}{config: config}, nil
}

// Config returns the {{.Name | to_exported_go_type}} processor config.
func (p *{{.Name | to_exported_go_type}}) Config() Config {
	return p.config
}

func (p *{{.Name | to_exported_go_type}}) String() string {
	return processor.ConfigString(processorName, p.config)
}

func (p *{{.Name | to_exported_go_type}}) Process(event processor.Event) error {
	return nil
}
`

var templateFuncs = template.FuncMap{
	"to_lower":            strings.ToLower,
	"to_exported_go_type": goTypeName,
	"to_title":            strings.Title,
	"description":         descriptionToComment,
	"select_defaults": func(in []ConfigurationOption) []ConfigurationOption {
		var defaults []ConfigurationOption
		for _, d := range in {
			if d.Default != nil {
				defaults = append(defaults, d)
			}
		}
		return defaults
	},
	"quote_strings": func(in interface{}) string {
		switch v := in.(type) {
		case string:
			return strconv.Quote(v)
		default:
			return fmt.Sprintf("%v", in)
		}
	},
}

type ProcessorTemplateVar struct {
	License string
	Name    string
	Processor
}