package main

import "fmt"

type Animal interface {
	sound()
}

type Dog struct {
	Animal
}

type Cat struct {
	Animal
}

func (d *Dog) sound() {
	fmt.Println("Гав гав гав")
}

func (c *Cat) sound() {
	fmt.Println("Мяу мяу мяу")
}

func (d *Dog) soun321d() {
	fmt.Println("Гав гав гав")
}

func (c *Cat) sou312nd() {
	fmt.Println("Мяу мяу мяу")
}

func main() {
	var dog Animal = &Dog{}
	var cat Animal = &Cat{}

	dog.sound()
	cat.sound()
}
