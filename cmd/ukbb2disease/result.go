package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/carbocation/pfx"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

type Result struct {
	SampleID                int64                `bigquery:"sample_id"`
	HasDisease              bigquery.NullInt64   `bigquery:"has_disease"`
	IncidentDisease         bigquery.NullInt64   `bigquery:"incident_disease"`
	PrevalentDisease        bigquery.NullInt64   `bigquery:"prevalent_disease"`
	MetExclusion            bigquery.NullInt64   `bigquery:"met_exclusion"`
	PhenotypeDateCensor     bigquery.NullDate    `bigquery:"date_censor"`
	RoughPhenotypeAgeCensor bigquery.NullFloat64 `bigquery:"age_censor"`      // Note: just uses days/365. Don't use.
	PhenotypeAgeCensorDays  bigquery.NullFloat64 `bigquery:"age_censor_days"` // This one uses proper days and is directly reliable.
	BirthDate               bigquery.NullDate    `bigquery:"birthdate"`
	EnrollDate              bigquery.NullDate    `bigquery:"enroll_date"`
	EnrollAge               bigquery.NullFloat64 `bigquery:"enroll_age"`
	EnrollAgeDays           bigquery.NullFloat64 `bigquery:"enroll_age_days"`
	HasDied                 bigquery.NullInt64   `bigquery:"has_died"`
	DeathDate               bigquery.NullDate    `bigquery:"death_date"`
	DeathAge                bigquery.NullFloat64 `bigquery:"death_age"`
	DeathAgeDays            bigquery.NullFloat64 `bigquery:"death_age_days"`
	ComputedDate            bigquery.NullDate    `bigquery:"computed_date"`
	MissingFields           bigquery.NullString  `bigquery:"missing_fields"`
}

// PhenotypeAgeCensor computes the days since the date of birth, then divides by
// 365.25 to get the phenotype censor age in years.
func (r Result) PhenotypeAgeCensor() (bigquery.NullFloat64, error) {
	if !r.BirthDate.Valid || !r.PhenotypeDateCensor.Valid {
		return bigquery.NullFloat64{}, nil
	}

	res := bigquery.NullFloat64{
		Float64: float64(r.PhenotypeDateCensor.Date.DaysSince(r.BirthDate.Date)) / 365.25,
		Valid:   true,
	}
	if res.Float64 < 0 {
		return res, fmt.Errorf("censoring happened before birth: %v", res)
	}

	return res, nil
}

// DeathAgeCensor computes the days since the date of birth, then divides by
// 365.25 to get the death censor age in years.
func (r Result) DeathAgeCensor() (bigquery.NullFloat64, error) {
	if !r.BirthDate.Valid || !r.DeathDate.Valid {
		return bigquery.NullFloat64{}, nil
	}

	res := bigquery.NullFloat64{
		Float64: float64(r.DeathDate.Date.DaysSince(r.BirthDate.Date)) / 365.25,
		Valid:   true,
	}
	if res.Float64 < 0 {
		return res, fmt.Errorf("death censoring happened before birth: %v", res)
	}
	return res, nil
}

func ExecuteQuery(BQ *WrappedBigQuery, query *bigquery.Query, diseaseName string, missingFields []string) error {
	itr, err := query.Read(BQ.Context)
	if err != nil {
		return pfx.Err(fmt.Sprint(err.Error(), query.Parameters))
	}
	todayDate := time.Now().Format("2006-01-02")
	missing := strings.Join(missingFields, ",")
	fmt.Fprintf(STDOUT, "disease\tsample_id\thas_disease\tincident_disease\tprevalent_disease\tmet_exclusion\tcensor_date\tcensor_age\tcensor_age_days\tbirthdate\tenroll_date\tenroll_age\tenroll_age_days\thas_died\tdeath_censor_date\tdeath_censor_age\tdeath_censor_age_days\tcensor_computed_date\tcensor_missing_fields\tcomputed_date\tmissing_fields\n")
	for {
		var r Result
		err := itr.Next(&r)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return pfx.Err(err)
		}

		censoredPhenoAge, err := r.PhenotypeAgeCensor()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: setting censor_age to enroll_age for %d (birthdate %s phenotype date %s) because %s\n", diseaseName, r.SampleID, r.BirthDate, r.PhenotypeDateCensor, err.Error())

			// UK Biobank uses impossible values (e.g., 1900-01-01) to indicate that
			// the date is not known. See, e.g., FieldID 42000. This does not mean
			// that the value is illegal, so it shouldn't be null. Instead, it should
			// be some legal value. Here, we set the age of incidence to be 0 years,
			// and we set the date of incidence to be the birthdate.
			censoredPhenoAge = r.EnrollAge
			r.PhenotypeAgeCensorDays = r.EnrollAgeDays
			r.PhenotypeDateCensor = r.EnrollDate
		}

		censoredDeathAge, err := r.DeathAgeCensor()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: setting death_censor_age to enroll_age for %d (birthdate %s death date %s) because %s\n", diseaseName, r.SampleID, r.BirthDate, r.DeathDate, err.Error())

			// UK Biobank uses impossible values (e.g., 1900-01-01) to indicate that
			// the date is not known. See, e.g., FieldID 42000. This does not mean
			// that the value is illegal, so it shouldn't be null. Instead, it should
			// be some legal value. Here, we set the age of incidence to be 0 years,
			// and we set the date of incidence to be the birthdate.
			censoredDeathAge = r.EnrollAge
			r.DeathAgeDays = r.EnrollAgeDays
			r.DeathDate = r.EnrollDate
		}

		fmt.Fprintf(STDOUT, "%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			diseaseName, r.SampleID, NA(r.HasDisease), NA(r.IncidentDisease), NA(r.PrevalentDisease), NA(r.MetExclusion), NA(r.PhenotypeDateCensor), NA(censoredPhenoAge), NA(r.PhenotypeAgeCensorDays), NA(r.BirthDate), NA(r.EnrollDate), NA(r.EnrollAge), NA(r.EnrollAgeDays), NA(r.HasDied), NA(r.DeathDate), NA(censoredDeathAge), NA(r.DeathAgeDays), NA(r.ComputedDate), NA(r.MissingFields), todayDate, missing)
	}

	return nil
}

// NA emits an empty string instead of "NULL" since this plays better with
// BigQuery
func NA(input interface{}) interface{} {
	invalid := ""

	switch v := input.(type) {
	case bigquery.NullInt64:
		if !v.Valid {
			return invalid
		}
	case bigquery.NullFloat64:
		if !v.Valid {
			return invalid
		}
	case bigquery.NullString:
		if !v.Valid {
			return invalid
		}
	case bigquery.NullDate:
		if !v.Valid {
			return invalid
		}
	}

	return input
}

func BuildQuery(BQ *WrappedBigQuery, tabs *TabFile, displayQuery bool, biobankSource string) (*bigquery.Query, error) {
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
			params = append(params, bigquery.QueryParameter{Name: fmt.Sprintf("IncludeParts%d", i+1), Value: v.FormattedValues(biobankSource)})
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
			params = append(params, bigquery.QueryParameter{Name: fmt.Sprintf("ExcludeParts%d", i+1), Value: v.FormattedValues(biobankSource)})
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

	// Assemble all the parts
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

	// (execute the template)
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
