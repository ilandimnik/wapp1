package main

import (
  "html/template"
  "labix.org/v2/mgo/bson"
  "log"
  "net/http"
  "time"
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
//    return handleNewUser(w, req, ctx)
  }

  u := &User{
    Email:   email,
    ID:      bson.NewObjectId(),
    Created: time.Now(),
  }
  u.SetPassword(password)



  if err := ctx.C("users").Insert(u); err != nil {
    ctx.Session.AddFlash("Problem registering user.")
    return handleNewUser(w, req, ctx)
  }

  //store the user id in the values and redirect to index
  ctx.Session.Values["user"] = u.ID
  ctx.Session.AddFlash("Welcome " + email + ".", "success")
  http.Redirect(w, req, reverse("homepage_route"), http.StatusSeeOther)
  return nil
}


func handleLogout(w http.ResponseWriter, req *http.Request, ctx *Context) (err error) {
  delete(ctx.Session.Values, "user")
  http.Redirect(w, req, reverse("homepage_route"), http.StatusSeeOther)
  return nil
}


func handleLogin(w http.ResponseWriter, req *http.Request, ctx *Context) (err error) {
  email, password := req.FormValue("email"), req.FormValue("password")

  user, e := Login(ctx, email, password)
  if e != nil {
    ctx.Session.AddFlash("Invalid Username/Password", "danger")
    http.Redirect(w, req, reverse("homepage_route"), http.StatusSeeOther)
    return nil
  }

  //store the user id in the values and redirect to index
  ctx.Session.Values["user"] = user.ID
  ctx.Session.AddFlash("Welcome back " + email + ".", "success")
  http.Redirect(w, req, reverse("homepage_route"), http.StatusSeeOther)
  return nil
}



