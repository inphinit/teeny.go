## About Teeny.go

The main objective of this project is to be light, simple, easy to learn, to serve other projects that need a route system to use together with other libraries and mainly to explore the native resources from language and engine (Go).

## Configure your project

Create a file any name with extesion `.go`

For use local server use like this:

``` golang
package main

import (
    "fmt"
    "net/http"
    "github.com/inphinit/teeny.go"
)

func main() {
    app := teeny.Serve("localhost", 7000)

    app.SetPublic("/home/user/Documents/")

    app.Action("GET", "/", func (response http.ResponseWriter, request *http.Request) {
        fmt.Fprint(response, "Homepage")
    })

    app.Action("GET", "/about", func (response http.ResponseWriter, request *http.Request) {
        fmt.Fprint(response, "About page")
    })

    app.Exec()
}
```

Using TLS:

``` golang
package main

import (
    "fmt"
    "net/http"
    "github.com/inphinit/teeny.go"
)

func main() {
    app := teeny.Serve("localhost", 7000)

    app.SetTLS(true)

    app.SetCertificate("/home/foo/cert.pem", "/home/foo/key.pem")

    ...

    app.Exec()
}
```

Using Fast-CGI:

``` golang
package main

import (
    "fmt"
    "net/http"
    "github.com/inphinit/teeny.go"
)

func main() {
    app := teeny.Serve("localhost", 7000)

    app.SetFcgi(true)

    ...

    app.Exec()
}
```

## Handling errors

``` golang
func main() {
    app := teeny.Serve("localhost", 7000)

    ...
    
    var codes = []int {403, 404, 405, 500}

    app.HandlerCodes(codes, func (response http.ResponseWriter, request *http.Request, code int) {
        fmt.Fprintf(response, "Error %d", code)
    })

    app.Exec()
}
```

Different handlers:

``` golang
func main() {
    app := teeny.Serve("localhost", 7000)

    ...

    app.HandlerCodes([]int {404, 405}, func (response http.ResponseWriter, request *http.Request, code int) {
        fmt.Fprintf(response, "Router error: %d", code)
    })

    app.HandlerCodes([]int {403, 500}, func (response http.ResponseWriter, request *http.Request, code int) {
        fmt.Fprintf(response, "Error from ISAPI: %d", code)
    })

    app.Exec()
}
```

## Methods for config Teeny.go

Method | Description
--- | ------
`app := teeny.Serve(host string, port int)` | configure routes host and port
`app.SetDebug(enable bool)` | Define if debug is on (`true`) or off (`false`), by default is `false`
`app.SetFcgi(enable bool)` | Enable Fast-CGI
`app.SetTLS(enable bool)` | Enable TLS for server (not Fast-CGI)
`app.SetCertificate(certFile string, keyFile string)` | Set certificate, use with `app.SetTLS(true)`
`app.SetPublic(path string)` | Define path for use static files
`app.Action(method string, path string, func TeenyCallback)` | Define a route (from HTTP path in URL) for execute a function, arrow function or anonymous function
`app.Pattern(method string, path string, func TeenyPatternCallback)`,
`app.HandlerCodes(codes []int, func TeenyStatusCallback)` | Catch http errors (like `ErrorDocument` or `error_page`) from ISPAI or if try access a route not defined (emits `404 Not Found`) or if try access a defined route with not defined http method (emits `405 Method Not Allowed`)
`app.SetPattern(name string, regex string)` | Create a pattern for use in route params
`app.Exec()` | Starts server

## Patterns supported by param routes

You can create your own patterns to use with the routes in "Teeny.go", but there are also ready-to-use patterns:

Pattern | Regex used | Description
--- | --- | ---
`alnum` | `[\\da-zA-Z]+` | Matches routes with param using alpha-numeric in route
`alpha` | `[a-zA-Z]+` | Matches routes with param using A to Z letters in route
`decimal` | `\\d+\\.\\d+` | Matches routes with param using decimal format (like `1.2`, `3.5`, `100.50`) in route
`num` | `\\d+` | Matches routes with param using numeric format in route
`noslash` | `[^\\/]+` | Matches routes with param using any character except slashs (`\/` or `/`) in route
`nospace` | `\\S+` | Matches routes with param using any character except spaces, tabs or NUL in route
`uuid` | `[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}` | Matches routes with param using uuid format in route
`version` | `\\d+\\.\\d+(\\.\\d+(-[\\da-zA-Z]+(\\.[\\da-zA-Z]+)*(\\+[\\da-zA-Z]+(\\.[\\da-zA-Z]+)*)?)?)?` | Matches routes with param using [`semver.org`](https://semver.org/) format in route
