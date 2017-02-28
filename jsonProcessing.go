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

  err := json.Unmarshal([]byte(s), &js)
  
  return js, err == nil
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
      glog.V(2).Info("Marshalling a verb now")
      var verb VerbStruct
      verb = pss.paths[i].verbs[j]
      bytes, err := json.Marshal(&verb)
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

func (ver *VerbStruct) MarshalJSON() ([]byte, error) {
  glog.V(2).Info("Inside marshal for verb struct")
  var jsonBytes []byte = []byte("")

  tag, err := json.Marshal(ver.Tags)
  summary, err := json.Marshal(ver.Summary)
  description, err := json.Marshal(ver.Description)
  operationId, err := json.Marshal(ver.OperationID)
  consumes, err := json.Marshal(ver.Consumes)
  produces, err := json.Marshal(ver.Produces)
  connection, err := json.Marshal(ver.Connection)
  parameters, err := json.Marshal(ver.Parameters)

  if bytes.HasPrefix(parameters, []byte("\"")) {
    parameters = bytes.TrimPrefix(parameters, []byte("\""))
  }
  if bytes.HasSuffix(parameters, []byte("\"")) {
    parameters = bytes.TrimSuffix(parameters, []byte("\""))
  }

  parameters = bytes.Replace(parameters, []byte("\\\""), []byte("\""), -1)

  if err != nil {
    glog.Error(err)
    return []byte(""), err
  }

  jsonBytes = append(jsonBytes, []byte("{\"tags\":")...)
  jsonBytes = append(jsonBytes, tag...)
  jsonBytes = append(jsonBytes, []byte(",\"summary\":")...)
  jsonBytes = append(jsonBytes, summary...)
  jsonBytes = append(jsonBytes, []byte(",\"description\":")...)
  jsonBytes = append(jsonBytes, description...)
  jsonBytes = append(jsonBytes, []byte(",\"operatonId\":")...)
  jsonBytes = append(jsonBytes, operationId...)
  jsonBytes = append(jsonBytes, []byte(",\"consumes\":")...)
  jsonBytes = append(jsonBytes, consumes...)
  jsonBytes = append(jsonBytes, []byte(",\"produces\":")...)
  jsonBytes = append(jsonBytes, produces...)
  jsonBytes = append(jsonBytes, []byte(",\"connection\":")...)
  jsonBytes = append(jsonBytes, connection...)
  jsonBytes = append(jsonBytes, []byte(",\"parameters\":[")...)
  if string(parameters) != "" {
    jsonBytes = append(jsonBytes, parameters...)
    jsonBytes = append(jsonBytes, []byte("]}")...)
  } else {
    jsonBytes = append(jsonBytes, []byte("]}")...)
  }

  glog.V(2).Info("verb jsonBytes is: " + string(jsonBytes))

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
