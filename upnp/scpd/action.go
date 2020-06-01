// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package scpd

import (
	"fmt"
	"reflect"
	"strconv"
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
		if (parts[1] == "i4" || parts[1] == "ui4") && len(parts) > 2 {
			sv.AllowedValueRange = &AllowedValueRange{}
			for _, part := range parts[2:] {
				switch {
				case strings.HasPrefix(part, "min="):
					v, err := strconv.Atoi(strings.TrimPrefix(part, "min="))
					if err != nil {
						return nil, nil, fmt.Errorf("field %s has an invalid constraint: %s", field.Name, part)
					}
					sv.AllowedValueRange.Minimum = v
				case strings.HasPrefix(part, "max="):
					v, err := strconv.Atoi(strings.TrimPrefix(part, "max="))
					if err != nil {
						return nil, nil, fmt.Errorf("field %s has an invalid constraint: %s", field.Name, part)
					}
					sv.AllowedValueRange.Maximum = v
				case strings.HasPrefix(part, "step="):
					v, err := strconv.Atoi(strings.TrimPrefix(part, "step="))
					if err != nil {
						return nil, nil, fmt.Errorf("field %s has an invalid constraint: %s", field.Name, part)
					}
					sv.AllowedValueRange.Step = v
				default:
					return nil, nil, fmt.Errorf("field %s has an unexpected constraint: %s", field.Name, part)
				}
			}
		}

		arguments = append(arguments, arg)
		variables = append(variables, sv)
	}
	return arguments, variables, nil
}
