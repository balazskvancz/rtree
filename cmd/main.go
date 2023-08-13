package main

import (
	"fmt"
	"os"

	"github.com/balazskvancz/rtree"
)

type foo interface {
	toString() string
}

type s struct{}

func (s *s) toString() string {
	return "foobar"
}

func main() {
	t := rtree.New[foo]()

	const key string = "/foo"

	if err := t.Insert(key, &s{}); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	val := t.Find(key)

	if val == nil {
		fmt.Println("no match")
	}

	fmt.Println(val.GetValue().toString())
}
