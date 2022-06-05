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

package proctemplate

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"github.com/mitchellh/go-wordwrap"
)

var TemplateFuncs = template.FuncMap{
	"to_lower":            strings.ToLower,
	"to_exported_go_type": goTypeName,
	"to_title":            strings.Title,
	"description":         descriptionToComment,
	"select_defaults":     selectDefaults,
	"quote_strings":       quoteStrings,
	"config_type_imports": configTypeImports,
	"trim_import":         trimImportPrefix,
	"bool_to_x":           boolToX,
	"replace":             strings.ReplaceAll,
}

// descriptionToComment builds a comment string that is wrapped at 80 chars.
func descriptionToComment(indent, desc string) (string, error) {
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
	return trimTrailingWhitespace(strings.Join(lines, "\n"+indent+"// "))
}

func trimTrailingWhitespace(text string) (string, error) {
	var lines [][]byte
	s := bufio.NewScanner(bytes.NewBufferString(text))
	for s.Scan() {
		lines = append(lines, bytes.TrimRightFunc(s.Bytes(), unicode.IsSpace))
	}
	if err := s.Err(); err != nil {
		return "", err
	}
	return string(bytes.Join(lines, []byte("\n"))), nil
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

// abbreviations capitalizes common abbreviations.
func abbreviations(abv string) string {
	switch strings.ToLower(abv) {
	case "id", "ppid", "pid", "pgid", "mac", "ip", "iana", "uid", "ecs", "as", "icmp":
		return strings.ToUpper(abv)
	default:
		return abv
	}
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

func selectDefaults(in []ConfigurationOption) []ConfigurationOption {
	var defaults []ConfigurationOption
	for _, d := range in {
		if d.Default != nil {
			defaults = append(defaults, d)
		}
	}
	return defaults
}

func quoteStrings(in interface{}) string {
	switch v := in.(type) {
	case string:
		return strconv.Quote(v)
	default:
		return fmt.Sprintf("%v", in)
	}
}

func configTypeImports(opts []ConfigurationOption) []string {
	imports := map[string]struct{}{}

	for _, conf := range opts {
		idx := strings.LastIndex(conf.Type, ".")
		if idx == -1 {
			continue
		}
		imports[conf.Type[:idx]] = struct{}{}
	}

	list := make([]string, 0, len(imports))
	for k := range imports {
		list = append(list, k)
	}

	return list
}

func trimImportPrefix(dataType string) string {
	idx := strings.LastIndex(dataType, "/")
	if idx == -1 {
		return dataType
	}
	return dataType[idx+1:]
}

func boolToX(b bool) string {
	if b {
		return "x"
	}
	return ""
}
