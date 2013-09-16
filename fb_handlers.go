package main

import (
  "code.google.com/p/goauth2/oauth"
  "encoding/json"
  "errors"
  "fmt"
  "labix.org/v2/mgo/bson"
  "log"
  "net/http"
  "net/url"
  "strings"
  "time"

//  "io/ioutil"
)

var (
  ErrInternalError              = errors.New(msgInternalError)
  ErrUnableToObtainToken        = errors.New(msgUnableToObtainToken)
  ErrParsingPhotosJson          = errors.New(msgErrorParsingPhotosJson)
  ErrFBFailedRetrieveUserPhotos = errors.New(msgFBFailedRetrieveUserPhotos)
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

type FBAlbums struct {
  Id     string
  Albums FBAlbum
}

type FBAlbum struct {
  Data   []FBAlbumData
  Paging FBPaging
}

type FBPaging struct {
  Previous string
  Next     string
}

type FBAlbumData struct {
  Id     string
  Name   string
  Photos FBPhoto
}

type FBPhoto struct {
  Data   []FBPhotoData
  Paging FBPaging
}

type FBPhotoData struct {
  Id      string
  Pictrue string
  Source  string
  Height  int
  Width   int
  Images  []FBImage
  Link    string
}

type FBImage struct {
  Source string
  Height int
  Width  int
}

var oauthCfg = &oauth.Config{ //setup
  ClientId:     "1386314934930605",
  ClientSecret: "38c8bffc9e8eba7e8c4e216af883b9ab",
  AuthURL:      "https://www.facebook.com/dialog/oauth",
  TokenURL:     "https://graph.facebook.com/oauth/access_token",
  //RedirectURL:  config["domain"] + "/facebook/redir", - postponed to later, when we have the config struct
  Scope: "",
}

func handleAuthorize(w http.ResponseWriter, r *http.Request, ctx *Context) (err error) {

  oauthCfg.RedirectURL = "http://" + config["domain"] + "/facebook/redir"

  //Get the Google URL which shows the Authentication page to the user
  url := oauthCfg.AuthCodeURL("")

  //redirect user to that page
  http.Redirect(w, r, url, http.StatusFound)
  return nil
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
    // not found
    new_u := User{
      Email:   fbuserdata.Email,
      ID:      bson.NewObjectId(),
      Created: time.Now(),
    }
    if err := ctx.C("users").Insert(&new_u); err != nil {
      log.Println(err)
      ctx.Session.AddFlash(msgUserRegisterFail, "danger")
      return handleNewUser(w, r, ctx)
    }
    u = new_u
  }

  // check if the user already have a facebook identity document
  identity := Identity{}
  if err = ctx.C("identities").Find(bson.M{"user_id": u.ID}).One(&identity); err != nil {
    // if there wasn't an idenity, create one
    //create new identity
    identity := Identity{
      ID:           bson.NewObjectId(),
      Created:      time.Now(),
      UID:          fbuserdata.Id,
      User_id:      u.ID,
      AccessToken:  t.AccessToken,
      RefreshToken: t.RefreshToken,
      TokenExpiry:  t.Expiry,
      SNetwork:     "Facebook",
    }
    if err := ctx.C("identities").Insert(&identity); err != nil {
      log.Println(err)
      ctx.Session.AddFlash(msgUserRegisterFail, "danger")
      return handleNewUser(w, r, ctx)
    }
  } else {
    colQuerier := bson.M{"_id": identity.ID}
    change := bson.M{"$set": bson.M{"Updated": time.Now(), "AccessToken": t.AccessToken, "RefreshToken": t.RefreshToken, "TokenExpiry": t.Expiry}}
    if err := ctx.C("identities").Update(colQuerier, change); err != nil {
      log.Println(err)
      ctx.Session.AddFlash(msgUserRegisterFail, "danger")
      http.Redirect(w, r, reverse("homepage_route"), http.StatusSeeOther)
    }
  }

  // User
  ctx.Session.Values["user"] = u.ID
  ctx.Session.AddFlash(fmt.Sprintf(msgWelcomeNewUser, u.Email), "success")

  uc := NewUCache(ctx.Session.Values["user"].(bson.ObjectId))
  oauthCfg.TokenCache = uc
  getUserPhotos(uc)
  uc.Close()

  http.Redirect(w, r, reverse("homepage_route"), http.StatusSeeOther)
  return nil
}

