package main

import (
	"github.com/kontora13-go/event-schema-registry/go/gen"
	"log"
)

type Gen = gen.Gen

func NewGen(source string, target string) *Gen {
	return gen.NewGen(source, target)
}

func main() {
	//source := os.Args[1]
	//target := os.Args[2]

	source := "./message"
	target := "../schema"

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
