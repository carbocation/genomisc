# TraceOverlay

`go get github.com/broadinstitute/ml4cvd/jamesp/traceoverlay/web`

## Running:
1. Build
1. Pass the folder name for your current overlay
1. Pass a (subsetted) manifest file

```sh
# Build:
go build -o to.linux *.go

# Then run:
./to.linux \
  -manifest /tmp/trace-test/trace-test.tsv \
  -project teapot
```

### Manifest
Since every entry in the manifest will be listed, you should aggressively prune the list of entries you want to deal with before launching traceoverlay.

### Project
A folder with this name will be created in the same path where you are running traceoverlay. The purpose of "projects" is to allow you to trace different structures with different projects: e.g., the `rv` with one project, the `lv` with another, `aorta`, etc. TODO: Need to consider whether this is the optimal approach. If we end up using this tool frequently, then knowing which files have been annotated *for each project* may be important, but currently that is not possible since only one project at a time is consumed.

## Output
Each traced overlay is output as `zipfilename_dicomname.png`. They are PNGs because go's BMP writer made these all black, even though they are natively transparent BMPs...