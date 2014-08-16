package main

import (
    "flag"
    "fmt"
    "github.com/gorilla/mux"
    "github.com/inSituo/apiServer/Api"
    "github.com/inSituo/apiServer/DBAccess"
    "github.com/inSituo/apiServer/LeveledLogger"
    "github.com/inSituo/apiServer/MiddlewareChain"
    "net/http"
    "os"
)

type GoogleConf struct {
    clientId     *string
    clientSecret *string
    redirectUrl  *string
    scope        *string
}

type ServerConf struct {
    mongo       DBAccess.MongoConf
    google      GoogleConf
    port        *int
    apiPrefix   *string
    debug       *bool
    sessionSec  *string
    sessionName *string
}

func main() {
    conf := ServerConf{
        debug:       flag.Bool("debug", false, "Enable debug log messages"),
        port:        flag.Int("port", 80, "Server listening port"),
        apiPrefix:   flag.String("prefix", "", "API endpoints prefix"),
        sessionSec:  flag.String("sessionsecret", "SeCrEt", "HTTP session secret"),
        sessionName: flag.String("sessionname", "inSituoSes", "HTTP cookie name"),
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

    showHelp := flag.Bool("help", false, "Show help")

    flag.Parse()
    if *showHelp {
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
        "Connecting to MongoDB at %s:%s/%s",
        *conf.mongo.Host,
        *conf.mongo.Port,
        *conf.mongo.DB,
    )
    db, err := DBAccess.New(conf.mongo)
    if err != nil {
        log.Errorf("Unable to connect to MongoDB: %s", err)
    }
    defer db.Close()

    r := mux.NewRouter()

    // init apis:
    Api.InitAnswersApi(r, db, log, *conf.apiPrefix)

    // register middleware:
    chain := MiddlewareChain.Chain{}
    chain.PushHandler(r)

    http.Handle("/", chain.MakeHandler())
    addr := fmt.Sprintf(":%d", *conf.port)
    log.Infof("Starting web service at %s", addr)
    if err := http.ListenAndServe(addr, nil); err != nil {
        log.Errorf("%s", err)
    }
}
