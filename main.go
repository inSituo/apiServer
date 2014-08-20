package main

import (
    "flag"
    "fmt"
    "github.com/gorilla/context"
    "github.com/gorilla/mux"
    "github.com/inSituo/apiServer/Api"
    "github.com/inSituo/apiServer/DBAccess"
    "github.com/inSituo/apiServer/LeveledLogger"
    "github.com/inSituo/apiServer/Middleware"
    "net/http"
    "os"
    "runtime"
)

const (
    SERVER_VER_MAJOR = 0
    SERVER_VER_MINOR = 0
    SERVER_VER_PATCH = 0
)

type GoogleConf struct {
    clientId     *string
    clientSecret *string
    redirectUrl  *string
    scope        *string
}

type ServerConf struct {
    mongo      DBAccess.MongoConf
    google     GoogleConf
    port       *int
    debug      *bool
    serverName *string
    serverVer  *string
}

func main() {
    conf := ServerConf{
        debug:      flag.Bool("debug", false, "Enable debug log messages"),
        port:       flag.Int("port", 80, "Server listening port"),
        serverName: flag.String("servername", "InSituo API Server", "Server name for HTTP header"),
        serverVer: flag.String(
            "serverver",
            fmt.Sprintf("v%d.%d.%d",
                SERVER_VER_MAJOR,
                SERVER_VER_MINOR,
                SERVER_VER_PATCH),
            "Server version for HTTP header",
        ),
        mongo: DBAccess.MongoConf{
            Port:       flag.Int("mport", 27017, "MongoDB server port"),
            Host:       flag.String("mhost", "127.0.0.1", "MongoDB server host"),
            DB:         flag.String("mdb", "insituo-dev", "MongoDB database name"),
            CUsers:     flag.String("cusers", "users", "Name of users collection in DB"),
            CLogins:    flag.String("clogins", "logins", "Name of logins collection in DB"),
            CQuestions: flag.String("cquestions", "questions", "Name of questions collection in DB"),
            CAnswers:   flag.String("canswers", "answers", "Name of answers collection in DB"),
        },
        google: GoogleConf{
            clientId:     flag.String("gclientid", "", "Google API client ID"),
            clientSecret: flag.String("gclientsecret", "", "Google API client secret"),
            redirectUrl:  flag.String("gredirecturl", "", "Google API redirect URL"),
            scope:        flag.String("gscope", "", "Google API scope"),
        },
    }
    flag.StringVar(&Api.API_KEY_REQ_HEADER, "apikeyheader", "X-API-KEY", "API key request-header name")
    flag.StringVar(&Api.API_FMT_REQ_HEADER, "apifmtheader", "X-API-FORMAT", "API response format request-header name")

    showHelp := flag.Bool("help", false, "Show help")

    flag.Parse()
    if *showHelp {
        fmt.Printf("%s/%s\n\n", *conf.serverName, *conf.serverVer)
        flag.PrintDefaults()
        return
    }

    ll_level := LeveledLogger.LL_INFO
    if *conf.debug {
        ll_level = LeveledLogger.LL_DEBUG
    }
    log := LeveledLogger.New(os.Stdout, ll_level)

    log.Debugf("Debug mode enabled")

    log.Infof(
        "Connecting to MongoDB at %s:%d/%s",
        *conf.mongo.Host,
        *conf.mongo.Port,
        *conf.mongo.DB,
    )
    db, err := DBAccess.New(conf.mongo)
    if err != nil {
        log.Errorf("Unable to connect to MongoDB: %s", err) // this will panic
    }
    defer db.Close()

    // breakable middleware chain:
    chain := Middleware.NewChain(true)

    serverHttpHeader := fmt.Sprintf(
        "%s/%s (%s)",
        *conf.serverName,
        *conf.serverVer,
        runtime.GOOS,
    )
    log.Debugf("'Server' HTTP header set to '%s'", serverHttpHeader)
    chain.Push(func(res http.ResponseWriter, req *http.Request) {
        res.Header().Set(
            "Server",
            serverHttpHeader,
        )
    })
    if *conf.debug {
        chain.Push(Middleware.RequestDebugInfo(log))
    }

    r := mux.NewRouter()
    r.KeepContext = true

    // init apis:
    Api.InitAnswersApi(r.PathPrefix("/answers/").Subrouter(), db, log, Middleware.SetResponse)

    chain.PushHandler(r)

    // unbrakeable middleware:
    ubchain := Middleware.NewChain(false)
    ubchain.Push(chain.MakeHandler())
    ubchain.Push(Middleware.AutoFormatResponder(log, Api.API_FMT_REQ_HEADER))
    ubchain.Push(func(res http.ResponseWriter, req *http.Request) {
        context.Clear(req)
        log.Debugf("Request context cleared")
    })

    http.Handle("/", ubchain.MakeHandler())
    addr := fmt.Sprintf(":%d", *conf.port)
    log.Infof("Starting web service at %s", addr)
    if err := http.ListenAndServe(addr, nil); err != nil {
        log.Errorf("%s", err)
    }
}
