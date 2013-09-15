package main

import (
  "code.google.com/p/goauth2/oauth"
  "errors"
  "fmt"
  "labix.org/v2/mgo"
  "labix.org/v2/mgo/bson"
  "time"
)

type UCache struct {
  Id       bson.ObjectId 
  Database *mgo.Database
}

func (u *UCache) C(name string) *mgo.Collection {
  return u.Database.C(name)
}

func NewUCache(id bson.ObjectId) (uc *UCache) {
  uc = &UCache{
    Database: session.Clone().DB(config["db_name"]),
    Id:       id,
  }
  return
}

func (u *UCache) Close() {
  u.Database.Session.Close()
}

// Implementing Oauth cache interface to allow storing and retrieving all the tokens from persistant storage
func (u *UCache) Token() (*oauth.Token, error) {
  if u.Id != "" {
    // is user is signed it, find out if he has idenity
    // get token values from database
    identity := Identity{}
    if err := u.C("identities").Find(bson.M{"user_id": u.Id}).One(&identity); err == nil {
      tok := &oauth.Token{}
      tok.AccessToken = identity.AccessToken
      tok.RefreshToken = identity.RefreshToken
      tok.Expiry = identity.TokenExpiry
      return tok, nil
    }
  }
  return nil, nil
}

func (u *UCache) PutToken(t *oauth.Token) error {
  if u.Id != "" {
    identity := Identity{}
    if err := u.C("identities").Find(bson.M{"user_id": u.Id}).One(&identity); err != nil {
      // if there's an identity   
      colQuerier := bson.M{"_id": identity.ID}
      change := bson.M{"$set": bson.M{"Updated": time.Now(), "AccessToken": t.AccessToken, "RefreshToken": t.RefreshToken, "TokenExpiry": t.Expiry}}
      if err := u.C("identities").Update(colQuerier, change); err != nil {
        return errors.New(msgInternalError)
      }
      fmt.Println("In context put Token - updated token in database")
    }
  }
  return nil
}
