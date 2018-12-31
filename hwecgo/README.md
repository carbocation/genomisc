# HWECgo

HWECgo is a Go wrapper around Chris Chang's exact HWE P value calculator from
his stats package. Specifically, https://github.com/chrchang/stats at commit
67c3f71 (2014/04/03).

It is much faster than the go-based hwe package and should be used instead of
that, unless cgo cannot be used in a specific environment.