package main

import (
  "code.google.com/p/go.crypto/bcrypt"
  "errors"
  "labix.org/v2/mgo/bson"
  "regexp"
  "time"
  "unicode/utf8"
)

var (
  ErrCSRFMissing            = errors.New(msgCSRFMissing)
  ErrCSRFDoesNotMatch       = errors.New(msgCSRFDoesNotMatch)
  ErrPasswordTooShort       = errors.New(msgPasswordTooShort)
  ErrIllegalEmail           = errors.New(msgIllegalEmail)
  ErrEmailAlreadyRegistered = errors.New(msgEmailAlreadyRegistered)
  ErrRegistering            = errors.New(msgRegistering)
)

type User struct {
  ID       bson.ObjectId `bson:"_id,omitempty"`
  Created  time.Time     `bson:"c"`
  Updated  time.Time     `bson:"u,omitempty"`
  Email    string
  Password []byte
}

//SetPassword takes a plaintext password and hashes it with bcrypt and sets the
//password field to the hash.
func (u *User) SetPassword(password string) {
  hpass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
  if err != nil {
    panic(err) //this is a panic because bcrypt errors on invalid costs
  }
  u.Password = hpass
}

//Login validates and returns a user object if they exist in the database.
func Login(ctx *Context, email, password string) (u *User, err error) {
  err = ctx.C("users").Find(bson.M{"email": email}).One(&u)
  if err != nil {
    return
  }

  err = bcrypt.CompareHashAndPassword(u.Password, []byte(password))
  if err != nil {
    u = nil
  }
  return
}

func ValidateNewUser(ctx *Context, email, password string) error {

  // email uniquness 
  n, err := ctx.C("users").Find(bson.M{"email": email}).Count()
  if err != nil {
    return ErrRegistering
  }
  if n != 0 {
    return ErrEmailAlreadyRegistered
  }

  // email validation
  emailRx := regexp.MustCompile(`(?i)\A[\w+\-.]+@[a-z\d\-.]+\.[a-z]+\z`)
  if ok := emailRx.MatchString(email); !ok {
    return ErrIllegalEmail
  }

  // password length 
  if ok := utf8.RuneCountInString(password) >= 6; !ok {
    return ErrPasswordTooShort
  }

  return nil
}
