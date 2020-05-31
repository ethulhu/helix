package scpd

import (
	"fmt"
	"reflect"
	"sort"
)

func Merge(docs ...Document) (Document, error) {
	var actions []Action
	var variables []StateVariable
	for _, doc := range docs {
		actions = append(actions, doc.Actions...)
		variables = append(variables, doc.StateVariables...)
	}

	mergedActions, err := mergeActions(actions)
	if err != nil {
		return Document{}, err
	}

	mergedVariables, err := mergeVariables(variables)
	if err != nil {
		return Document{}, err
	}

	return Document{
		SpecVersion:    Version,
		Actions:        mergedActions,
		StateVariables: mergedVariables,
	}, nil
}

func mergeActions(as []Action) ([]Action, error) {
	m := map[string]Action{}
	for _, a := range as {
		b, ok := m[a.Name]
		if !ok {
			m[a.Name] = a
			continue
		}
		if !reflect.DeepEqual(a, b) {
			return nil, fmt.Errorf("conflicting definitions for action %q", a.Name)
		}
	}

	var names []string
	for n := range m {
		names = append(names, n)
	}
	sort.Strings(names)

	var sortedActions []Action
	for _, n := range names {
		sortedActions = append(sortedActions, m[n])
	}
	return sortedActions, nil
}

func mergeVariables(vs []StateVariable) ([]StateVariable, error) {
	m := map[string]StateVariable{}
	for _, v := range vs {
		w, ok := m[v.Name]
		if !ok {
			m[v.Name] = v
			continue
		}
		if !reflect.DeepEqual(v, w) {
			return nil, fmt.Errorf("conflicting definitions for state variable %q", v.Name)
		}
	}

	var names []string
	for n := range m {
		names = append(names, n)
	}
	sort.Strings(names)

	var sortedVars []StateVariable
	for _, n := range names {
		sortedVars = append(sortedVars, m[n])
	}
	return sortedVars, nil
}
