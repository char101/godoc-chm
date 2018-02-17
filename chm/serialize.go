package chm

import (
	"fmt"
	"io/ioutil"
	"log"
)

// Serializer can be serialized
type Serializer interface {
	Serialize(b *Buffer)
}

// Save saves the serializer content into a file
func Save(s Serializer, filename string) {
	var b Buffer
	s.Serialize(&b)

	fmt.Println("Creating", filename)
	if err := ioutil.WriteFile(filename, b.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
}
