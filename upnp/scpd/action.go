package scpd

import (
	"fmt"
	"reflect"
	"strings"
)

func FromAction(name string, req, rsp interface{}) (Document, error) {
	inArgs, inVars, err := argumentsAndVariables(req, In)
	if err != nil {
		return Document{}, err
	}
	outArgs, outVars, err := argumentsAndVariables(rsp, Out)
	if err != nil {
		return Document{}, err
	}
	allVars, err := mergeVariables(append(inVars, outVars...))
	if err != nil {
		return Document{}, err
	}

	return Document{
		SpecVersion: Version,
		Actions: []Action{{
			Name:      name,
			Arguments: append(inArgs, outArgs...),
		}},
		StateVariables: allVars,
	}, nil
}

func argumentsAndVariables(obj interface{}, d Direction) ([]Argument, []StateVariable, error) {
	var arguments []Argument
	var variables []StateVariable
	t := reflect.TypeOf(obj)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		scpdTag, ok := field.Tag.Lookup("scpd")
		if !ok {
			continue
		}
		xmlTag, ok := field.Tag.Lookup("xml")
		if !ok {
			return nil, nil, fmt.Errorf("field %s must have an XML tag", field.Name)
		}

		parts := strings.Split(scpdTag, ",")
		if len(parts) < 2 {
			return nil, nil, fmt.Errorf("field %s SCPD tag must have at least 2 parts", field.Name)
		}

		arg := Argument{
			Name:                 xmlTag,
			Direction:            d,
			RelatedStateVariable: parts[0],
		}
		sv := StateVariable{
			Name:     parts[0],
			DataType: parts[1],
		}

		if parts[1] == "string" && len(parts) == 3 {
			sv.AllowedValues = &AllowedValues{}
			for _, allowed := range strings.Split(parts[2], "|") {
				sv.AllowedValues.Values = append(sv.AllowedValues.Values, allowed)
			}
		}

		arguments = append(arguments, arg)
		variables = append(variables, sv)
	}
	return arguments, variables, nil
}
