// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

// Binary mk-unmarshal-didllite builds the unmarshalDIDLLite() function using compile-time reflection.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/ethulhu/helix/upnpav"
)

var (
	out = flag.String("out", "", "output path (unset for stdout)")
)

func main() {
	flag.Parse()

	didllite := reflect.TypeOf(upnpav.DIDLLite{})

	var buf bytes.Buffer
	fmt.Fprintln(&buf, header)
	mkStatements(&buf, "root", "didllite", didllite)
	fmt.Fprintf(&buf, footer)

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		log.Print(buf.String())
		panic(fmt.Sprintf("could not format source: %v", err))
	}

	if *out != "" {
		ioutil.WriteFile(*out, formatted, os.ModePerm)
	} else {
		os.Stdout.Write(formatted)
	}
}

func mkStatements(w io.Writer, element, variable string, t reflect.Type) {
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("unsupported kind: %v", t.Kind()))
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		variableFieldName := fmt.Sprintf("%s.%s", variable, field.Name)

		tag, ok := field.Tag.Lookup("upnpav")
		if !ok && field.Type.Kind() == reflect.Struct {
			mkStatements(w, element, variableFieldName, field.Type)
			continue
		}
		if !ok {
			continue
		}

		if tag == ",innerxml" {
			fmt.Fprintf(w, "\n%s = %s.Text()\n", variableFieldName, element)
			continue
		}

		tagParts := strings.Split(tag, ",")
		name := tagParts[0]

		switch field.Type.Kind() {
		case reflect.Slice:
			switch field.Type.Elem().Kind() {
			case reflect.String:
				fmt.Fprintf(w, `
					for _, el := range %s.SelectElements("%s") {
						%s = append(%s, el.Text())
					}
				`, element, name, variableFieldName, variableFieldName)
			default:
				fmt.Fprintf(w, `
					for _, el := range %s.SelectElements("%s") {
						%s := %s{}
				`, element, name, strings.ToLower(field.Name), typeName(field.Type.Elem()))
				mkStatements(w, "el", strings.ToLower(field.Name), field.Type.Elem())
				fmt.Fprintf(w, `
						%s = append(%s, %s)
					}
				`, variableFieldName, variableFieldName, strings.ToLower(field.Name))
			}

		case reflect.Struct:
			switch field.Type {
			case reflect.TypeOf(time.Time{}):
				fmt.Fprintf(w, `
					if el := %s.SelectElement("%s"); el != nil {
						t, err := time.Parse("2006-01-02", el.Text())
						if err != nil {
							return nil, fmt.Errorf("could not parse date %%q: %%w", el.Text(), err)
						}
						%s = t
					}
				`, element, name, variableFieldName)
			default:
				panic(fmt.Sprintf("unsupported struct type %v for field %v", field.Type, field.Name))
			}

		case reflect.Bool:
			if len(tagParts) >= 3 && tagParts[2] == "inverse" {
				fmt.Fprintf(w, `
					%s = true
					if value := %s.SelectAttrValue("%s", ""); value != "" {
						b, err := unmarshalBool(value)
						if err != nil {
							return nil, fmt.Errorf("could not parse bool %%q: %%w", value, err)
						}
						%s = !b
					}
				`, variableFieldName, element, name, variableFieldName)
			} else {
				fmt.Fprintf(w, `
					if value := %s.SelectAttrValue("%s", ""); value != "" {
						b, err := unmarshalBool(value)
						if err != nil {
							return nil, fmt.Errorf("could not parse bool %%q: %%w", value, err)
						}
						%s = b
					}
				`, element, name, variableFieldName)
			}

		case reflect.Int:
			if len(tagParts) == 2 && tagParts[1] == "attr" {
				fmt.Fprintf(w, `
					if value := %s.SelectAttrValue("%s", ""); value != "" {
						parsed, err := strconv.Atoi(value)
						if err != nil {
							return nil, fmt.Errorf("could not parse int %%q: %%w", value, err)
						}
						%s = parsed
					}
				`, element, name, variableFieldName)
			} else {
				fmt.Fprintf(w, `
					if el := %s.SelectElement("%s"); el != nil {
						parsed, err := strconv.Atoi(el.Text())
						if err != nil {
							return nil, fmt.Errorf("could not parse int %%q: %%w", el.Text(), err)
						}
						%s = parsed
					}
				`, element, name, variableFieldName)
			}

		case reflect.Uint:
			if len(tagParts) == 2 && tagParts[1] == "attr" {
				fmt.Fprintf(w, `
					if value := %s.SelectAttrValue("%s", ""); value != "" {
						parsed, err := strconv.ParseUint(value, 10, 64)
						if err != nil {
							return nil, fmt.Errorf("could not parse int %%q: %%w", value, err)
						}
						%s = uint(parsed)
					}
				`, element, name, variableFieldName)
			} else {
				fmt.Fprintf(w, `
					if el := %s.SelectElement("%s"); el != nil {
						parsed, err := strconv.ParseUint(el.Text(), 10, 64)
						if err != nil {
							return nil, fmt.Errorf("could not parse int %%q: %%w", el.Text(), err)
						}
						%s = uint(parsed)
					}
				`, element, name, variableFieldName)
			}

		case reflect.Int64:
			if field.Type != reflect.TypeOf(time.Duration(0)) {
				panic(fmt.Sprintf("got non-time.Duration int64 for field %q", field.Name))
			}
			if len(tagParts) == 2 && tagParts[1] == "attr" {
				fmt.Fprintf(w, `
					if value := %s.SelectAttrValue("%s", ""); value != "" {
						parsed, err := ParseDuration(value)
						if err != nil {
							return nil, fmt.Errorf("could not parse int %%q: %%w", value, err)
						}
						%s = parsed
					}
				`, element, name, variableFieldName)
			} else {
				fmt.Fprintf(w, `
					if el := %s.SelectElement("%s"); el != nil {
						parsed, err := ParseDuration(el.Text())
						if err != nil {
							return nil, fmt.Errorf("could not parse %s %%q: %%w", el.Text(), err)
						}
						%s = parsed
					}
				`, element, name, name, variableFieldName)
			}

		case reflect.String:
			constructor := typeName(field.Type)
			if len(tagParts) == 2 && tagParts[1] == "attr" {
				fmt.Fprintf(w, `
					if value := %s.SelectAttrValue("%s", ""); value != "" {
						%s = %s(value)
					}
				`, element, name, variableFieldName, constructor)
			} else {
				fmt.Fprintf(w, `
					if el := %s.SelectElement("%s"); el != nil {
						%s = %s(el.Text())
					}
				`, element, name, variableFieldName, constructor)
			}

		case reflect.Ptr:
			switch field.Type {
			case reflect.TypeOf(&upnpav.ProtocolInfo{}):
				fmt.Fprintf(w, `
					if value := %s.SelectAttrValue("%s", ""); value != "" {
						parsed, err := ParseProtocolInfo(value)
						if err != nil {
							return nil, fmt.Errorf("could not parse protocolInfo %%q: %%w", value, err)
						}
						%s = parsed
					}
				`, element, name, variableFieldName)
			case reflect.TypeOf(&upnpav.Resolution{}):
				fmt.Fprintf(w, `
					if value := %s.SelectAttrValue("%s", ""); value != "" {
						parsed, err := ParseResolution(value)
						if err != nil {
							return nil, fmt.Errorf("could not parse resolution %%q: %%w", value, err)
						}
						%s = parsed
					}
				`, element, name, variableFieldName)
			case reflect.TypeOf(&url.URL{}):
				fmt.Fprintf(w, `
					if value := %s.SelectAttrValue("%s", ""); value != "" {
						parsed, err := url.Parse(value)
						if err != nil {
							return nil, fmt.Errorf("could not parse %s %%q: %%w", value, err)
						}
						%s = parsed
					}
				`, element, name, name, variableFieldName)
			default:
				panic(fmt.Sprintf("unxepected pointer type %v at field %v", field.Type, field.Name))
			}

		}
	}
}

var (
	header = `
		package upnpav

		import (
			"errors"
			"fmt"
			"net/url"
			"strconv"
			"time"

			"github.com/beevik/etree"
		)

		func unmarshalDIDLLite(document *etree.Document) (*DIDLLite, error) {
			root := document.SelectElement("DIDL-Lite")
			if root == nil {
				return nil, errors.New("missing <DIDL-Lite> root element")
			}

			didllite := &DIDLLite{}
	`

	footer = `

			return didllite, nil
		}

		func unmarshalBool(raw string) (bool, error) {
			if raw == "0" || raw == "false" || raw == "no" {
				return false, nil
			}
			if raw == "1" || raw == "true" || raw == "yes" {
				return true, nil
			}
			return false, fmt.Errorf("must be one of {1, 0, true, false, yes, no}")
		}
		`

	simpleElement = template.Must(template.New("simpleElement").Parse(`
if el := {{ .Element }}.SelectElement("{{ .Name }}"); el != nil {
	raw := 
	{{ .VariableFieldName }} = 
}
	`))
)

func typeName(t reflect.Type) string {
	return strings.Replace(t.String(), "upnpav.", "", -1)
}
