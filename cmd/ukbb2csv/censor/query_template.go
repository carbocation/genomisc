package main

import (
	"embed"
	"fmt"
	"strings"
	"text/template"

	_ "embed"

	"cloud.google.com/go/bigquery"
)

//go:embed *.sql
var embeddedQueryTemplates embed.FS

// //go:embed query_template.sql
// var queryTemplateString string

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

func BuildQuery(BQ *WrappedBigQuery, displayQuery bool, biobankSource string) (*bigquery.Query, error) {
	params := []bigquery.QueryParameter{}

	// Our query is dynamic, not static, so we assemble it from a text template,
	// which we will populate with a map:
	queryParts := map[string]interface{}{
		// True variables
		"database": BQ.Database,
	}

	// Fetch desired template from embedded resources
	templateBytes, err := embeddedQueryTemplates.ReadFile(fmt.Sprintf("query_template_%s.sql", biobankSource))
	if err != nil {
		return nil, err
	}

	// Parse the selected template
	queryTemplate, err := template.New("").
		Funcs(template.FuncMap(map[string]interface{}{"mkMap": mkMap})).
		Parse(string(templateBytes))
	if err != nil {
		return nil, err
	}

	// Assemble all the parts (execute the template)
	populatedQuery := &strings.Builder{}
	if err := queryTemplate.Execute(populatedQuery, queryParts); err != nil {
		return nil, err
	}

	if displayQuery {

		fmt.Println(populatedQuery.String())
		fmt.Println("Query parameters:")

		for _, v := range params {
			if x, ok := v.Value.([]string); ok {
				fmt.Printf("%v: (\"%s\")\n", v.Name, strings.Join(x, `","`))
				continue
			}
			fmt.Printf("%v: %v\n", v.Name, v.Value)
		}
		return nil, nil
	}

	// Generate the bigquery query object, but don't call it
	bqQuery := BQ.Client.Query(populatedQuery.String())
	bqQuery.QueryConfig.Parameters = append(bqQuery.QueryConfig.Parameters, params...)

	return bqQuery, nil
}
