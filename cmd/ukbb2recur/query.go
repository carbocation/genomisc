package main

import (
	"fmt"
	"strings"

	"cloud.google.com/go/bigquery"
)

func BuildQuery(BQ *WrappedBigQuery, tabs *TabFile, displayQuery bool) (*bigquery.Query, error) {
	params := []bigquery.QueryParameter{}

	// By default, if there is no undated query, pull no data (the query will be
	// `AND TRUE AND FALSE`)
	standardPart := "AND FALSE"
	if len(tabs.Include.Standard)+len(tabs.Exclude.Standard) > 0 {
		standardPart = "AND p.FieldID IN UNNEST(@StandardFieldIDs)"
		fieldIDs := tabs.AllStandardFields()
		params = append(params, bigquery.QueryParameter{Name: "StandardFieldIDs", Value: fieldIDs})
	}

	includePart := ""
	includedValues := tabs.AllIncluded()
	if len(includedValues) > 0 {
		i := 0
		for _, v := range includedValues {
			includePart = includePart + fmt.Sprintf("\nOR (hd.FieldID = @IncludeParts%d AND hd.value IN UNNEST(@IncludeParts%d) )", i, i+1)
			params = append(params, bigquery.QueryParameter{Name: fmt.Sprintf("IncludeParts%d", i), Value: v.FieldID})
			params = append(params, bigquery.QueryParameter{Name: fmt.Sprintf("IncludeParts%d", i+1), Value: v.FormattedValues()})
			i += 2
		}
	}

	excludePart := ""
	exludedValues := tabs.AllExcluded()
	if len(exludedValues) > 0 {
		i := 0
		for _, v := range exludedValues {
			excludePart = excludePart + fmt.Sprintf("\nOR (hd.FieldID = @ExcludeParts%d AND hd.value IN UNNEST(@ExcludeParts%d) )", i, i+1)
			params = append(params, bigquery.QueryParameter{Name: fmt.Sprintf("ExcludeParts%d", i), Value: v.FieldID})
			params = append(params, bigquery.QueryParameter{Name: fmt.Sprintf("ExcludeParts%d", i+1), Value: v.FormattedValues()})
			i += 2
		}
	}

	// Our query is dynamic, not static, so we assemble it from a text template,
	// which we will populate with a map:
	queryParts := map[string]interface{}{
		// True variables
		"database":             BQ.Database,
		"materializedDatabase": BQ.MaterializedDB,
		"use_gp":               BQ.UseGP,

		// Composed chunks of query
		"standardPart": standardPart,
		"includePart":  includePart,
		"excludePart":  excludePart,
	}

	// Assemble all the parts (execute the template)
	populatedQuery := &strings.Builder{}
	if err := queryTemplate.Execute(populatedQuery, queryParts); err != nil {
		return nil, err
	}

	if displayQuery {

		fmt.Println(populatedQuery.String())
		fmt.Println("Query parameters:")

		lastV := ""
		for _, v := range params {
			if v.Name[:4] != lastV {
				fmt.Printf("==%s==\n", v.Name[:4])
				lastV = v.Name[:4]
			}

			if x, ok := v.Value.([]string); ok {
				fmt.Printf("AND hd.value IN UNNEST([\"%s\"]) )\n", strings.Join(x, `","`))
				continue
			}
			fmt.Printf("OR (hd.FieldID = %v ", v.Value)
		}
		return nil, nil
	}

	// Generate the bigquery query object, but don't call it
	bqQuery := BQ.Client.Query(populatedQuery.String())
	bqQuery.QueryConfig.Parameters = append(bqQuery.QueryConfig.Parameters, params...)

	return bqQuery, nil
}
