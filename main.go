package main

import "fmt"

func main() {
	e, err := NewEngine()
	if err != nil {
		panic(err)
	}
	e.Set("foo", "bar")
	value, err := e.Get("foo")
	if err != nil {
		panic(err)
	}
	fmt.Println(value)
}
