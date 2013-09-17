package main

import (
  "encoding/gob"
  "github.com/gorilla/mux"
  "github.com/gorilla/sessions"
  "labix.org/v2/mgo"
  "labix.org/v2/mgo/bson"
  "log"
  "net/http"
  "code.google.com/p/go.net/websocket"
)

func init() {
  gob.Register(bson.ObjectId(""))
}

var store sessions.Store
var session *mgo.Session
var router *mux.Router

var ws_chan = make(chan string)

func main() {
  var err error
  session, err = mgo.Dial(config["db_url"])
  if err != nil {
    panic(err)
  }

  //create an index for the username field on the users collection
  if err := session.DB(config["db_name"]).C("users").EnsureIndex(mgo.Index{
    Key:    []string{"email"},
    Unique: true,
  }); err != nil {
    panic(err)
  }

  //store = sessions.NewCookieStore([]byte(config["auth_key"]), []byte(config["enc_key"]))
  store = sessions.NewCookieStore([]byte(config["auth_key"]))

  router = mux.NewRouter()

  //--- Router initialization
  router.HandleFunc("/", makeHandler(homeHandler)).Methods("GET").Name("homepage_route")
  router.HandleFunc("/register", makeHandler(handleNewUser)).Methods("GET").Name("signup_route")
  router.HandleFunc("/register", makeHandler(handleCreateUser)).Methods("POST")
  router.HandleFunc("/logout", makeHandler(handleLogout)).Methods("GET").Name("logout_route")
  router.HandleFunc("/login", makeHandler(handleLogin)).Methods("POST").Name("login_route")
  router.HandleFunc("/login", makeHandler(handleLogin)).Methods("POST").Name("login_route")
  router.HandleFunc("/photos", makeHandler(handlePhotosIndex)).Methods("GET").Name("photos_index_route")
  router.HandleFunc("/authorize", makeHandler(handleAuthorize)).Methods("GET").Name("fb_authe_route")
  router.HandleFunc("/facebook/redir", makeHandler(handleOAuth2Callback)) 

  // Router 404 handler
  router.NotFoundHandler = http.HandlerFunc(notfoundHandler)

  // Registering router
  http.Handle("/", router)

  // Handling statid assets
  http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets/"))))

  // Register websocket handler
  go h.run()
  http.Handle("/ws", websocket.Handler(wsHandler))



  // Start server
  if err := http.ListenAndServe(":3000", nil); err != nil {
    log.Fatal("ListenAndServe: ", err)
  }

}


