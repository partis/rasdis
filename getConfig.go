package main

import (
  "encoding/json"
  "io/ioutil"
  "github.com/golang/glog"
)

type Config struct {
  ForumURL string `json:"forumURL"`
  ForumUsername string `json:"forumUsername"`
  ForumPass string `json:"forumPass"`
}

func ReadConfig(configfile string) Config {
  raw, err := ioutil.ReadFile(configfile)
  if err != nil {
    glog.Fatal(err)
  }

  var config Config
  json.Unmarshal(raw, &config)

  glog.Info(config)
  return config
}
