<div align="center">
    <a href="https://github.com/inphinit/teeny/">
        <img src="./badges/php.png" width="160" alt="Teeny route system for PHP">
    </a>
    <a href="https://github.com/inphinit/teeny.js/">
    <img src="./badges/javascript.png" width="160" alt="Teeny route system for JavaScript (Node.js)">
    </a>
    <a href="https://github.com/inphinit/teeny.go/">
    <img src="./badges/golang.png" width="160" alt="Teeny route system for Golang">
    </a>
    <a href="https://github.com/inphinit/teeny.py/">
    <img src="./badges/python.png" width="160" alt="Teeny route system for Python">
    </a>
</div>

## About Teeny.go

The main objective of this project is to be light, simple, easy to learn, to serve other projects that need a route system to use together with other libraries and mainly to explore the native resources from language and engine (Go).

## Configure your project

For install use:

```
go get -u github.com/inphinit/teeny.go
```

Create a file any name with extesion `.go`, example: `test.go`

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

For a simples test execution (`test.go` is a example):

```
go run test.go
```

For build execute:

```
go build
```

## Using TLS:

For use TLS in local server with certificate files

``` golang
package main

import (
    "fmt"
    "net/http"
    "github.com/inphinit/teeny"
)

func main() {
    app := teeny.Serve("localhost", 7000)

    app.SetTLS(true)

    app.SetCertificate("/home/foo/cert.pem")

    app.SetKey("/home/foo/key.pem")

    ...

    app.Exec()
}
```

## Using Fast-CGI:

For use teeny with Apache or Ngnix, enable Fast-CGI: 

``` golang
package main

import (
    "fmt"
    "net/http"
    "github.com/inphinit/teeny"
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

## Static files

Set absolute path

``` golang
func main() {
    app := teeny.Serve("localhost", 7000)

    app.SetPublic("/home/foo/bar")

    ...
```

## Methods for config teeny

Method | Description
--- | ------
`app := teeny.Serve(host string, port int)` | configure routes host and port
`app.SetDebug(enable bool)` | Define if debug is on (`true`) or off (`false`), by default is `false`
`app.SetFcgi(enable bool)` | Enable Fast-CGI
`app.SetTLS(enable bool)` | Enable TLS for server (not Fast-CGI)
`app.SetCertificate(certFile string)` | Set certificate, use with `app.SetTLS(true)`
`app.SetKey(keyFile string)` | Set certificate, use with `app.SetTLS(true)`
`app.SetPublic(path string)` | Define **absolute** path for use static files
`app.Action(method string, path string, func TeenyCallback)` | Define a route (from HTTP path in URL) for execute a function, arrow function or anonymous function
`app.Params(method string, path string, func TeenyPatternCallback)`,
`app.HandlerCodes(codes []int, func TeenyStatusCallback)` | Catch http errors (like `ErrorDocument` or `error_page`) from ISPAI or if try access a route not defined (emits `404 Not Found`) or if try access a defined route with not defined http method (emits `405 Method Not Allowed`)
`app.SetPattern(name string, regex string)` | Create a pattern for use in route params
`app.Exec()` | Starts server
`app.CliMode()` | Starts server with support for configure with commands

## Patterns supported by param routes (`app.Params()`)

You can create your own patterns to use with the routes in "teeny", but there are also ready-to-use patterns:

Pattern | Regex used | Description
--- | --- | ---
`alnum` | `[\da-zA-Z]+` | Matches routes with param using alpha-numeric in route
`alpha` | `[a-zA-Z]+` | Matches routes with param using A to Z letters in route
`decimal` | `\d+\.\d+` | Matches routes with param using decimal format (like `1.2`, `3.5`, `100.50`) in route
`num` | `\d+` | Matches routes with param using numeric format in route
`noslash` | `[^\/]+` | Matches routes with param using any character except slashs (`\/` or `/`) in route
`nospace` | `\S+` | Matches routes with param using any character except spaces, tabs or NUL in route
`uuid` | `[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}` | Matches routes with param using uuid format in route
`version` | `\d+\.\d+(\.\d+(-[\da-zA-Z]+(\.[\da-zA-Z]+)*(\+[\da-zA-Z]+(\.[\da-zA-Z]+)*)?)?)?` | Matches routes with param using [`semver.org`](https://semver.org/) format in route

## CLI mode

For create a application with support for command line arguments you can create manualy using `os.Args` or using `app.CliMode()`, example:

``` golang
func main() {
    //Default host and port
    app := teeny.Serve("localhost", 7000)

    app.Action("GET", "/", func (response http.ResponseWriter, request *http.Request) {
        fmt.Fprint(response, "Homepage")
    })

    app.Action("GET", "/about", func (response http.ResponseWriter, request *http.Request) {
        fmt.Fprint(response, "About page")
    })

    app.HandlerCodes([]int {403, 404, 405}, func (response http.ResponseWriter, request *http.Request, code int) {
        fmt.Fprintf(response, "Error: %d", code)
    })

    app.CliMode()
}
```

Usage example:

```
program.exe --debug --host 0.0.0.0 --port 8080 --public "/home/foo/bar/assets"
```

### CLI mode arguments

Argument | Example | Description
--- | --- | ---
`--tls` | `program --tls` | Enable TLS mode in your program (use with `--cert` and `--key` if it is not pre-configured)
`--tls` | `program --no-tls` | Disable TLS (if in the initial configuration of the script it was enabled)
`--debug` | `program --debug` | Enable debug mode in your program
`--debug` | `program --no-debug` | Disable debug (if in the initial configuration of the script it was enabled)
`--fcgi` | `program --fcgi` | Enable Fast-CGI mode in your program
`--fcgi` | `program --no-fcgi` | Disable Fast-CGI (if in the initial configuration of the script it was enabled)
`--cert` | `program --cert /home/foo/cert.pem` | Define certificate file (use with `--tls` if it is not pre-configured)
`--key` | `program --key /home/foo/key.pem` | Define key file (use with `--tls` if it is not pre-configured)
`--public` | `program --public /home/foo/assets` | Define folder for access static files
`--host` | `program --host 0.0.0.0` | Define host address
`--port` | `program --port 9000` | Define port addres
