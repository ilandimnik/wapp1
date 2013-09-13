package main

import (
  "labix.org/v2/mgo/bson"
  "time"
)

type Identity struct {
  ID           bson.ObjectId `bson:"_id,omitempty"`
  Created      time.Time     `bson:"c"`
  Updated      time.Time     `bson:"u,omitempty"`
  UID          string        `bson:"uid"`
  AccessToken  string
  RefreshToken string
  SNetwork     string
  TokenExpiry  time.Time
  User_id      bson.ObjectId 
}


