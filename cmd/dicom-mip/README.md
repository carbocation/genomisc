dicom-mip creates a projection of a stack of dicom files that are within a UK Biobank-structured .zip.

E.g.:
`go run *.go -sequence_column_name series_number -folder gs://bulkml4cvd/bodymri/all/raw -manifest demo2.tsv -donotsort=true`