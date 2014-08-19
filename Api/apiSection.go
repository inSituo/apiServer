package Api

import (
    "github.com/gorilla/context"
    "github.com/gorilla/mux"
    "github.com/inSituo/apiServer/DBAccess"
    "github.com/inSituo/apiServer/LeveledLogger"
    "github.com/inSituo/apiServer/Middleware"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "net/http"
    "time"
)

var (
    API_KEY_REQ_HEADER = "X-API-KEY"
)

type ApiSection struct {
    db  *DBAccess.DB
    log *LeveledLogger.Logger
    r   *mux.Router
    c   *Middleware.Chain
}

func (as *ApiSection) setupRoute(method, endpoint string, f http.HandlerFunc) {
    if as.r == nil {
        panic("'setupRoute' cannot be called before route is init'd.")
    }
    if as.log == nil {
        panic("'setupRoute' cannot be called before logger is init'd.")
    }
    as.log.Debugf("Setting up route %s %s", method, endpoint)
    g := f
    if as.c != nil {
        as.c.Push(f)
        g = as.c.MakeHandler()
        as.c.Pop()
    }
    as.r.HandleFunc(endpoint, g).Methods(method)
}

// This middleware checks if the request comes from an authenticated user.
// Set two context vars:
//   0. loggedIn: true/false
//   0. userId: ...
// API key is sent with an http header
func (as *ApiSection) GetUserId(res http.ResponseWriter, req *http.Request) {
    apiKey := req.Header.Get(API_KEY_REQ_HEADER)
    var login Login
    if err := as.db.Logins.
        Find(bson.M{"key": apiKey}).
        One(&login); err != nil {
        if err != mgo.ErrNotFound {
            as.log.Warnf("In 'ApiSection.GetUserId', Query returned error for %s: %s", apiKey, err)
            res.Header().Set("X-DEBUG-CHAIN-BREAK", "Api Key Verification")
            res.WriteHeader(http.StatusInternalServerError)
            return
        }
        as.log.Debugf("In 'ApiSection.GetUserId', Query returned empty for %s", apiKey)
        context.Set(req, "loggedIn", false)
        return
    }
    if login.Expires < time.Now().Unix() {
        as.log.Debugf("In 'ApiSection.GetUserId', Login expired for %s", apiKey)
        context.Set(req, "loggedIn", false)
        return
    }
    context.Set(req, "loggedIn", true)
    context.Set(req, "userId", login.UID)
}

func (as *ApiSection) use(f http.HandlerFunc) {
    if as.c == nil {
        as.c = &Middleware.Chain{}
    }
    as.c.Push(f)
}

type Location struct {
    Lat     float32 `bson:"lat" json:"lat"`
    Lon     float32 `bson:"lon" json:"lon"`
    Title   string  `bson:"title" json:"title"`
    Address string  `bson:"address" json:"address"`
}

type ContentRevision struct {
    UID     bson.ObjectId `bson:"uid" json:"uid"`
    TS      int           `bson:"ts" json:"ts"`
    Locs    []Location    `bson:"locs" json:"locs"`
    Content string        `bson:"content" json:"content"`
}

type Login struct {
    Key     string        `bson:"key"`
    UID     bson.ObjectId `bson:"uid"`
    Expires int64         `bson:"expires"`
}
