package main

import (
  "os"
  "flag"
  "fmt"
  "github.com/golang/glog"
)

func usage() {
  //print usage to iniate logging
  fmt.Fprintf(os.Stderr, "usage: example -stderrthreshold=[INFO|WARN|FATAL] -log_dir=[string]\n", )
  flag.PrintDefaults()
  os.Exit(2)
}

func init() {
  //set the usage to the above func
  flag.Usage = usage
  //parse the flags from the command line to configre logging
  flag.Parse()
}

func main() {
  glog.Info("Starting rasdis")
  //rasdis()
  
  startServer()
  glog.Flush()
}
