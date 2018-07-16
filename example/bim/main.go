package main

import (
	"flag"
	"log"

	"github.com/carbocation/plink"
)

func main() {
	path := flag.String("path", "", "Path to a .bim file")
	flag.Parse()

	if *path == "" {
		flag.PrintDefaults()
		log.Fatalln("No path provided")
	}

	b, err := plink.OpenBIM(*path)
	if err != nil {
		log.Fatalln(err)
	}
	defer b.Close()

	j := 0
	for v, i := b.Read(), 0; v != nil; v, i = b.Read(), i+1 {
		// log.Printf("%d) %+v\n", i, v)
		j++
	}
	if b.Err() != nil {
		log.Fatalln(err)
	}

	log.Println(j, "variants")
}
