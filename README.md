##LARS
<img align="right" src="https://raw.githubusercontent.com/go-playground/lars/master/examples/README/test.gif">
![Project status](https://img.shields.io/badge/version-2.0-green.svg)
[![Build Status](https://semaphoreci.com/api/v1/projects/4351aa2d-2f94-40be-a6ef-85c248490378/679708/badge.svg)](https://semaphoreci.com/joeybloggs/lars)
[![Coverage Status](https://coveralls.io/repos/github/go-playground/lars/badge.svg?branch=master)](https://coveralls.io/github/go-playground/lars?branch=master)
[![Go Report Card](https://goreportcard.com/badge/go-playground/lars)](https://goreportcard.com/report/go-playground/lars)
[![GoDoc](https://godoc.org/github.com/go-playground/lars?status.svg)](https://godoc.org/github.com/go-playground/lars)
![License](https://img.shields.io/dub/l/vibe-d.svg)

LARS is a fast radix-tree based, zero allocation, HTTP router for Go.  [ view examples](https://github.com/go-playground/lars/tree/master/examples)

Why Another HTTP Router?
------------------------
Have you ever been painted into a corner by a framework, **ya me too!** and I've noticed that allot of routers out there, IMHO, are adding so much functionality that they are turning into Web Frameworks, (which is fine, frameworks are important) however, not at the expense of flexibility and configurability. So with no further ado, introducing LARS an HTTP router that can be your launching pad in creating a framework for your needs. How? Context is an interface [see example here](https://github.com/go-playground/lars/blob/master/examples/all-in-one/main.go), where you can add as little or much as you want or need and most importantly...under your control. ( I will be creating a full example app in the near future that can be used as a starting point for any project. )

Unique Features 
--------------
- [x] Context is an interface allowing passing of framework/globals/application specific variables. [example](https://github.com/go-playground/lars/blob/master/examples/all-in-one/main.go)
- [x] Contains helpful logic to help prevent adding bad routes, keeping your url's consistent.
  * i.e. /user/:id and /user/:user_id - the second one will fail to add letting you know that :user_id should be :id
- [x] Has an uber simple middleware + handler definitions!!! middleware and handlers actually have the exact same definition!
- [x] Can register custom handlers for making other middleware + handler patterns usable with this router
  * best part about this is can register one for your custom context and not have to do type casting everywhere [see here](https://github.com/go-playground/lars/blob/master/examples/custom-handler/main.go)
- [x] Full support for standard/native http Handler + HandlerFunc [see here](https://github.com/go-playground/lars/blob/master/examples/native/main.go)
  * When Parsing a form call Context's ParseForm amd ParseMulipartForm functions and the URL params will be added into the Form object, just like query parameters are, so no extra work


**Note:** Since this router has only explicit matches, you can not register static routes and parameters for the same path segment. For example you can not register the patterns /user/new and /user/:user for the same request method at the same time. The routing of different request methods is independent from each other. I was initially against this, and this router allowed it in a previous version, however it nearly cost me in a big app where the dynamic param value say :type actually could have matched another static route and that's just too dangerous, so it is no longer allowed.

Installation
-----------

Use go get 

```go
go get github.com/go-playground/lars
``` 

or to update

```go
go get -u github.com/go-playground/lars
``` 

Then import lars package into your code.

```go
import "github.com/go-playground/lars"
``` 

Usage
------
Below is a full example, for a simpler example [see here](https://github.com/go-playground/lars/blob/master/examples/groups/main.go)
```go
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/lars"
)

// This is a contrived example of how I would use in production
// I would break things into separate files but all here for simplicity

// ApplicationGlobals houses all the application info for use.
type ApplicationGlobals struct {
	// DB - some database connection
	Log *log.Logger
	// Translator - some i18n translator
	// JSON - encoder/decoder
	// Schema - gorilla schema
	// .......
}

// Reset gets called just before a new HTTP request starts calling
// middleware + handlers
func (g *ApplicationGlobals) Reset() {
	// DB = new database connection or reset....
	//
	// We don't touch translator + log as they don't change per request
}

// Done gets called after the HTTP request has completed right before
// Context gets put back into the pool
func (g *ApplicationGlobals) Done() {
	// DB.Close()
}

func newGlobals() *ApplicationGlobals {

	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	// translator := ...
	// db := ... base db connection or info
	// json := ...
	// schema := ...

	return &ApplicationGlobals{
		Log: logger,
		// Translator: translator,
		// DB: db,
		// JSON: json,
		// schema:schema,
	}
}

// MyContext is a custom context
type MyContext struct {
	*lars.Ctx  // a little dash of Duck Typing....
	AppContext *ApplicationGlobals
}

// RequestStart overriding
func (mc *MyContext) RequestStart(w http.ResponseWriter, r *http.Request) {

	// call lars context reset, must be done
	mc.Ctx.RequestStart(w, r)
	mc.AppContext.Reset()
}

// RequestEnd overriding
func (mc *MyContext) RequestEnd() {
	mc.AppContext.Done()
	mc.Ctx.RequestEnd()
}

func newContext(l *lars.LARS) lars.Context {
	return &MyContext{
		Ctx:        lars.NewContext(l),
		AppContext: newGlobals(),
	}
}

func castCustomContext(c lars.Context, handler lars.Handler) {

	// could do it in all one statement, but in long form for readability
	h := handler.(func(*MyContext))
	ctx := c.(*MyContext)

	h(ctx)
}

func main() {

	l := lars.New()
	l.RegisterContext(newContext) // all gets cached in pools for you
	l.RegisterCustomHandler(func(*MyContext) {}, castCustomContext)
	l.Use(Logger)

	l.Get("/", Home)

	users := l.Group("/users")
	users.Get("", Users)

	// you can break it up however you with, just demonstrating that you can
	// have groups of group
	user := users.Group("/:id")
	user.Get("", User)
	user.Get("/profile", UserProfile)

	http.ListenAndServe(":3007", l.Serve())
}

// Home ...
func Home(c *MyContext) {

	var username string

	// username = c.AppContext.DB.find(user by .....)

	c.AppContext.Log.Println("Found User")

	c.Response().Write([]byte("Welcome Home " + username))
}

// Users ...
func Users(c *MyContext) {

	c.AppContext.Log.Println("In Users Function")

	c.Response().Write([]byte("Users"))
}

// User ...
func User(c *MyContext) {

	id := c.Param("id")

	var username string

	// username = c.AppContext.DB.find(user by id.....)

	c.AppContext.Log.Println("Found User")

	c.Response().Write([]byte("Welcome " + username + " with id " + id))
}

// UserProfile ...
func UserProfile(c *MyContext) {

	id := c.Param("id")

	var profile string

	// profile = c.AppContext.DB.find(user profile by .....)

	c.AppContext.Log.Println("Found User Profile")

	c.Response().Write([]byte("Here's your profile " + profile + " user " + id))
}

// Logger ...
func Logger(c lars.Context) {

	start := time.Now()

	c.Next()

	stop := time.Now()
	path := c.Request().URL.Path

	if path == "" {
		path = "/"
	}

	log.Printf("%s %d %s %s", c.Request().Method, c.Response().Status(), path, stop.Sub(start))
}
```

Native Handler Support
```go
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-playground/lars"
)

func main() {

	l := lars.New()
	l.Use(Logger)

	l.Get("/", HelloWorld)

	http.ListenAndServe(":3007", l.Serve())
}

// HelloWorld ...
func HelloWorld(w http.ResponseWriter, r *http.Request) {

	// lars's context! get it and ROCK ON!
	ctx := lars.GetContext(w)

	ctx.Response().Write([]byte("Hello World"))
}

// Logger ...
func Logger(c lars.Context) {

	start := time.Now()

	c.Next()

	stop := time.Now()
	path := c.Request().URL.Path

	if path == "" {
		path = "/"
	}

	log.Printf("%s %d %s %s", c.Request().Method, c.Response().Status(), path, stop.Sub(start))
}
```

Middleware
-----------
There are some pre-defined middlewares within the middleware folder; NOTE: that the middleware inside will
comply with the following rule(s):

* Are completely reusable by the community without modification

Other middleware will be listed under the examples/middleware/... folder for a quick copy/paste modify. as an example a logging or
recovery middleware are very application dependent and therefore will be listed under the examples/middleware/...

Benchmarks
-----------
Run on MacBook Pro (Retina, 15-inch, Late 2013) 2.6 GHz Intel Core i7 16 GB 1600 MHz DDR3 using Go version go1.6 darwin/amd64

NOTE: you may have noticed the benchmark get a tiny bit slower since turning Context into an interface, but in the real world when
using your own Context ( even if only for passing around globals ), there is a single pool that your objects are stored in so the
small hit now will save you on the flip side in real world usage.

```go
go test -bench=. -benchmem=true

   githubAPI: 52600 Bytes
   gplusAPI: 3624 Bytes
   parseAPI: 6616 Bytes
   staticAPI: 30104 Bytes

PASS
BenchmarkLARS_GithubStatic-8	20000000	        96.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GithubParam-8 	10000000	       160 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GithubAll-8   	   50000	     34821 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GPlusStatic-8 	20000000	        73.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GPlusParam-8  	20000000	       101 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GPlus2Params-8	10000000	       127 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GPlusAll-8    	 1000000	      1705 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_ParseStatic-8 	20000000	        75.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_ParseParam-8  	20000000	        81.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_Parse2Params-8	20000000	        98.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_ParseAll-8    	  500000	      3290 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_StaticAll-8   	  100000	     22990 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_Param-8       	20000000	        81.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_Param5-8      	10000000	       138 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_Param20-8     	 5000000	       312 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_ParamWrite-8  	20000000	        85.0 ns/op	       0 B/op	       0 allocs/op
```

Package Versioning
----------
I'm jumping on the vendoring bandwagon, you should vendor this package as I will not
be creating different version with gopkg.in like allot of my other libraries.

Why? because my time is spread pretty thin maintaining all of the libraries I have + LIFE,
it is so freeing not to worry about it and will help me keep pouring out bigger and better
things for you the community.

This package is inspired by the following 
-----------
- [httptreemux](https://github.com/dimfeld/httptreemux)
- [httprouter](https://github.com/julienschmidt/httprouter)
- [echo](https://github.com/labstack/echo)
- [gin](https://github.com/gin-gonic/gin)

License 
--------
This project is licensed unter MIT, for more information look into the LICENSE file.
Copyright (c) 2016 Go Playground


