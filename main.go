package main

import (
	"fmt"
	"strconv"
)

func thisFunction() error {
	err := someFunction()
	return fmt.Errorf("this function: %w", err)
}

func someFunction() error {
	_, err := stringToInt()
	return fmt.Errorf("some function: %w", err)
}

func stringToInt() (int, error) {
	number, err := strconv.Atoi("five")
	if err != nil {
		return 0, fmt.Errorf("string to int: %w", err)
	}
	return number, nil
}

func main() {
	err := thisFunction()
	if err != nil {
		fmt.Println(err)
	}
}
