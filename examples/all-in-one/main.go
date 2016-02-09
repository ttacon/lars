package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/lars"
)

// This is a contrived example using globals as I would use it in production
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
func (g *ApplicationGlobals) Reset(c *lars.Context) {
	// DB = new database connection or reset....
	//
	// We don't touch translator + log as they don't change per request
}

// Done gets called after the HTTP request has completed right before
// Context gets put back into the pool
func (g *ApplicationGlobals) Done() {
	// DB.Close()
}

var _ lars.IAppContext = &ApplicationGlobals{} // ensures ApplicationGlobals complies with lasr.IGlobals at compile time

func main() {

	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	// translator := ...
	// db := ... base db connection or info
	// json := ...
	// schema := ...

	globalsFn := func() lars.IAppContext {
		return &ApplicationGlobals{
			Log: logger,
			// Translator: translator,
			// DB: db,
			// JSON: json,
			// schema:schema,
		}
	}

	l := lars.New()
	l.RegisterAppContext(globalsFn)
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
func Home(c *lars.Context) {

	app := c.AppContext.(*ApplicationGlobals)

	var username string

	// username = app.DB.find(user by .....)

	app.Log.Println("Found User")

	c.Response.Write([]byte("Welcome Home " + username))
}

// Users ...
func Users(c *lars.Context) {

	app := c.AppContext.(*ApplicationGlobals)

	app.Log.Println("In Users Function")

	c.Response.Write([]byte("Users"))
}

// User ...
func User(c *lars.Context) {

	app := c.AppContext.(*ApplicationGlobals)

	id := c.Param("id")

	var username string

	// username = app.DB.find(user by id.....)

	app.Log.Println("Found User")

	c.Response.Write([]byte("Welcome " + username + " with id " + id))
}

// UserProfile ...
func UserProfile(c *lars.Context) {

	app := c.AppContext.(*ApplicationGlobals)

	id := c.Param("id")

	var profile string

	// profile = app.DB.find(user profile by .....)

	app.Log.Println("Found User Profile")

	c.Response.Write([]byte("Here's your profile " + profile + " user " + id))
}

// Logger ...
func Logger(c *lars.Context) {

	start := time.Now()

	c.Next()

	stop := time.Now()
	path := c.Request.URL.Path

	if path == "" {
		path = "/"
	}

	log.Printf("%s %d %s %s", c.Request.Method, c.Response.Status(), path, stop.Sub(start))
}
