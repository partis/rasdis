package main

import (
  "encoding/json"
  "io/ioutil"
  "bytes"
  "strings"
  "github.com/golang/glog"
)

func swaggerToJson(template *SwaggerTemplate) string {
  bytes, err := json.Marshal(template)
  if err != nil {
    glog.Fatal(err.Error())
  }

  return string(bytes)
}

func toJson(template interface{}) string {
  bytes, err := json.Marshal(template)
  if err != nil {
    glog.V(2).Info(template)
    glog.V(2).Info(err.Error())
    glog.Warning("Unable to process json, as it is not valid. Please check the names and details configured for all policies")
    return "{\"error\": \"Unable to process json, as it is not valid. Please check the names and details configured for all policies\"}"
  }

  return string(bytes)
}

func isJson(s string) (map[string]interface{}, bool) {
  var js map[string]interface{}
  
  return js, json.Unmarshal([]byte(s), &js) == nil
}

func (def *DefinitionStruct) MarshalJSON() ([]byte, error) {
  var jsonBytes []byte = []byte("")
  
  for i := range def.definitions {
    
    properties := def.definitions[i].Properties
    //strip the escape characters from properties
    //properties = strings.Replace(properties, "\\", "", -1)


    xml, err := json.Marshal(def.definitions[i].Xml)
    if err != nil {
      return nil, err
    }


    properties = strings.Replace(properties, "\"properties\":{", "\"type\":\"" + string(def.definitions[i].Type) + "\",\"xml\":" + string(xml) + "," +  "\"properties\":{", 1)
    glog.Info("Properties are: " + properties)
    //properties = strings.Replace(properties, "\"properties\": {", string(def.definition[i]Type) + "," +  "\"properties\": {", 1)

    jsonBytes = append(jsonBytes, []byte(properties)...)
  }

  glog.V(2).Info("json from definiaton structs is " +  string(jsonBytes))

  if string(jsonBytes) == "" {
    jsonBytes = []byte("{}")
  }
  return jsonBytes, nil
}

func (pss *PathsStruct) MarshalJSON() ([]byte, error) {
  var jsonBytes []byte
  comma := []byte(",")
  currentPath := ""

  jsonBytes = append(jsonBytes, []byte("{")...)

  for i := range pss.paths {
    tag, err := json.Marshal(string(pss.paths[i].verbs[0].Tags[0]))
    operationId, err := json.Marshal(string(pss.paths[i].verbs[0].OperationID))
    if err != nil {
      glog.V(2).Infof("Unable to marshal tag or operationid: %s %s", string(pss.paths[i].verbs[0].Tags[0]), string(pss.paths[i].verbs[0].OperationID))
      return nil, err
    }

    path, verb := getPathAndVerbFromJson(string(tag), string(operationId))

    if path == currentPath {
      jsonBytes = bytes.TrimSuffix(jsonBytes, comma)
      jsonBytes = bytes.TrimSuffix(jsonBytes, []byte("}"))
      jsonBytes = append(jsonBytes, comma...)
    } else {
      jsonBytes = append(jsonBytes, []byte("\"" + path + "\": {")...)
      currentPath = path
    }

    for j := range pss.paths[i].verbs {
      jsonBytes = append(jsonBytes, []byte("\"" + verb + "\":")...)
      bytes, err := json.Marshal(pss.paths[i].verbs[j])
      if err != nil {
        glog.V(2).Infof("Unable to marshal verb: %s", pss.paths[i].verbs[j])
        return nil, err
      }
      jsonBytes = append(jsonBytes, bytes...)
      jsonBytes = append(jsonBytes, comma...)
      glog.V(2).Infof("json bytes after path %s: %s", path, string(jsonBytes))
    }
    jsonBytes = bytes.TrimSuffix(jsonBytes, comma)
    jsonBytes = append(jsonBytes, []byte("}")...)
    jsonBytes = append(jsonBytes, comma...)
  }

  jsonBytes = bytes.TrimSuffix(jsonBytes, comma)
  jsonBytes = append(jsonBytes, []byte("}")...)
  glog.V(2).Info("json bytes after all paths: %s", string(jsonBytes))
  return jsonBytes, nil
}

func getPathAndVerbFromJson(tag string, operationID string) (path string, verb string) {
  glog.V(1).Info(operationID)
  context, verb := getContextAndVerb(removeQuotes(string(operationID)), removeQuotes(string(tag))) 
  path = "/" + removeQuotes(string(tag)) + context
  return path, verb
}

func getContextAndVerb(operationID string, tag string) (context string, verb string) {
  
  switch true {
    case strings.HasPrefix(operationID, "add"):
      context = strings.TrimPrefix(operationID, "add")
      verb = "post"
    case strings.HasPrefix(operationID, "post"):
      context = strings.TrimPrefix(operationID, "post")
      verb = "post"
    case strings.HasPrefix(context, "upload"):
      context = strings.TrimPrefix(operationID, "upload")
      verb = "post"
    case strings.HasPrefix(operationID, "update"):
      context = strings.TrimPrefix(operationID, "update")
      verb = "put"
    case strings.HasPrefix(operationID, "get"):
      context = strings.TrimPrefix(operationID, "get")
      verb = "get"
    case strings.HasPrefix(operationID, "find"):
      context = operationID
      verb = "get"
    }
    glog.V(1).Info("Verb is : " + verb)
    glog.V(1).Info("Tag is : " + tag)

    context = strings.Replace(context, strings.Title(tag), "", 1)
    glog.V(1).Info("Context is : " + context)
    if context != "" {
      context = "/" + context
    }

  return context, verb  
}

func getSwaggerTemplate() SwaggerTemplate {
  raw, err := ioutil.ReadFile("./Swagger_UI_Poc.json")
  if err != nil {
    glog.Fatal(err.Error())
  }

  var c SwaggerTemplate
  json.Unmarshal(raw, &c)
  return c
}

func removeQuotes(quotedString string) (unquotedString string) {
  unquotedString = strings.TrimPrefix(quotedString, "\"")
  return strings.TrimSuffix(unquotedString, "\"")
}
