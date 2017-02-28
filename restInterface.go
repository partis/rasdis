package main

import (
  "net/http"
  "io/ioutil"
  "encoding/json"
  "log"
  "github.com/golang/glog"
  "fmt"
  "crypto/tls"
  "strings"
  "bytes"
)

func callRest(baseUrl string, config Config) []byte {

  url := fmt.Sprintf(config.ForumURL + baseUrl)

  glog.Info("url is: ", url)

  req, err := http.NewRequest("GET", url, nil)
  if err != nil {
    glog.Fatal("Unable to create new request: ", err.Error)
  }

  req.SetBasicAuth(config.ForumUsername, config.ForumPass)

  client := &http.Client{}

  if strings.HasPrefix(url, "https") {

    tr := &http.Transport{
      TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }

    client = &http.Client{Transport: tr}
  }

  resp, err := client.Do(req)
  if err != nil {
    glog.Fatal("Error connecting to url: ", err.Error)
  }

  defer resp.Body.Close()
  
  var response = []byte("")

  glog.V(1).Info(resp.StatusCode)

  response, err = ioutil.ReadAll(resp.Body)
    if err != nil {
      glog.Fatal(err.Error)
    }

  switch resp.StatusCode {
  case 401:
    glog.Fatal("User not authenticated, please check the user in the config file\n" + string(response))
  case 200:
    glog.Info("200 ok back from forum")
  default:
  //if resp.StatusCode != 200 {
    glog.Info("Response was not 200 ok: " + string(response))
    response = []byte("")
  }

  return response
}

func getContentPolicyList(policyType string, config Config) policyList {
  var policy policyList

  url := "/restApi/v1.0/policies/" + policyType + "Policies"

  glog.V(1).Info("url is: ", url)

  resp := callRest(url, config)

  glog.V(2).Info(resp)

  json.Unmarshal(resp, &policy)

  glog.V(1).Info("Name: ", policy.Policy[0].Name)
  glog.V(1).Info("Url: ", policy.Policy[0].URL)

  return policy
}

func getContentPolicy(policyName string, policyType string, config Config) contentPolicy {
  var policy  contentPolicy

  url := "/restApi/v1.0/policies/" + policyType + "Policies/" + policyName

  resp := callRest(url, config)

  glog.V(2).Info(resp)

  json.Unmarshal(resp, &policy)

  return policy
}

func getVirtualDirectoryList(policyName string, policyType string, config Config) policyList {
  var policy policyList

  url := "/restApi/v1.0/policies/" + policyType + "Policies/" + policyName + "/virtualDirectories"

  resp := callRest(url, config)

  glog.V(2).Info(resp)

  json.Unmarshal(resp, &policy)

  return policy
}

func getVirtualDirectory(policyName string, virtualDirectoryName string, policyType string, config Config) virtualDirectory {
  var policy virtualDirectory

  url := "/restApi/v1.0/policies/" + policyType + "Policies/" + policyName + "/virtualDirectories/" + virtualDirectoryName

  resp := callRest(url, config)

  glog.V(2).Info(resp)

  json.Unmarshal(resp, &policy)

  return policy
}

func getListenerPolicy(listenerType string, listenerName string, config Config) map[string]json.RawMessage {
  var policy map[string]json.RawMessage

  url := "/restApi/v1.0/policies/" + listenerType + "ListenerPolicies/" + listenerName

  resp := callRest(url, config)

  glog.V(2).Info(resp)

  err := json.Unmarshal(resp, &policy)
  if err != nil {
    log.Fatal(err)
  }

  return policy
}

func getDocument(projectDetails string, description string, docType string, config Config) (Document string, docExists bool) {
  var inter interface{}
  url := "/restApi/v1.0/policies/documents/" + projectDetails + "_document_policy_" + docType + "_" + description
  
  resp := callRest(url, config)

  buffer := new(bytes.Buffer)
  err := json.Compact(buffer, resp)
  if err != nil {
    glog.Warning("Unable to remove whitespace from JSON")
  }

  resp = buffer.Bytes()
 
  glog.V(2).Info(string(resp)) 
  if string(resp) != "" {
    docExists = true
    err := json.Unmarshal(resp, &inter)
    if err != nil {
      log.Fatal(err)
    }
    Document = (inter.(map[string]interface{}))["document"].(string)

    Document = removeQuotes(Document)

    if docType == "parameters" {
      Document = "[" + Document + "]"
    }

    glog.V(2).Info("Document retrieved from forum is: " + Document)

    if jsonDoc, ok := isJson(Document); ok {
      //var jsonDoc map[string]interface{}
      //err = json.Unmarshal([]byte(Document), &jsonDoc)
      //if err != nil {
      //  log.Fatal(err)
      //}

      glog.V(2).Info(jsonDoc)
      i := 0

      keys := make([]string, len(jsonDoc))
      for s, _ := range jsonDoc {
        keys[i] = s
        i++
      }
    
      glog.V(2).Info(Document)
     
      return Document, docExists
    } else {
      glog.Warning("Your document isn't valid JSON, please check and upload valid JSON to create your model definition")
    }
  }

  return "", false
}

