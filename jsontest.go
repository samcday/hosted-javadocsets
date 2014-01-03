package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Foo struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	decoder := json.NewDecoder(strings.NewReader(`{"name": "Sam", "age": 24}`))
	var m Foo
	if err := decoder.Decode(&m); err != nil {
		panic(err)
	}
	fmt.Println(m.Name)
}
