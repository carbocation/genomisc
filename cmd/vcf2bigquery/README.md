# Compiling vcf2bigquery

This tool statically links the chromosome start/stop for grch37 and grch38. To do so, it uses packr2. This requires one additional startup step and one additional cleanup step.

To make sure you have the right packages installed:
```sh
go get -u
```

To build:
```sh
packr2

GOOS=linux go build -o vcf2bigquery.linux *.go

packr2 clean
```