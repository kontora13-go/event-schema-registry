package main

import (
	"github.com/kontora13-go/event-schema-registry/schemagen/gen"
	"log"
	"os"
)

func main() {
	g := gen.NewGen(os.Args[0], os.Args[1])

	err := g.Generate()
	if err != nil {
		log.Fatal(err)
	}
}
