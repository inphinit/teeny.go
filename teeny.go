package teeny

import (
    "fmt"
    "net"
    "net/http"
    "net/http/fcgi"
    "os"
    "regexp"
    "strconv"
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
    hasParams   bool
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
            "decimal": `(\d|[1-9]\d+)\.\d+`,
            "nospace": `[^/\s]+`,
            "num": `\d+`,
            "uuid": `[\da-fA-F]{8}-[\da-fA-F]{4}-[\da-fA-F]{4}-[\da-fA-F]{4}-[\da-fA-F]{12}`,
            "version": `\d+\.\d+(\.\d+(-[\da-zA-Z]+(\.[\da-zA-Z]+)*(\+[\da-zA-Z]+(\.[\da-zA-Z]+)*)?)?)?`,
        },
        false,
        nil,
        regexp.MustCompile(`\<\>\)`),
        regexp.MustCompile(`/`),
    }
}

func (e *TeenyServe) SetDebug(enable bool) {

    e.debug = enable
}

func (e *TeenyServe) SetHost(host string) {

    e.host = host
}

func (e *TeenyServe) SetPort(port int) {

    e.port = port
}

func (e *TeenyServe) SetFcgi(enable bool) {

    e.fcgi = enable
}

func (e *TeenyServe) SetTLS(enable bool) {

    e.tls = enable
}

func (e *TeenyServe) SetCertificate(certFile string) {

    e.certFile = certFile
}

func (e *TeenyServe) SetKey(keyFile string) {

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

    e.patternRE = regexp.MustCompile(`[<]([A-Za-z]\w+)(\:(` + strings.Join(patternKeys, "|") + `)|)[>]`)
}

func (e *TeenyServe) Action(method string, path string, callback TeenyCallback) {

    if _, ok := e.routes[path]; !ok {
        e.routes[path] = make(map[string]TeenyCallback)
    }

    e.routes[path][method] = callback
}

func (e *TeenyServe) Params(method string, path string, callback TeenyPatternCallback) {

    path = regexp.QuoteMeta(path)

    if strings.Index(path, "<") == -1 {
        panic(fmt.Sprintf("Invalid parameterized route %v", path))
    }

    if _, ok := e.pRoutes[path]; !ok {
        e.pRoutes[path] = make(map[string]TeenyPatternCallback)
    }

    if callback != nil  {
        e.hasParams = true;
    }

    e.pRoutes[path][method] = callback
}

func (e *TeenyServe) HandlerCodes(codes []int, callback TeenyStatusCallback) {

    for _, code := range codes {
        e.codes[code] = callback
    }
}

func (e *TeenyServe) Exec() {

    var address = fmt.Sprintf("%s:%d", e.host, e.port)

    http.HandleFunc("/", func (response http.ResponseWriter, request *http.Request) {
        e.handler(response, request)
    })

    if e.debug {
        fmt.Printf("Listing %s ...\n", address)
    }

    if e.fcgi {
        serve, err := net.Listen("tcp", address)

        e.panic(err)

        fcgi.Serve(serve, nil)
    } else if e.tls {
        e.panic(http.ListenAndServeTLS(address, e.certFile, e.keyFile, nil))
    } else {
        e.panic(http.ListenAndServe(address, nil))
    }
}

func (e *TeenyServe) CliMode() {

    var ignoreNext = false

    for index, arg := range os.Args {
        if ignoreNext || index == 0 {
            ignoreNext = false
            continue
        }

        switch arg {
        case "--tls":
            e.SetTLS(true)

        case "--no-tls":
            e.SetTLS(false)

        case "--debug":
            e.SetDebug(true)

        case "--no-debug":
            e.SetDebug(false)

        case "--fcgi":
            e.SetFcgi(true)

        case "--no-fcgi":
            e.SetFcgi(false)

        case "--cert":
            ignoreNext = true
            e.SetCertificate(os.Args[index + 1])

        case "--key":
            ignoreNext = true
            e.SetKey(os.Args[index + 1])

        case "--public":
            ignoreNext = true
            e.SetPublic(os.Args[index + 1])

        case "--host":
            ignoreNext = true
            e.SetHost(os.Args[index + 1])

        case "--port":
            ignoreNext = true

            port, err := strconv.Atoi(os.Args[index + 1])
            if err != nil {
                panic(err)
            }

            e.SetPort(port)

        default:
            panic(fmt.Sprintf("Invalid argument %v", arg))

        }
    }

    if ignoreNext {
        panic("Missing an argument")
    }

    e.Exec()
}

func (e *TeenyServe) handler(response http.ResponseWriter, request *http.Request) {

    var path = request.URL.Path
    var method = request.Method
    var code = http.StatusOK

    if e.debug {
        fmt.Printf("[%s] %s %s %s\n", time.Now().Format(time.RFC1123), method, path, request.Proto)
    }

    if e.publicPath != "" {
        code = e.public(response, request, path)

        if code == 0 {
            return
        }
    }

    if methods, ok := e.routes[path]; ok {
        if callback, ok := methods[method]; ok {
            callback(response, request)
        } else if callback, ok := methods["ANY"]; ok {
            callback(response, request)
        } else {
            code = http.StatusMethodNotAllowed
        }
    } else if e.hasParams {
        code = e.findParams(response, request, method, path)
    } else {
        code = http.StatusNotFound
    }

    if code != http.StatusOK {
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

func (e *TeenyServe) findParams(
    response http.ResponseWriter,
    request *http.Request,
    method string,
    pathinfo string,
) int {

    for path, methods := range e.pRoutes {
        path = e.scapesRE.ReplaceAllString(path, `\/`)
        path = e.patternRE.ReplaceAllString(path, "(?P<$1><$3>)")
        path = e.signsRE.ReplaceAllString(path, "[^/]+)")

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

            if callback, ok := methods[method]; ok {
                callback(response, request, params)

                return http.StatusOK
            } else if callback, ok := methods["ANY"]; ok {
                callback(response, request, params)

                return http.StatusOK
            } else {
                return http.StatusMethodNotAllowed
            }
        }
    }

    return http.StatusNotFound
}

func (e *TeenyServe) panic(err error) {

    if err != nil {
        panic(err)
    }
}
