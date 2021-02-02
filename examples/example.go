package main

import (
    "fmt"
    "net/http"
    // "github.com/inphinit/teeny.go"
    "../"
)

func main() {
    // re := regexp.MustCompile(`a(x*)b(y|z)c`)
    // fmt.Printf("%q\n", re.FindStringSubmatch("-axxxbyc-"))
    // fmt.Printf("%q\n", re.FindStringSubmatch("-abzc-"))
    // fmt.Printf("%q\n\n", re.FindStringSubmatch("aaaa"))

    app := teeny.Serve("localhost", 7000)

    app.SetDebug(true)

    app.SetPublic("/home/user/Documents/")

    app.Action("GET", "/", func (response http.ResponseWriter, request *http.Request) {
        fmt.Fprintf(response, "Homepage")
    })

    app.Action("GET", "/foo/bar/", func (response http.ResponseWriter, request *http.Request) {
        fmt.Fprintf(response, "Test for /foo/bar")
    })

    app.Action("GET", "/bigfile", func (response http.ResponseWriter, request *http.Request) {
        http.ServeFile(response, request, "./file.rar")
    })

    app.Pattern("GET", "/test/<id:alnum>", func (response http.ResponseWriter, request *http.Request, params map[string]string) {
        fmt.Fprint(response, "Params:\n")

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
