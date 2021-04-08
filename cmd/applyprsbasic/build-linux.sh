# Statically link, so that if CGO is used, it doesn't dynamically link to 
# a GCC on the local system which may have a different version.
GOOS=linux go build -ldflags="-extldflags=-static" -o applyprsbasic.linux *.go
