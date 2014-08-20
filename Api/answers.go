package Api

import (
    "encoding/json"
    "fmt"
    "github.com/gorilla/context"
    "github.com/gorilla/mux"
    "github.com/inSituo/apiServer/DBAccess"
    "github.com/inSituo/apiServer/LeveledLogger"
    "github.com/inSituo/apiServer/Middleware"
    "gopkg.in/mgo.v2"
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

    a.use(Middleware.IdVerifier(a.log))
    a.use(a.GetUserInfo)
    // all methods in this api require a logged-in user
    a.use(func(res http.ResponseWriter, req *http.Request) {
        if context.Get(req, "loggedIn") != true {
            a.respondNotLoggedIn(res)
            context.Set(req, "break-chain", true)
        }
    })

    a.setupRoute("GET", "/{id}", a.getById)
    a.setupRoute("GET", "/revs/{id}", a.revsById)
    a.setupRoute("DELETE", "/{id}", a.deleteById)
    a.setupRoute("POST", "/rev/{id}", a.newRevById)
    a.setupRoute("POST", "/{id}", a.newByQid)
}

type Answer struct {
    ID   bson.ObjectId   `bson:"_id" json:"id"`
    QID  bson.ObjectId   `bson:"qid" json:"qid"`
    OUID bson.ObjectId   `bson:"ouid" json:"ouid"`
    OTS  int             `bson:"ots" json:"ots"`
    Rev  ContentRevision `bson:"rev" json:"rev"`
}

type Revisions struct {
    ID   bson.ObjectId     `bson:"_id" json:"id"`
    QID  bson.ObjectId     `bson:"qid" json:"qid"`
    Revs []ContentRevision `bson:"revs" json:"revs"`
}

func (a *AnswersApi) getById(res http.ResponseWriter, req *http.Request) {
    id := bson.ObjectIdHex(mux.Vars(req)["id"])
    pipe := a.db.Answers.Pipe([]bson.M{
        {
            "$match": bson.M{
                "_id": id,
            },
        },
        {
            "$unwind": "$revs",
        },
        {
            "$sort": bson.M{
                "revs.ts": 1,
            },
        },
        {
            "$group": bson.M{
                "_id": bson.M{
                    "_id": "$_id",
                    "qid": "$qid",
                },
                "first_rev": bson.M{
                    "$first": "$revs",
                },
                "last_rev": bson.M{
                    "$last": "$revs",
                },
            },
        },
        {
            "$project": bson.M{
                "_id":  "$_id._id",
                "qid":  "$_id.qid",
                "ouid": "$first_rev.uid",
                "ots":  "$first_rev.ts",
                "rev":  "$last_rev",
            },
        },
    })
    var answer Answer
    if err := pipe.One(&answer); err != nil {
        if err != mgo.ErrNotFound {
            a.log.Warnf("In 'AnswersApi.getById', Pipe returned error for %s: %s", id, err)
            res.WriteHeader(http.StatusInternalServerError)
            return
        }
        a.log.Debugf("In 'AnswersApi.getById', Pipe returned empty for %s", id)
        res.WriteHeader(http.StatusNoContent)
        return
    }
    js, _ := json.Marshal(answer)
    res.Write(js)
}

func (a *AnswersApi) revsById(res http.ResponseWriter, req *http.Request) {
    id := bson.ObjectIdHex(mux.Vars(req)["id"])
    var revs Revisions
    if err := a.db.Answers.
        Find(bson.M{"_id": id}).
        One(&revs); err != nil {
        if err != mgo.ErrNotFound {
            a.log.Warnf("In 'AnswersApi.revsById', Query returned error for %s: %s", id, err)
            res.WriteHeader(http.StatusInternalServerError)
            return
        }
        a.log.Debugf("In 'AnswersApi.revsById', Query returned empty for %s", id)
        res.WriteHeader(http.StatusNoContent)
        return
    }
    js, _ := json.Marshal(revs)
    res.Write(js)
}

func (a *AnswersApi) deleteById(res http.ResponseWriter, req *http.Request) {
    params := mux.Vars(req)
    fmt.Fprintf(res, "delete answer %s", params["id"])
}

func (a *AnswersApi) newRevById(res http.ResponseWriter, req *http.Request) {
    params := mux.Vars(req)
    fmt.Fprintf(res, "save answer %s", params["id"])
}

func (a *AnswersApi) newByQid(res http.ResponseWriter, req *http.Request) {
    params := mux.Vars(req)
    fmt.Fprintf(res, "save answer %s", params["id"])
}
