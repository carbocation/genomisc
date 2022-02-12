package main

import (
	"text/template"

	_ "embed"
)

//go:embed query_template.sql
var queryTemplateString string

// mkMap allows you to create a map within a template, so that you can pass more
// than one parameter to a template block. Inspired by
// https://stackoverflow.com/a/25013152/199475 .
func mkMap(args ...interface{}) map[interface{}]interface{} {
	out := make(map[interface{}]interface{})
	for k, v := range args {
		if k%2 == 0 {
			continue
		}
		out[args[k-1]] = v
	}
	return out
}

// TODO: create a `has_died` field and apply the censor table's
// death_censor_date properly. Rename death_date to death_censor_date. Rename
// death_age to death_censor_age.
//
// TODO: Resolve age_censor vs enroll_age. Choose one or the other (likely the
// latter, so you end up with enroll_age, censor_age, death_censor_age).
var queryTemplate = template.Must(template.New("").Funcs(template.FuncMap(map[string]interface{}{"mkMap": mkMap})).Parse(queryTemplateString))
