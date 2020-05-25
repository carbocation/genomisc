package main

import (
	"fmt"

	"google.golang.org/api/iterator"
)

type sexEthnicity struct {
	SampleID  int64  `bigquery:"sample_id"`
	Sex       string `bigquery:"sex"`
	Ethnicity string `bigquery:"ethnicity"`
}

func BigQuerySexEthnicity(wbq *WrappedBigQuery) (map[int64]sexEthnicity, error) {
	out := make(map[int64]sexEthnicity)

	query := wbq.Client.Query(fmt.Sprintf(`
WITH sex AS (
	SELECT 
		sample_id, 
		c.meaning sex
	FROM %s.phenotype p
	JOIN %s.dictionary d USING(FieldID)
	JOIN %s.coding c ON c.coding_file_id = d.coding_file_id AND c.coding = p.value
	WHERE TRUE
		-- Sex is not instanced
		AND d.FieldID = 31
), ethnicity AS (
	SELECT 
		sample_id, 
		SAFE.REPLACE(c.meaning, " ", "_") ethnicity
	FROM %s.phenotype p
	JOIN %s.dictionary d USING(FieldID)
	JOIN %s.coding c ON c.coding_file_id = d.coding_file_id AND c.coding = p.value
	WHERE TRUE
		AND d.FieldID = 21000
		AND p.instance = 0
), all_samples AS (
	SELECT 
		sample_id
	FROM %s.phenotype
	GROUP BY sample_id
)
	
SELECT 
	sample_id,
	CASE WHEN sex.sex IS NULL THEN 'NA' ELSE sex.sex END sex,
	CASE WHEN ethnicity.ethnicity IS NULL THEN 'NA' ELSE ethnicity.ethnicity END ethnicity
FROM all_samples
LEFT JOIN sex USING(sample_id)
LEFT JOIN ethnicity USING(sample_id)
`, wbq.Database, wbq.Database, wbq.Database, wbq.Database, wbq.Database, wbq.Database, wbq.Database))

	itr, err := query.Read(wbq.Context)
	if err != nil {
		return nil, err
	}
	for {
		var values sexEthnicity
		err := itr.Next(&values)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		// Take only the first, since we use this for things like enrollment
		// date. If someone came to a follow-up visit, we don't want to say that
		// they "enrolled" at the time of their follow-up, for example. Relies
		// on sort order specified above in the query.
		if _, exists := out[values.SampleID]; exists {
			continue
		}
		out[values.SampleID] = values
	}

	return out, nil
}
