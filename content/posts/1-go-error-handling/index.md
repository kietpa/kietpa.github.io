---
title: "Learning Error Handling in Go"
date: 2024-01-18
draft: false
description: "The many ways of error handling in Golang"
slug: "custom-error-handling-in-go"
tags: ["Golang"]
showTableOfContents: true
---

Error handling is one of the things I've struggled with the most while learning Go. The issue lies in **where** and **how** to handle these errors. It's mostly agreed upon that it should be done separately from the main application, but there are many ways to do this and it's hard to decide which one is the 'best practice'. So I thought: *why not make a compilation of them?* 

Now, of course I know that it differs from project to project and there is no 'best practice'. I just wanted to make a small compilation for myself for fun 😁.

Let's start with the basics.

## Daisy-chaining

One of the first things I learned in Go, **daisy-chaining** errors is when errors are wrapped with the method or function that they are in. This is used to make *debugging* easier as errors show a clear *trace* when returned. I was really attracted to this idea due to my background of working with PLCs, and the whole concept of a daisy-chain was just neat.

```go
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
```
In this example the `thisFunction` calls `someFunction`, which calls another function `stringToInt` that returns an error by default. Each of these errors are wrapped with the name of the function that can be unwrapped with `errors.Unwrap` or `errors.Is`.
```
this function: some function: string to int: strconv.Atoi: parsing "five": invalid syntax
```

This is extremely useful to detect errors that do not return the filename and line of where the error occurred (eg: `app/cmd/main.go:23 main.function()`) and in applications that have many layers. I found it very useful in my project [clothera](https://github.com/kiet-asmara/clothera), a basic CLI app made with native Go.

Daisy-chaining worked perfectly in that project because a CLI app didn't need to send these errors to higher layer to be handled if it wasn't needed. All you need to do was `fmt.Println(err)` at the same function and it appears in the terminal.

This is why I somewhat abandoned this way of error handling once I started building APIs. From seeing other people's code, it was common practice to *standardize* errors by creating custom ones and handling these errors should ideally be *centralized* in one layer. Also, tracing errors could be done through [logging package](https://github.com/sirupsen/logrus/issues/63#issuecomment-433746888) using [flags](https://stackoverflow.com/questions/24809287/how-do-you-get-a-golang-program-to-print-the-line-number-of-the-error-it-just-ca) and whatnot. 

So, daisy-chaining seemed a bit obsolete. If I really wanted to make it work, a lot of effort would be needed for little in return. I thought to just follow the common practice and spend my time learning other things. I'll probably come back to this later in the future.

## Standardized Errors 

The next thing I 
