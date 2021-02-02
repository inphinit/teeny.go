package teeny

import (
    "fmt"
    "net"
    "net/http"
    "net/http/fcgi"
    "regexp"
    "strings"
    "time"
)

type TeenyCallback func(http.ResponseWriter, *http.Request)
type TeenyPatternCallback func(http.ResponseWriter, *http.Request, map[string]string)
type TeenyStatusCallback func(http.ResponseWriter, *http.Request, int)

type TeenyServe struct {
    debug       bool
    host        string
    port        int
    fcgi        bool
    tls         bool
    certFile    string
    keyFile     string
    publicPath  string
    routes      map[string]map[string]TeenyCallback
    pRoutes     map[string]map[string]TeenyPatternCallback
    codes       map[int]TeenyStatusCallback
    patterns    map[string]string
}

func Serve(host string, port int) TeenyServe {

    return TeenyServe {
        false,
        host,
        port,
        false,
        false,
        "",
        "",
        "",
        make(map[string]map[string]TeenyCallback),
        make(map[string]map[string]TeenyPatternCallback),
        make(map[int]TeenyStatusCallback),
        map[string]string {
            "alnum": `[\da-zA-Z]+`,
            "alpha": `[a-zA-Z]+`,
            "decimal": `\d+\.\d+`,
            "num": `\d+`,
            "noslash": `[^\/]+`,
            "nospace": `\S+`,
            "uuid": `[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`,
            "version": `\d+\.\d+(\.\d+(-[\da-zA-Z]+(\.[\da-zA-Z]+)*(\+[\da-zA-Z]+(\.[\da-zA-Z]+)*)?)?)?`,
        },
    }
}

func (e *TeenyServe) SetDebug(enable bool) {
    e.debug = enable
}

func (e *TeenyServe) SetFcgi(enable bool) {
    e.fcgi = enable
}

func (e *TeenyServe) SetTLS(enable bool) {
    e.tls = enable
}

func (e *TeenyServe) SetCertificate(certFile string, keyFile string) {
    e.certFile = certFile
    e.keyFile = keyFile
}

func (e *TeenyServe) SetPublic(path string) {
    e.publicPath = path
}

func (e *TeenyServe) Action(method string, path string, callback TeenyCallback) {

    if e.routes == nil {
        fmt.Print("Create routes...\n")

        e.routes = make(map[string]map[string]TeenyCallback)
    }

    if _, ok := e.routes[path]; !ok {
        e.routes[path] = make(map[string]TeenyCallback)
    }

    e.routes[path][method] = callback
}

func (e *TeenyServe) Pattern(method string, path string, callback TeenyPatternCallback) {

    if e.routes == nil {
        e.pRoutes = make(map[string]map[string]TeenyPatternCallback)
    }

    if _, ok := e.pRoutes[path]; !ok {
        e.pRoutes[path] = make(map[string]TeenyPatternCallback)
    }

    e.pRoutes[path][method] = callback
}

func (e *TeenyServe) HandlerCodes(codes []int, callback TeenyStatusCallback) {

    for _, code := range codes {
        e.codes[code] = callback
    }
}

func (e *TeenyServe) params(
    response http.ResponseWriter,
    request *http.Request,
    method string,
    path string,
) bool {

    re := regexp.MustCompile(`[<](.*?)(\:(` + "x" + `)|)[>]`)

    for path, methods := range e.pRoutes { 
        if strings.Index(path, "<") != -1 {
            match := re.FindStringSubmatch(path)

            if len(match) > 0 {
                var params = make(map[string]string)

                for index, name := range re.SubexpNames() {
                    if index > 0 {
                        params[name] = match[index]

                        fmt.Printf("%v\n", methods)
                    }
                }



                continue
            }
        }
    }

    return false
}

func (e *TeenyServe) Exec() {

    var host = fmt.Sprintf("%s:%d", e.host, e.port)

    http.HandleFunc("/", func (response http.ResponseWriter, request *http.Request) {
        e.handler(response, request)
    })

    if e.debug {
        fmt.Printf("Listing %s ...\n", host)
    }

    if e.fcgi {
        serve, err := net.Listen("tcp", host)

        if err != nil {
            e.panic(err)
        }

        fcgi.Serve(serve, nil)
    } else if e.tls {
        e.panic(http.ListenAndServeTLS(host, e.certFile, e.keyFile, nil))
    } else {
        e.panic(http.ListenAndServe(host, nil))
    }
}

func (e *TeenyServe) panic(err error) {

    if err != nil {
        panic(err)
    }
}

func (e *TeenyServe) handler(response http.ResponseWriter, request *http.Request) {

    var path = request.URL.Path
    var method = request.Method
    var code = 200

    // fmt.Print("\n")
    // fmt.Printf("handler() -> %s %s %s\n", method, path, request.Proto)
    // fmt.Printf("handler() -> e.routes[%s]: %v\n", path, e.routes[path])
    // fmt.Printf("handler() -> e.routes[%s][%s]: %v\n", path, method, e.routes[path][method])

    if e.debug {
        fmt.Printf("[%s] %s %s %s\n", time.Now().Format(time.RFC1123), method, path, request.Proto)
    }

    if methods, ok := e.routes[path]; ok {
        if callback, ok := methods[method]; ok {
            callback(response, request)
        } else {
            response.WriteHeader(http.StatusMethodNotAllowed)
            code = 405
        }
    } else {
        response.WriteHeader(http.StatusNotFound)
        code = 404
    }

    if code != 200 {
        if callback, ok := e.codes[code]; ok {
            callback(response, request, code)
        } else {
            fmt.Fprint(response, "")
        }
    }
}
