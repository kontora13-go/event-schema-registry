package main

import (
	"github.com/kontora13-go/event-schema-registry/schemagen/gen"
	"log"
	"os"
)

func main() {
	source := os.Args[1]
	target := os.Args[2]

	log.Printf("%s -> %s", source, target)
	if source == "" || target == "" {
		return
	}

	g := gen.NewGen(source, target)

	err := g.Generate()
	if err != nil {
		log.Fatal(err)
	}
}
