package Api

import (
    "fmt"
    "github.com/gorilla/mux"
    "github.com/inSituo/apiServer/DBAccess"
    "github.com/inSituo/apiServer/LeveledLogger"
    "net/http"
)

type Answers struct {
    ApiSection
}

func InitAnswersApi(
    r *mux.Router,
    db *DBAccess.DB,
    log *LeveledLogger.Logger,
    prefix string,
) {
    log.Debugf("Setting up answers API")

    a := Answers{
        ApiSection{
            db:     db,
            log:    log,
            r:      r,
            prefix: prefix,
            base:   "answers",
        },
    }

    a.setupRoute("GET", "get/{id}", a.getById)
    a.setupRoute("GET", "delete/{id}", a.deleteById)
    a.setupRoute("POST", "edit/{id}", a.editById)
}

func (a *Answers) getById(res http.ResponseWriter, req *http.Request) {
    params := mux.Vars(req)
    fmt.Fprintf(res, "get answer %s", params["id"])
}

func (a *Answers) deleteById(res http.ResponseWriter, req *http.Request) {
    params := mux.Vars(req)
    fmt.Fprintf(res, "delete answer %s", params["id"])
}

func (a *Answers) editById(res http.ResponseWriter, req *http.Request) {
    params := mux.Vars(req)
    fmt.Fprintf(res, "edit answer %s", params["id"])
}
