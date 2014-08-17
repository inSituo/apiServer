package Api

import (
    "fmt"
    "github.com/gorilla/mux"
    "github.com/inSituo/apiServer/DBAccess"
    "github.com/inSituo/apiServer/LeveledLogger"
    "github.com/inSituo/apiServer/Middleware"
    "gopkg.in/mgo.v2/bson"
    "net/http"
)

type AnswersApi struct {
    ApiSection
}

func InitAnswersApi(
    r *mux.Router,
    db *DBAccess.DB,
    log *LeveledLogger.Logger,
) {
    log.Debugf("Setting up answers API")

    a := AnswersApi{
        ApiSection{
            db:  db,
            log: log,
            r:   r,
        },
    }

    a.use(Middleware.IdVerifier)

    a.setupRoute("GET", "/get/{id}", a.getById)
    a.setupRoute("GET", "/delete/{id}", a.deleteById)
    a.setupRoute("POST", "/edit/{id}", a.editById)
}

type Answer struct {
    id  bson.ObjectId   `bson:"_id"`
    qid bson.ObjectId   `bson:"qid"`
    uid bson.ObjectId   `bson:"uid"`
    ts  int             `bson:"ts"`
    rev ContentRevision `bson:"rev"`
}

func (a *AnswersApi) getById(res http.ResponseWriter, req *http.Request) {
    params := mux.Vars(req)
    // id :=
    fmt.Fprintf(res, "get answer %s", params["id"])
}

func (a *AnswersApi) deleteById(res http.ResponseWriter, req *http.Request) {
    params := mux.Vars(req)
    fmt.Fprintf(res, "delete answer %s", params["id"])
}

func (a *AnswersApi) editById(res http.ResponseWriter, req *http.Request) {
    params := mux.Vars(req)
    fmt.Fprintf(res, "edit answer %s", params["id"])
}
