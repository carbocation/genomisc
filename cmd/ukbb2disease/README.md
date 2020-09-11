# ukbb2disease
ukbb2disease computes derived phenotypes, with incidence, prevalence, death, exclusion due to a competing hazard, and censoring, from disease definition files called "tabfiles" (Seung Hoan Choi's original phenotype definition format). As a brief review, tabfiles contain 3 tab-delimited columns: Field, Coding, and exclude. E.g.:

```
Field	Coding	exclude
20002	1076,1079	0
```

*Field* is the UK Biobank FieldID (e.g., `phenotype.FieldID`)

*Coding* is the UK Biobank value (e.g., `phenotype.value` or `coding.coding`)

*Exclude* is whether the row represents an exclusion criterion (1) or an inclusion criterion (0)

# Install updated dependencies
`go get -u`

# Build (examples)
`go build -o ukbb2disease.osx *.go`

`GOOS=linux go build -o ukbb2disease.linux *.go`

# Google Cloud application default credentials
Note that `ukbb2disease` points to a Google Cloud database. It requires that your application default credentials be configured. If you are running it on a GCP virtual machine, this is automatically done. If you are running the program locally, the Google Cloud application default credentials can be configured by running `gcloud auth applicaiton-default login`.

# Database dependencies
This requires the materialized tables (defined in the SQL files in the `ukbb2csv` directory) to exist in tables with the same name as their filename (except the suffix).

# Which UK Biobank fields are understood by the program?
These fields can be listed by running `ukbb2disease -verbose`

# How are exclusions handled?
(If your disease definition does not have exclusions, then this section can be ignored.) For incident disease (disease occurring after UK Biobank enrollment), there are 3 possible states with exclusion criteria:
1. Neither an exclusion nor an inclusion criterion were met.
1. An inclusion criterion was met, and then subsequently an exclusion criterion either was or was not met.
1. An exclusion criterion was met, and then subsequently an inclusion criterion either was or was not met.

As an example, imagine a disease definition of non-ischemic cardiomyopathy (NICM). You may define this as dilated cardiomyopathy in the absence of coronary artery disease (CAD). So, coronary artery disease is an exclusion criterion. However, it is possible for someone with non-ischemic cardiomyopathy to, eventually, develop coronary artery disease years later. So, in a study like the UK Biobank, it would not be reasonable to reach into the future, ascertain someone's future coronary disease status, and exclude them from being considered to have non-ischemic coronary artery disease in the past.

Now going through those 3 categories with the NICM example in mind:
1. If someone never develops a NICM diagnosis nor a CAD diagnosis, they will have an `incident_disease` field set to 0 and a `met_exclusion` field set to 0.
1. If someone develops a diagnosis of NICM, and then years later gets a CAD diagnosis, they will have an `incident_disease` field set to 1 and a `met_exclusion` field set to 0. Even though they eventually met an exclusion criterion, they did not do so prior to disease diagnosis.
1. If someone first develops a diagnosis of CAD, then regardless of whether they subsequently meet an inclusion criterion for NICM, their `incident_disease` field will be marked as 0. To identify this circumstance, note that the `met_exclusion` field will be set as 1.

So, to summarize:
* `incident_disease`:0 and `met_exclusion`:0 means never developed disease nor hit the exclusion criterion
* `incident_disease`:1 and `met_exclusion`:0 means developed the disease, and did not hit the exclusion criterion prior to disease onset (and may or may not have hit the exclusion criterion subsequently)
* `incident_disease`:0 and `met_exclusion`:1 means developed the exclusion criterion, and did not develop the disease prior to the exclusion criterion (and may or may not have hit the disease inclusion criterion subsequently)
