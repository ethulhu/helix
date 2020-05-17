// Binary mk-marshal-didllite builds the marshalDIDLLite() function using compile-time reflection.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"text/template"
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

		tag, ok := field.Tag.Lookup("upnpav")
		if !ok && field.Type.Kind() == reflect.Struct {
			mkStatements(w, element, variable, field.Type)
			continue
		}
		if !ok {
			continue
		}

		if tag == ",innerxml" {
			fmt.Fprintf(w, "\n%s.CreateText(fmt.Sprintf(\"%%v\", %s.%s))\n", element, variable, field.Name)
			continue
		}

		tagParts := strings.Split(tag, ",")
		name := tagParts[0]

		simpleIfTmpl := simpleIfElement
		if len(tagParts) >= 2 && tagParts[1] == "attr" {
			simpleIfTmpl = simpleIfAttr
		}

		switch field.Type.Kind() {
		case reflect.Slice:
			switch field.Type.Elem().Kind() {
			case reflect.String:
				params := simpleParams{
					Variable:      variable,
					Field:         field.Name,
					ZeroValue:     `""`,
					ParentElement: element,
					Name:          name,
				}
				if err := simpleForElement.Execute(w, params); err != nil {
					panic(fmt.Sprintf("could not write template: %v", err))
				}
			default:
				fmt.Fprintf(w, `
					for _, %s := range %s.%s {
						el := %s.CreateElement("%s")
				`, strings.ToLower(field.Name), variable, field.Name, element, name)
				mkStatements(w, "el", strings.ToLower(field.Name), field.Type.Elem())
				fmt.Fprint(w, `
					}
				`)
			}

		case reflect.Struct:
			switch field.Type {
			case reflect.TypeOf(time.Time{}):
				params := simpleParams{
					Variable:      variable,
					Field:         field.Name,
					ParentElement: element,
					Name:          name,
				}
				if err := timeIfElement.Execute(w, params); err != nil {
					panic(fmt.Sprintf("could not write template: %v", err))
				}
			default:
				panic(fmt.Sprintf("unsupported struct type %v for field %v", field.Type, field.Name))
			}

		case reflect.Bool:
			if len(tagParts) >= 3 && tagParts[2] == "inverse" {
				fmt.Fprintf(w, `
						%s.CreateAttr("%s", marshalBool(!%s.%s))
				`, element, name, variable, field.Name)
			} else {
				fmt.Fprintf(w, `
						%s.CreateAttr("%s", marshalBool(%s.%s))
				`, element, name, variable, field.Name)
			}

		case reflect.Int64:
			if field.Type != reflect.TypeOf(time.Duration(0)) {
				panic(fmt.Sprintf("got non-time.Duration int64 for field %q", field.Name))
			}
			fallthrough
		case reflect.Int:
			fallthrough
		case reflect.Uint:
			params := simpleParams{
				Variable:      variable,
				Field:         field.Name,
				ZeroValue:     "0",
				ParentElement: element,
				Name:          name,
			}
			if err := simpleIfTmpl.Execute(w, params); err != nil {
				panic(fmt.Sprintf("could not write template: %v", err))
			}
		case reflect.String:
			params := simpleParams{
				Variable:      variable,
				Field:         field.Name,
				ZeroValue:     `""`,
				ParentElement: element,
				Name:          name,
			}
			if err := simpleIfTmpl.Execute(w, params); err != nil {
				panic(fmt.Sprintf("could not write template: %v", err))
			}
		case reflect.Ptr:
			params := simpleParams{
				Variable:      variable,
				Field:         field.Name,
				ZeroValue:     "nil",
				ParentElement: element,
				Name:          name,
			}
			if err := simpleIfTmpl.Execute(w, params); err != nil {
				panic(fmt.Sprintf("could not write template: %v", err))
			}

		default:
			panic(fmt.Sprintf("unsupported kind %v for field %v", field.Type.Kind(), field.Name))
		}
	}
}

type (
	simpleParams struct {
		Variable, Field     string
		ZeroValue           interface{}
		ParentElement, Name string
	}
)

var (
	header = `
		package upnpav

		import (
			"fmt"
			"time"

			"github.com/beevik/etree"
		)

		func marshalDIDLLite(didllite *DIDLLite) *etree.Document {
			document := etree.NewDocument()

			document.CreateProcInst("xml", "version=\"1.0\" encoding=\"utf-8\"")

			root := document.CreateElement("DIDL-Lite")
			root.CreateAttr("xmlns", "urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/")
			root.CreateAttr("xmlns:dc", "http://purl.org/dc/elements/1.1/")
			root.CreateAttr("xmlns:upnp", "urn:schemas-upnp-org:metadata-1-0/upnp/")
			root.CreateAttr("xmlns:dlna", "urn:schemas-dlna-org:metadata-1-0/")`

	footer = `
			document.AddChild(root)

			return document
		}

		func marshalBool(b bool) string {
			if b {
				return "1"
			}
			return "0"
		}`

	simpleForElement = template.Must(template.New("simpleForElement").
				Funcs(map[string]interface{}{"toLower": strings.ToLower}).
				Parse(`
for _, {{ .Field | toLower }} := range {{ .Variable }}.{{ .Field }} {
	{{ .ParentElement }}.CreateElement("{{ .Name }}").CreateText(fmt.Sprintf("%v", {{ .Field | toLower }}))
}
`))

	simpleIfElement = template.Must(template.New("simpleIfElement").Parse(`
if {{ .Variable }}.{{ .Field }} != {{ .ZeroValue }} {
	{{ .ParentElement }}.CreateElement("{{ .Name }}").CreateText(fmt.Sprintf("%v", {{ .Variable }}.{{ .Field }}))
}
`))

	simpleIfAttr = template.Must(template.New("simpleIfAttr").Parse(`
if {{ .Variable }}.{{ .Field }} != {{ .ZeroValue }} {
	{{ .ParentElement }}.CreateAttr("{{ .Name }}", fmt.Sprintf("%v", {{ .Variable }}.{{ .Field }}))
}

`))

	timeIfElement = template.Must(template.New("timeIfElement").Parse(`
if {{ .Variable }}.{{ .Field }} != (time.Time{}) {
	{{ .ParentElement }}.CreateElement("{{ .Name }}").CreateText({{ .Variable }}.{{ .Field }}.Format("2006-01-02"))
}
`))
)
