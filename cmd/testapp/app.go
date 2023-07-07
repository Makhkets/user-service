package main

import "fmt"

type animal struct {
	name string
	age  uint
}

type Animal interface {
	Go()
}

func (a *animal) Go() {
	fmt.Println(a.name, a.age)
}

func GetAnimalStruct() Animal {
	return &animal{
		name: "dog",
		age:  12,
	}
}

func main() {
	animal := GetAnimalStruct()
	animal.Go()
}
