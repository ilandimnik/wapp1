package main

import (
  "code.google.com/p/goauth2/oauth"
  "encoding/json"
  "labix.org/v2/mgo/bson"
  "log"
  "net/http"
  "time"
  "fmt"
)

type FBUserData struct {
  Id         string `json:"id"`
  Name       string
  First_name string
  Last_name  string
  Link       string
  Username   string
  Email      string
}

var oauthCfg = &oauth.Config{ //setup
  ClientId:     "1386314934930605",
  ClientSecret: "38c8bffc9e8eba7e8c4e216af883b9ab",
  AuthURL:      "https://www.facebook.com/dialog/oauth",
  TokenURL:     "https://graph.facebook.com/oauth/access_token",
  //RedirectURL:  config["domain"] + "/facebook/redir", - postponed to later, when we have the config struct
  Scope: "",
}

func handleAuthorize(w http.ResponseWriter, r *http.Request) {

  oauthCfg.RedirectURL = "http://" + config["domain"] + "/facebook/redir"

  //Get the Google URL which shows the Authentication page to the user
  url := oauthCfg.AuthCodeURL("")

  //redirect user to that page
  http.Redirect(w, r, url, http.StatusFound)
}

// handleOAuth2Callback handles the callback from the Google server
func handleOAuth2Callback(w http.ResponseWriter, r *http.Request, ctx *Context) (err error) {

  //Get the code from the response
  code := r.FormValue("code")

  t := &oauth.Transport{Config: oauthCfg}
  // Exchange the received code for a token
  t.Exchange(code)

  //Get all user details
  resp, err := t.Client().Get("https://graph.facebook.com/me")
  if err != nil {
    log.Println("Received error from Facebook:" + err.Error())
    ctx.Session.AddFlash(msgFBFailedRetrieveUserInfo, "danger")
    http.Redirect(w, r, reverse("homepage_route"), http.StatusSeeOther)
    return nil
  }
  var fbuserdata FBUserData
  if resp.StatusCode == 200 { // OK 

    if err = json.NewDecoder(resp.Body).Decode(&fbuserdata); err != nil {
      log.Println(err)
    }
  } else {
    //unable to obtain user data
    ctx.Session.AddFlash(msgFBFailedRetrieveUserInfo, "danger")
    http.Redirect(w, r, reverse("homepage_route"), http.StatusSeeOther)
    return nil
  }

  // Check user email isn't empty
  if fbuserdata.Email == "" {
    ctx.Session.AddFlash(msgFBFailedRetrieveUserInfo, "danger")
    http.Redirect(w, r, reverse("homepage_route"), http.StatusSeeOther)
    return nil
  }

  // find if there's a user with the same email
  u := User{}
  err = ctx.C("users").Find(bson.M{"email": fbuserdata.Email}).One(&u)
  if err != nil {
    log.Println(err)
    ctx.Session.AddFlash(msgFBFailedRetrieveUserInfo, "danger")
    http.Redirect(w, r, reverse("homepage_route"), http.StatusSeeOther)
    return nil
  }

  //user doesn't exist - we need to create one
  if u.ID == "" {
    new_u := User{
      Email:   fbuserdata.Email,
      ID:      bson.NewObjectId(),  
      Created: time.Now(),
    }
    if err := ctx.C("users").Insert(&new_u); err != nil {
      ctx.Session.AddFlash(msgUserRegisterFail, "danger")
      return handleNewUser(w, r, ctx)
    }
    u = new_u
  }

  //create new identity
  identity := Identity {
    ID:      bson.NewObjectId(),
    Created: time.Now(),
    UID: fbuserdata.Id,
    User_id: u.ID,
    AccessToken: t.AccessToken,
    RefreshToken: t.RefreshToken,
    TokenExpiry: t.Expiry,
  }

  if err := ctx.C("identities").Insert(&identity); err != nil {
    ctx.Session.AddFlash(msgUserRegisterFail, "danger")
    return handleNewUser(w, r, ctx)
  }

  // User
  ctx.Session.Values["user"] = u.ID
  ctx.Session.AddFlash(fmt.Sprintf(msgWelcomeNewUser,  u.Email), "success")
  http.Redirect(w, r, reverse("homepage_route"), http.StatusSeeOther)
  return nil

  // Get user data

  //  // Get all user details
  //resp, err := t.Client().Get("https://graph.facebook.com/me")
  //  resp, err := t.Client().Get("https://graph.facebook.com/me?fields=id,name,email,photos,albums")
  //
  //  if err != nil {
  //    fmt.Println("Received error:" + err.Error())
  //  }
  //
  //  fmt.Println("Received Response")
  //  fmt.Println(resp)
  //
  //  if resp.StatusCode == 200 { // OK 
  //    bodyBytes, _ := ioutil.ReadAll(resp.Body) 
  //    bodyString := string(bodyBytes) 
  //    fmt.Println(bodyString)
  //  }
  //

  //now get user data based on the Transport which has the token
  //    resp, _ := t.Client().Get(profileInfoURL)
  //
  //    buf := make([]byte, 1024)
  //    resp.Body.Read(buf)
  //    userInfoTemplate.Execute(w, string(buf))
  return nil
}
