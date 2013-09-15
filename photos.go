package main

import (
  "labix.org/v2/mgo/bson"
  "time"
)

type Photo struct {
  ID          bson.ObjectId `bson:"_id,omitempty"`
  Created     time.Time     `bson:"c"`
  Updated     time.Time     `bson:"u,omitempty"`
  User_id     bson.ObjectId
  FB_photo_id string
  Source      string
  Width       int
  Height      int
}
