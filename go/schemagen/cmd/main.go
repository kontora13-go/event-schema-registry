package main

import (
	"github.com/kontora13-go/event-schema-registry/schemagen/gen"
	"log"
)

func main() {
	/*
		schema, err := readCommonSchema("../schema/message/v1.json")
		if err != nil {
			log.Fatal(err)
		}
	*/

	//log.Println(schema)

	// TODO: from os.Args
	g := gen.NewGen("./message", "../schema")

	err := g.Generate()
	if err != nil {
		log.Fatal(err)
	}

}
