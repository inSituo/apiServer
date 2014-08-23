package Api

import (
    "encoding/xml"
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
    *ApiSection
}

func InitAnswersApi(
    r *mux.Router,
    db *DBAccess.DB,
    log *LeveledLogger.Logger,
    setRes Middleware.ResponseSetter,
) {
    iname := "InitAnswersApi"
    log.Call(iname)

    a := AnswersApi{}
    a.ApiSection = NewApiSection(db, log, r, setRes)

    a.use(Middleware.IdVerifier(a.log, a.c, a.setRes))
    a.use(a.GetUserInfo)
    // all methods in this api require a logged-in user
    a.use(func(res http.ResponseWriter, req *http.Request) {
        if context.Get(req, "loggedIn") != true {
            a.respondNotLoggedIn(res, req)
            a.c.Break(req)
        }
    })

    a.setupRoute("GET", "/{id}", a.getById)
    a.setupRoute("GET", "/revs/{id}", a.revsById)
    a.setupRoute("DELETE", "/{id}", a.deleteById)
    a.setupRoute("POST", "/rev/{id}", a.newRevById)
    a.setupRoute("POST", "/{id}", a.newByQid)
}

type Answer struct {
    XMLName xml.Name        `bson:"-" json:"-" xml:"answer"`
    ID      bson.ObjectId   `bson:"_id" json:"id" xml:"id"`
    QID     bson.ObjectId   `bson:"qid" json:"qid" xml:"qid"`
    OUID    bson.ObjectId   `bson:"ouid" json:"ouid" xml:"ouid"`
    OTS     int             `bson:"ots" json:"ots" xml:"ots"`
    Rev     ContentRevision `bson:"rev" json:"rev" xml:"rev"`
}

type Revisions struct {
    XMLName xml.Name          `bson:"-" json:"-" xml:"revs"`
    ID      bson.ObjectId     `bson:"_id" json:"id" xml:"id"`
    QID     bson.ObjectId     `bson:"qid" json:"qid" xml:"qid"`
    Revs    []ContentRevision `bson:"revs" json:"revs" xml:"revs"`
}

func (a *AnswersApi) getById(res http.ResponseWriter, req *http.Request) {
    iname := "AnswersApi.getById"
    id := bson.ObjectIdHex(mux.Vars(req)["id"])
    user := context.Get(req, "user").(*UserInfo)
    a.log.Call(iname, user.ID.Hex(), id.Hex())
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
            a.log.Warn(iname, "pipe error", err, id.Hex())
            a.setRes(req, http.StatusInternalServerError, nil)
            return
        }
        a.log.Debug(iname, "pipe empty", id.Hex())
        a.setRes(req, http.StatusNotFound, nil)
        return
    }
    a.setRes(req, http.StatusOK, answer)
}

func (a *AnswersApi) revsById(res http.ResponseWriter, req *http.Request) {
    iname := "AnswersApi.revsById"
    id := bson.ObjectIdHex(mux.Vars(req)["id"])
    user := context.Get(req, "user").(*UserInfo)
    a.log.Call(iname, user.ID.Hex(), id.Hex())
    var revs Revisions
    if err := a.db.Answers.
        Find(bson.M{"_id": id}).
        One(&revs); err != nil {
        if err != mgo.ErrNotFound {
            a.log.Warn(iname, "query error", err, id.Hex())
            a.setRes(req, http.StatusInternalServerError, nil)
            return
        }
        a.log.Debug(iname, "query empty", id.Hex())
        a.setRes(req, http.StatusNotFound, nil)
        return
    }
    a.setRes(req, http.StatusOK, revs)
}

func (a *AnswersApi) deleteById(res http.ResponseWriter, req *http.Request) {
    iname := "AnswersApi.deleteById"
    id := bson.ObjectIdHex(mux.Vars(req)["id"])
    user := context.Get(req, "user").(*UserInfo)
    a.log.Call(iname, user.ID.Hex(), id.Hex())
    // need to check if the first revision of this answer was posted by this
    // user. if yes, can delete. otherwise, no permission.
    count, err := a.db.Answers.Find(bson.M{
        "_id":        id,
        "revs.0.uid": user.ID,
    }).Count()
    switch {
    case err != nil:
        // server error
        a.log.Warn(iname, "query error", err, id.Hex(), user.ID.Hex())
        a.setRes(req, http.StatusInternalServerError, nil)
    case count == 0:
        // can't delete
        a.log.Info(iname, "delete denied", user.ID.Hex(), id.Hex())
        a.setRes(req, http.StatusForbidden, ErrRes{Reason: "not owner of answer"})
    default:
        // can delete
        err = a.db.Answers.Remove(bson.M{"_id": id})
        if err != nil {
            // server error
            a.log.Warn(iname, "remove error", err, id)
            a.setRes(req, http.StatusInternalServerError, nil)
        } else {
            // success
            a.log.Action(iname, user.ID.Hex(), id.Hex())
            a.setRes(req, http.StatusNoContent, nil)
        }
    }
}

func (a *AnswersApi) newRevById(res http.ResponseWriter, req *http.Request) {
    iname := "AnswersApi.newRevById"
    id := bson.ObjectIdHex(mux.Vars(req)["id"])
    user := context.Get(req, "user").(*UserInfo)
    a.log.Call(iname, user.ID.Hex(), id.Hex())
    // first check if this answer exists

    a.setRes(req, http.StatusNoContent, nil)
}

func (a *AnswersApi) newByQid(res http.ResponseWriter, req *http.Request) {
    iname := "AnswersApi.newByQid"
    id := bson.ObjectIdHex(mux.Vars(req)["id"])
    user := context.Get(req, "user").(*UserInfo)
    a.log.Call(iname, user.ID.Hex(), id.Hex())
    a.setRes(req, http.StatusNoContent, nil)
}
