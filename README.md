# genomisc
Miscellaneous genomics tools and data structures in golang

# HWE
`go get github.com/carbocation/genomisc/hwe`

HWE computes `Exact` or `Approximate` Hardy-Weinberg P value. `Fast` computes
the approximate P value, then if it is significant according to your threshold,
it computes the `Exact` P-value to be certain. This is a pure go implementation
with naive algorithms, which can therefore be slow. It uses `big.Int` and can
handle extremely large sample sizes (~hundreds of thousands).

# RAMCSV
`go get github.com/carbocation/genomisc/ramcsv`

RAMCSV consumes a file handle and a csv.Reader (which exists only to provide
your csv parsing settings) and allows you to seek back and forth through a CSV
file to any line in the file, without having to actually keep the full file in
memory.

# PRSParser
`go get github.com/carbocation/genomisc/prsparser`

PRSParser consumes a variety of formats for polygenic scores.