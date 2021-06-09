package storage

import "errors"

var UserExists = errors.New("user exists")
var UserNotFound = errors.New("user or password is incorrect")
var FileNotFound = errors.New("file not found or you haven't access")
var ShareExists = errors.New("share exists")
var ShareUserNotExists = errors.New("user you share with not exists")
var NotOwner = errors.New("you are not owner or audio not exists")
var NotAacFile = errors.New("file is not Aac")
