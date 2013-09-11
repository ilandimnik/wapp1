package main

import (
  "code.google.com/p/goauth2/oauth"
  "fmt"
  "net/http"
  "io/ioutil"
)

var oauthCfg = &oauth.Config{ //setup
  ClientId:     "1386314934930605",
  ClientSecret: "38c8bffc9e8eba7e8c4e216af883b9ab",
  AuthURL:      "https://www.facebook.com/dialog/oauth",
  TokenURL:     "https://graph.facebook.com/oauth/access_token",
  //RedirectURL:  config["domain"] + "/facebook/redir", - postponed to later, when we have the config struct
  Scope:        "",
}

func handleAuthorize(w http.ResponseWriter, r *http.Request) {

  oauthCfg.RedirectURL = "http://" + config["domain"] + "/facebook/redir"

  //Get the Google URL which shows the Authentication page to the user
  url := oauthCfg.AuthCodeURL("")

  fmt.Println("Redirect to facebook")
  fmt.Println(oauthCfg.RedirectURL)

  //redirect user to that page
  http.Redirect(w, r, url, http.StatusFound)
}

// Function that handles the callback from the Google server
func handleOAuth2Callback(w http.ResponseWriter, r *http.Request) {

  //Get the code from the response
  code := r.FormValue("code")

  t := &oauth.Transport{Config: oauthCfg}
  // Exchange the received code for a token
  t.Exchange(code)

  // Get all user details
  //resp, err := t.Client().Get("https://graph.facebook.com/me")
  resp, err := t.Client().Get("https://graph.facebook.com/me?fields=id,name,email,photos,albums")

  if err != nil {
    fmt.Println("Received error:" + err.Error())
  }

  fmt.Println("Received Response")
  fmt.Println(resp)

  if resp.StatusCode == 200 { // OK 
    bodyBytes, _ := ioutil.ReadAll(resp.Body) 
    bodyString := string(bodyBytes) 
    fmt.Println(bodyString)
  }


  //now get user data based on the Transport which has the token
  //    resp, _ := t.Client().Get(profileInfoURL)
  //
  //    buf := make([]byte, 1024)
  //    resp.Body.Read(buf)
  //    userInfoTemplate.Execute(w, string(buf))
}
