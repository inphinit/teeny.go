package main

import (
    "fmt"
    "net/http"
    "github.com/inphinit/teeny.go"
)

func main() {

    app := teenygo.Serve("localhost", 7000)

    app.SetDebug(true)

    // Asset static files
    app.SetPublic("/foo/bar/baz/public")

    app.Action("GET", "/", func (response http.ResponseWriter, request *http.Request) {
        fmt.Fprintf(response, "Homepage")
    })

    app.Action("GET", "/about", func (response http.ResponseWriter, request *http.Request) {
        fmt.Fprintf(response, "About")
    })

    app.Action("GET", "/foo/bar/", func (response http.ResponseWriter, request *http.Request) {
        fmt.Fprintf(response, "Test for /foo/bar/")
    })

    app.Action("GET", "/bigfile", func (response http.ResponseWriter, request *http.Request) {
        http.ServeFile(response, request, "./file.rar")
    })

    app.Params("GET", "/users/<id:alnum>", func (response http.ResponseWriter, request *http.Request, params map[string]string) {
        fmt.Fprint(response, "Params:\n")

        for key, value := range params {
            fmt.Fprintf(response, "%s = %s\n", key, value)
        }
    })

    app.Params("GET", "/users/<id:num>/<name:alnum>", func (response http.ResponseWriter, request *http.Request, params map[string]string) {
        fmt.Fprint(response, "Params:\n")

        for key, value := range params {
            fmt.Fprintf(response, "%s = %s\n", key, value)
        }
    })

    // Set custom pattern basead in Regex (write using string)
    app.SetPattern("example", "[A-Z]\\d+")

    // Using custom pattern for get param in route (access http://localhost:7000/custom/A1000)
    app.Params("GET", "/custom/<myexample:example>", func (response http.ResponseWriter, request *http.Request, params map[string]string) {
        fmt.Fprint(response, "Custom Param:\n")

        for key, value := range params {
            fmt.Fprintf(response, "%s = %s\n", key, value)
        }
    })

    var codes = []int {403, 404, 405, 500}

    app.HandlerCodes(codes, func (response http.ResponseWriter, request *http.Request, code int) {
        fmt.Fprintf(response, "Error %d", code)
    })

    app.HandlerCodes([]int {500, 501}, func (response http.ResponseWriter, request *http.Request, code int) {
        fmt.Fprintf(response, "FATAL ERROR: %d", code)
    })

    app.Exec()
}
