package main

import (
  "html/template"
  "labix.org/v2/mgo/bson"
  "log"
  "net/http"
  "time"
  "fmt"
)


func notfoundHandler(w http.ResponseWriter, req *http.Request) {
  // fix this 
  t := template.Must(template.New("404.html").ParseFiles("templates/404.html"))
  if err := t.Execute(w, nil); err != nil {
    log.Panic(err)
  }
}

func homeHandler(w http.ResponseWriter, req *http.Request, ctx *Context) (err error) {
  return T("home.html").Execute(w, map[string]interface{}{
    "ctx": ctx,
  })
}

func handleNewUser(w http.ResponseWriter, req *http.Request, ctx *Context) (err error) {
  return T("register.html").Execute(w, map[string]interface{}{
    "ctx": ctx,
  })
}

func handleCreateUser(w http.ResponseWriter, req *http.Request, ctx *Context) (err error) {
  email, password := req.FormValue("email"), req.FormValue("password")

  if err := ValidateNewUser(ctx, email, password); err != nil {
    ctx.Session.AddFlash(err.Error()+".", "danger")
    http.Redirect(w, req, reverse("signup_route"), http.StatusSeeOther)
    return nil
  }

  u := &User{
    Email:   email,
    ID:      bson.NewObjectId(),
    Created: time.Now(),
  }
  u.SetPassword(password)


  if err := ctx.C("users").Insert(u); err != nil {
    ctx.Session.AddFlash(msgUserRegisterFail, "danger")
    return handleNewUser(w, req, ctx)
  }

  //store the user id in the values and redirect to index
  ctx.Session.Values["user"] = u.ID
  ctx.Session.AddFlash(fmt.Sprintf(msgWelcomeNewUser,  email), "success")
  http.Redirect(w, req, reverse("homepage_route"), http.StatusSeeOther)
  return nil
}

// handleLogout deletes the session user value and redirect to the homepage
func handleLogout(w http.ResponseWriter, req *http.Request, ctx *Context) (err error) {
  delete(ctx.Session.Values, "user")
  http.Redirect(w, req, reverse("homepage_route"), http.StatusSeeOther)
  return nil
}


func handleLogin(w http.ResponseWriter, req *http.Request, ctx *Context) (err error) {
  email, password := req.FormValue("email"), req.FormValue("password")

  user, e := Login(ctx, email, password)
  if e != nil {
    ctx.Session.AddFlash(msgInvalidUserPassword, "danger")
    http.Redirect(w, req, reverse("homepage_route"), http.StatusSeeOther)
    return nil
  }

  //store the user id in the values and redirect to index
  ctx.Session.Values["user"] = user.ID
  ctx.Session.AddFlash(fmt.Sprintf(msgWelcomeUser,email), "success")
  http.Redirect(w, req, reverse("homepage_route"), http.StatusSeeOther)
  return nil
}

// photoIndexHandler displays the progress bar and photo list 
func handlePhotosIndex(w http.ResponseWriter, req *http.Request, ctx *Context) (err error) {
  session_id := ctx.Session.Values["user"].(bson.ObjectId)
  
  return T("photos_index.html").Execute(w, map[string]interface{}{
    "ctx": ctx,
    "session_id": session_id.Hex(),
    "test": "wow!",
  })
}

type Message struct {
    RequestID      int
    Command        string
    SomeOtherThing string
    Success        bool
}


