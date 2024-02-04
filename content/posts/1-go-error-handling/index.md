---
title: "How do you even handle errors in Go?"
date: 2024-01-18
draft: false
description: "The many ways of error handling in Golang"
slug: "custom-error-handling-in-go"
tags: ["Golang"]
showTableOfContents: true
---

**Error handling** is one of the things I've struggled with the most while learning Go. The issue lies in **where** and **how** to handle these errors. It's mostly agreed upon that it should be done separately from the main application, but there are many ways to do this and it's hard to decide which one is the 'best practice'. So I thought: *why not make a compilation of them?* 

Now, of course I know that it differs from project to project and there is no 'best practice'. I just made this for fun 😁.

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
In this example the `thisFunction` calls `someFunction`, which calls another function `stringToInt` that returns an error by default. Each of these errors are followed by the name of the function below it.
```
this function: some function: string to int: strconv.Atoi: parsing "five": invalid syntax
```

This is extremely useful to detect errors that do not return the filename and line of where the error occurred (eg: `app/cmd/main.go:23 main.function()`) and in applications that have many layers. I found it very useful in my project [clothera](/kietpa.id/projects/#clothera), a basic CLI app made with native Go.

Daisy-chaining worked perfectly in that project because a CLI app didn't need to send these errors to higher layer to be handled if it wasn't needed. All you need to do was `fmt.Println(err)` at the same function and it appears in the terminal.

This is why I abandoned this way of error handling once I started building APIs. From seeing other people's code, it was common practice to *standardize* errors by creating custom ones and handling these errors should ideally be *centralized* in one layer. In the top layer, asserting the type of error would be difficult. Also, tracing errors could be done through [logging package](https://github.com/sirupsen/logrus/issues/63#issuecomment-433746888) using [flags](https://stackoverflow.com/questions/24809287/how-do-you-get-a-golang-program-to-print-the-line-number-of-the-error-it-just-ca) and whatnot. 

So, daisy-chaining seemed a bit obsolete. If I really wanted to make it work, a lot of effort would be needed for little in return. I thought to just follow the common practice and spend my time learning other things. Maybe I'll come back to this later in the future.

## Custom Errors 

The next error handling method I learned was custom errors. It's not exactly a method, but a *standard*. A separate file would be created to store these errors, also known as **error contracts**. These contracts would then be used to replace common, easily handled errors without needing to parse the error any further as there is already a defined global variable.

```go
import "errors"

var (
	ErrBadRequest      = errors.New("bad request")
	ErrInternalFailure = errors.New("internal failure")
	ErrNotFound        = errors.New("not found")
	ErrFailedBind      = errors.New("failed bind json")
	ErrUnauthorized    = errors.New("access unauthorized")
)
...
```
A basic implementation would include an array of custom errors created by `errors.New()`, which are then easily identfied using the `errors.Is()` and `errors.As()` functions. The errors are passed from the internals to the controllers to be handled in a specific way. For example, `ErrFailedBind` would indicate that the actual error needed to be parsed to display which of the input fields were invalid, whereas `ErrUnauthorized` would return the same message to the user everytime. Below is an example of parsing the `ErrFailedBind`.

```go
var Validate *validator.Validate = validator.New()

// handle binding errors
func ErrorBind(err error) string {
	var ve validator.ValidationErrors
	out := ""
	if errors.As(err, &ve) {
		for _, fe := range ve {
			out = fe.Field() + ": " + msgForTag(fe.Tag())
		}
		return out
	}
	return out
}

// error fields
func msgForTag(tag string) string {
	switch tag {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email"
	case "alpha":
		return "Must be alphabetical"
	case "gte":
		return "Input too short"
	}
	return ""
}
```
### A New Error Type
In my experience, a more convenient way of implementing this in an API is to create a whole new error type `APIError` that includes the HTTP status code along with the error message.

```go

type APIError struct {
	Code    int
	Message string
}

var (
	ErrInternalServer = APIError{
		Code:    http.StatusInternalServerError, //500
		Message: "Internal Server Error",
	}

	ErrDataNotFound = APIError{
		Code:    http.StatusOK, //200
		Message: "Data Not Found",
	}

	ErrBadRequest = APIError{
		Code:    http.StatusBadRequest, //400
		Message: "Bad request",
	}

	ErrUnauthorized = APIError{
		Code:    http.StatusUnauthorized, //401
		Message: "Request Unauthorized",
	}
)
```
These errors are used all over the handlers, where they are entered into a simple `ErrorMessage` function that logs and returns the error. One of my mentors provided this very simple template in Gin, which he used in his previous job (in a large company).
```go
import "github.com/gin-gonic/gin"

func ErrorMessage(c *gin.Context, apiError *APIError, err error) *gin.Context {
    log.Println(err)
	c.Abort()
	c.JSON(apiError.Code, gin.H{"error": APIError{
		Code:    apiError.Code,
		Message: apiError.Message,
	}})
	return c
}
```
An example handler:
```go
func AddProduct(c *gin.Context) {
	var product entity.Product

	err := c.BindJSON(&product)
	if err != nil {
		utils.ErrorMessage(c, &utils.ErrBadRequest, err)
		return
	}

	config.DB.Create(&product)

	c.JSON(200, product)
}
```
While this simple approach works well, the way it handles the application error (the actual error) is lacking. There needs to be another function added in some of the handlers which makes it a little messy. 

### Implementing the error Interface

A way to expand on the new error type is to implement the error interface by giving the type an `Error()` method that returns a string of the error. By doing this, it allows you to do some type assertion with `errors.As()` and handle the app errors semi-*gracefully*.

```go
type error interface {
	Error() string
}
```

In my [Hotel 626](https://github.com/kietpa/hotel-626) project, I used an error type that contains the generic http error and the actual application error, named service error (svcErr) and app error (appErr).

```go
type Error struct {
	svcErr error // generic http error
	appErr error // actual error
}

func NewError(svcErr, appErr error) error {
	return Error{
		svcErr: svcErr,
		appErr: appErr,
	}
}

// to be included in the error interface
func (e Error) Error() string {
	return errors.Join(e.svcErr, e.appErr).Error()
}
```

In the service layer, the two errors would be wrapped in the Error struct using the `NewError` function and sent to the handlers above to be parsed. The new error is accepted as a normal error type which makes this convenient.

```go
type APIError struct {
	Status  int
	Message string
}

func FromError(err error) APIError {
	var svcError Error
	var apiError APIError

	// errors as checks if both are type utils.Error
	if errors.As(err, &svcError) {
		// set actual error on message
		apiError.Message = svcError.AppError().Error()

		// check error
		svcErr := svcError.ServiceError()
		switch svcErr {
		case ErrFailedBind:
			apiError.Message = ErrorBind(svcError.AppError()) // check which field fails validation
			apiError.Status = http.StatusBadRequest
		case ErrBadRequest:
			apiError.Status = http.StatusBadRequest
		case ErrInternalFailure:
			apiError.Status = http.StatusInternalServerError
		case ErrNotFound:
			apiError.Status = http.StatusNotFound
		case ErrUnauthorized:
			apiError.Status = http.StatusUnauthorized
		}
	}

	return apiError
}
```

In the handlers, the error is parsed by the `FromError` function which is at the core of this pattern. This function checks whether the received error contains the appErr and SvcErr field using `errors.As`. If it does, the error wil by casted to the `APIError` struct that displays the status code and message for the user. This message will be unique for certain errors such as `ErrFailedBind` which shows which input field is invalid. For other errors, the message will be kept uniform to prevent unwanted information leaks.

In cases where the error check fails (usually in a panic), the error is immediately caught by a middleware. In the project I used the one provided by *Echo* with a basic logger. See the full [utils package](https://github.com/kietpa/hotel-626/tree/main/utils) for more details.

```go
// located in main
e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
	StackSize:    1 << 10, // 1 KB
	LogErrorFunc: utils.LogError,
}))

// in the utils package
func LogError(c echo.Context, err error, stack []byte) error {
	log.Println(err)
	return echo.NewHTTPError(http.StatusInternalServerError, err)
}
```

This method of error handling is the one that feels the most comfortable to me, because new errors and methods can be easily added to the error contracts and the switch in the `FromError` function. 

However, there is a flaw in the fact that all these errors are located in the utils package/folder and every part of application that uses them are dependant on them, creating tight coupling between packages.

## The Coupling Problem

In software engineering, [coupling](https://en.wikipedia.org/wiki/Coupling_(computer_programming)) or dependency is the degree to which each program module relies on each one of the other modules. This means when a system A is tightly coupled with B, if A changes, B must also change. Tight coupling is usually an unwanted quality to have in an API, as APIs wants its components to be flexible, reusable, and adaptable. 

The error package I described might not be suited for that since it uses **type-assertion** where the type of the error is checked. This means that the error types/contracts must be public, which creates a strong coupling with the caller, making for a brittle API. It also violates the **Open-closed principle** from the SOLID principles. [This blog post](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully) by Dave Cheney explains these issues much better than I can.

So what's the solution? Well, in the same post Cheney said that we should check errors by **behaviour** instead of type. In this case, the error package only needs to be in the top layer with the handlers. He provides some methods to do this, but I personally have never seen this being done in an actual application. It just seems to hard to implement in some cases.

The closest thing I've seen is not including the full error package in the service. In each service or domain, only the needed error contracts exist. For example, `ErrUnauthorized` only exists in the user domain.

So, should we check by behaviour or type? The answer probably lies in between. Use type when behaviour doesn't work. Though, I've never really seen this strongly applied anywhere. I might be not advanced enough to see or understand more complex implementations.

## Conclusion (I'm confused)

In the end, I am still confused on what 'perfect' error handling looks like in Go. Do big companies have a more advanced system? Does it even exist?

Either way, I think sticking to simple custom errors is enough for now. 


## Sources
- https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully
- https://blog.dreamfactory.com/the-importance-of-loose-coupling-in-rest-api-design/
- https://en.wikipedia.org/wiki/Coupling_(computer_programming)
- https://www.youtube.com/watch?v=CxcxRgwWtAk