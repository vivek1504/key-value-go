package main

import "fmt"

func main() {
	e := NewEngine()
	e.Set("foo", "bar")
	value, err := e.Get("foo")
	if err != nil {
		panic(err)
	}
	fmt.Println(value)
}
