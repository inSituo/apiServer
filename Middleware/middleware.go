package Middleware

import (
    "encoding/json"
    "github.com/gorilla/mux"
    "gopkg.in/mgo.v2/bson"
    "net/http"
)

func IdVerifier(res http.ResponseWriter, req *http.Request) {
    id := mux.Vars(req)["id"]
    if id != "" {
        if !bson.IsObjectIdHex(id) {
            res.Header().Set("X-DEBUG-CHAIN-BREAK", "Invalid Id")
            res.WriteHeader(http.StatusBadRequest)
            js, _ := json.Marshal(map[string]string{
                "error": "invalid id",
            })
            res.Write(js)
        }
    }
}
