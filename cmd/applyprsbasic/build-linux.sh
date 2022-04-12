# Statically link, so that if CGO is used, it doesn't dynamically link to 
# a GCC on the local system which may have a different version.
GOOS=linux GOARCH=amd64 go build -ldflags="-extldflags=-static" -o applyprsbasic.linux

# And build a copy without cgo.
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o applyprsbasic_nocgo.linux