//
// getUserPhotos fetch all the photos from the user account
//
func getUserPhotos(u *UCache) error {
  oauthCfg.TokenCache = u
  t := &oauth.Transport{Config: oauthCfg}
  token, err := oauthCfg.TokenCache.Token()

  if err != nil {
    log.Println(msgUnableToObtainToken)
    return ErrUnableToObtainToken
  }
  t.Token = token

  //Get all user details
  identity := Identity{}
  if err := u.C("identities").Find(bson.M{"user_id": u.Id}).One(&identity); err != nil {
    log.Println("Received error from facebook...")
    return ErrUnableToObtainToken
  }

  resp, err := t.Client().Get("https://graph.facebook.com/" + identity.UID + "?fields=albums.fields(photos)")
  //resp, err := t.Client().Get("https://graph.facebook.com/" + identity.UID + "?fields=albums.limit(2).fields(photos)")
  if err != nil {
    log.Println("Received error from facebook...")
    return ErrFBFailedRetrieveUserPhotos
  }

  fbalbums := FBAlbums{}
  if resp.StatusCode == 200 { // OK 
    if err = json.NewDecoder(resp.Body).Decode(&fbalbums); err != nil {
      log.Println(err)
      return ErrParsingPhotosJson
    }
    resp.Body.Close()
  } else {
    resp.Body.Close()
    return errors.New(fmt.Sprintf(msgReceivedStatusError, resp.StatusCode))
  }

  for {
    for _, album := range fbalbums.Albums.Data {
      for _, photo := range album.Photos.Data {
        if err := InsertOrUpdatePhoto(u, &photo); err != nil {
          log.Println(err)
          break
        }
      }
    }

    if fbalbums.Albums.Paging.Next == "" {
      break
    } else {
      u, err := url.Parse(fbalbums.Albums.Paging.Next)
      if err != nil {
        log.Println(err)
        break
      }
      values := u.Query()
      after_str := strings.Replace(values["after"][0], "=", "", -1)
      //resp, err = t.Client().Get("https://graph.facebook.com/" + identity.UID + "?fields=albums.after(" + after_str + ").limit(2).fields(photos)")
      resp, err = t.Client().Get("https://graph.facebook.com/" + identity.UID + "?fields=albums.after(" + after_str + ").fields(photos)")

      if resp.StatusCode == 200 { // OK 
        fbalbums = FBAlbums{}
        if err = json.NewDecoder(resp.Body).Decode(&fbalbums); err != nil {
          log.Println(err)
          return ErrParsingPhotosJson
        }
        resp.Body.Close()
      } else {
        resp.Body.Close()
        return errors.New(fmt.Sprintf(msgReceivedStatusError, resp.StatusCode))
      }
    }
  }

  return nil
}

// Insert or update photo to the database
func InsertOrUpdatePhoto(u *UCache, photo *FBPhotoData) error {
  ph := Photo{}
  if err := u.C("photos").Find(bson.M{"user_id": u.Id, "fb_photo_id": photo.Id}).One(&ph); err != nil {
    // if we couldn't find it, create a new one
    ph = Photo{
      ID:          bson.NewObjectId(),
      Created:     time.Now(),
      User_id:     u.Id,
      FB_photo_id: photo.Id,
      Source:      photo.Source,
      Width:       photo.Width,
      Height:      photo.Height,
    }
    if err := u.C("photos").Insert(&ph); err != nil {
      log.Println(err)
      return err
    }
  } else {
    // other wise just update
    colQuerier := bson.M{"_id": ph.ID}
    change := bson.M{"$set": bson.M{"Updated": time.Now(), "Source": photo.Source, "Width": photo.Width, "Height": photo.Height}}
    if err := u.C("identities").Update(colQuerier, change); err != nil {
      log.Println(err)
      return err
    }
  }
  return nil
}
