package teenygo

import (
    "fmt"
    "net"
    "net/http"
    "net/http/fcgi"
    "os"
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
    patternRE   *regexp.Regexp
    signsRE     *regexp.Regexp
    scapesRE    *regexp.Regexp
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
        nil,
        regexp.MustCompile(`\<\>\)`),
        regexp.MustCompile(`/`),
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

func (e *TeenyServe) SetPattern(pattern string, regex string) {

    e.patterns[pattern] = regex

    var patternKeys []string

    for key, _ := range e.patterns {
        patternKeys = append(patternKeys, key)
    }

    e.patternRE = regexp.MustCompile(`[<](.*?)(\:(` + strings.Join(patternKeys, "|") + `)|)[>]`)
}

func (e *TeenyServe) Action(method string, path string, callback TeenyCallback) {

    if e.routes == nil {
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
    pathinfo string,
) bool {

    for path, methods := range e.pRoutes {
        if strings.Index(path, "<") != -1 {
            callback := methods[method]

            if callback == nil {
                continue
            }

            path = e.scapesRE.ReplaceAllString(path, `\/`)
            path = e.patternRE.ReplaceAllString(path, "(?P<$1><$3>)")
            path = e.signsRE.ReplaceAllString(path, ".*?)")

            for pattern, replace := range e.patterns {
                path = regexp.MustCompile("<" + pattern +  ">").ReplaceAllString(path, replace)
            }

            re := regexp.MustCompile("^" + path +  "$")
            match := re.FindStringSubmatch(pathinfo)

            if len(match) != 0 {
                var params = make(map[string]string)

                for index, name := range re.SubexpNames() {
                    if index > 0 {
                        params[name] = match[index]
                    }
                }

                callback(response, request, params)

                return true
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

        e.panic(err)

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

    if e.debug {
        fmt.Printf("[%s] %s %s %s\n", time.Now().Format(time.RFC1123), method, path, request.Proto)
    }

    if e.publicPath != "" {
        code = e.public(response, request, path)

        if code == 0 {
            return
        }
    } else {
        code = 200
    }

    if methods, ok := e.routes[path]; ok {
        if callback, ok := methods[method]; ok {
            callback(response, request)
        } else {
            code = http.StatusMethodNotAllowed
        }
    } else if !e.params(response, request, method, path) {
        code = http.StatusNotFound
    }

    if code != 200 {
        response.WriteHeader(code)

        if callback, ok := e.codes[code]; ok {
            callback(response, request, code)
        }
    }
}

func (e *TeenyServe) public(response http.ResponseWriter, request *http.Request, path string) int {

    var fullpath = e.publicPath + path

    fi, err := os.Lstat(fullpath)

    if err != nil {
        if os.IsPermission(err) {
            return 403
        } else if os.IsNotExist(err) {
            return 200
        }

        return 500
    } else if fi.Mode().IsDir() {
        return 200
    }

    http.ServeFile(response, request, fullpath)

    return 0
}
