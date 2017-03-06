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

func isJsonArray(s string) ([]map[string]interface{}, bool) {
  var js []map[string]interface{}

  err := json.Unmarshal([]byte(s), &js)
  if err != nil {
    glog.V(1).Info("Json is not valid: " + err.Error())
  }

  return js, err == nil 
}

func jsonStringToMap(s string) map[string]interface{} {
  var js map[string]interface{}

  err := json.Unmarshal([]byte(s), &js)
  if err != nil {
    glog.V(1).Info("Json is not valid: " + err.Error())
  }
  
  return js
}

func isJson(s string) bool {
  var js interface{}

  err := json.Unmarshal([]byte(s), &js)
  if err != nil {
    glog.V(1).Info("Json is not valid: " + err.Error())
  }

  return err == nil
}

func (def *DefinitionStruct) MarshalJSON() ([]byte, error) {
  var jsonBytes []byte = []byte("{")
  comma := []byte(",")
  
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

    properties = strings.TrimPrefix(properties, "{")
    jsonBytes = append(jsonBytes, []byte(properties)...)
    jsonBytes = bytes.TrimSuffix(jsonBytes, []byte("}"))
    jsonBytes = append(jsonBytes, comma...)
  }

  jsonBytes = bytes.TrimSuffix(jsonBytes, comma)
  jsonBytes = append(jsonBytes, []byte("}")...)

  glog.V(2).Info("json from definiton structs is " +  string(jsonBytes))

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

/**
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
}**/

func (ver *VerbStruct) MarshalJSON() ([]byte, error) {
  var jsonBytes []byte
  comma := []byte(",")

  jsonBytes = append(jsonBytes, []byte("{")...)

  //Tags
  tags, err := json.Marshal(ver.Tags)
  if err != nil {
    glog.Error(err)
  }
  jsonBytes = append(jsonBytes, []byte("\"tags\":")...)
  jsonBytes = append(jsonBytes, tags...)
  jsonBytes = append(jsonBytes, comma...)

  //Summary
  summary, err := json.Marshal(ver.Summary)
  if err != nil {
    glog.Error(err)
  }
  jsonBytes = append(jsonBytes, []byte("\"summary\":")...)
  jsonBytes = append(jsonBytes, summary...)
  jsonBytes = append(jsonBytes, comma...)

  //Description
  description, err := json.Marshal(ver.Description)
  if err != nil {
    glog.Error(err)
  }
  jsonBytes = append(jsonBytes, []byte("\"description\":")...)
  jsonBytes = append(jsonBytes, description...)
  jsonBytes = append(jsonBytes, comma...)

  //OperationID
  operationId, err := json.Marshal(ver.OperationID)
  if err != nil {
    glog.Error(err)
  }
  jsonBytes = append(jsonBytes, []byte("\"operationId\":")...)
  jsonBytes = append(jsonBytes, operationId...)
  jsonBytes = append(jsonBytes, comma...)

  //Consumes
  consumes, err := json.Marshal(ver.Consumes)
  if err != nil {
    glog.Error(err)
  }
  jsonBytes = append(jsonBytes, []byte("\"consumes\":")...)
  jsonBytes = append(jsonBytes, consumes...)
  jsonBytes = append(jsonBytes, comma...)

  //Produces
  produces, err := json.Marshal(ver.Produces)
  if err != nil {
    glog.Error(err)
  }
  jsonBytes = append(jsonBytes, []byte("\"produces\":")...)
  jsonBytes = append(jsonBytes, produces...)
  jsonBytes = append(jsonBytes, comma...)

  //Connection
  connection, err := json.Marshal(ver.Connection)
  if err != nil {
    glog.Error(err)
  }
  jsonBytes = append(jsonBytes, []byte("\"connection\":")...)
  jsonBytes = append(jsonBytes, connection...)
  jsonBytes = append(jsonBytes, comma...)

  //Parameters
  jsonBytes = append(jsonBytes, []byte("\"parameters\":")...)
  parameters, err := json.Marshal(ver.Parameters)
  if err != nil {
    glog.Error(err)
  }

  parameters = bytes.Replace(parameters, []byte("\\\""), []byte("\""), -1)
  parameters = []byte(removeQuotes(string(parameters)))

  if string(parameters) == "" {
    jsonBytes = append(jsonBytes, []byte("[]")...)
  } else {
    jsonBytes = append(jsonBytes, parameters...)
  }
  //jsonBytes = append(jsonBytes, []byte("]")...)

  jsonBytes = append(jsonBytes, []byte("}")...)
  glog.V(2).Info("json bytes for verb struct: %s", string(jsonBytes))
  return jsonBytes, nil
}

func (ver *VerbStruct) UnmarshalJSON(b []byte) (err error) {
  s := strings.Trim(string(b), "\"")
  if s == "null" {
    glog.Warning("bytes are null, something isn't right")
  }

  glog.V(2).Info(s)
  var verb map[string]json.RawMessage
  json.Unmarshal(b, &verb)

  var tags []string
  json.Unmarshal(verb["tags"], &tags)
  //ver.Tags = removeQuotes(string(verb["tags"]))
  ver.Tags = tags
  ver.Summary = removeQuotes(string(verb["summary"]))
  ver.Description = removeQuotes(string(verb["description"]))
  ver.OperationID = removeQuotes(string(verb["operationId"]))

  var consumes []string
  json.Unmarshal(verb["consumes"], &consumes)
  //ver.Consumes = removeQuotes(string(verb["consumes"]))
  ver.Consumes = consumes

  var produces []string
  json.Unmarshal(verb["produces"], &produces)
  //ver.Produces = removeQuotes(string(verb["produces"]))
  ver.Produces = produces

  var conn ConnectionStruct
  json.Unmarshal(verb["connection"], &conn)

  ver.Connection = conn

  glog.V(1).Info("Parameters from bytes in unmarshal: " + string(verb["parameters"]))

  params := removeWhiteSpace(string(verb["parameters"]))

  params = removeQuotes(params)

  params = strings.TrimPrefix(params, "[")

  ver.Parameters = strings.TrimSuffix(params, "]")

  glog.V(1).Info("Parameters from verbStruct in unmarshal: " + ver.Parameters)

  return
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
    case strings.HasPrefix(operationID, "upload"):
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
    case strings.HasPrefix(operationID, "delete"):
      context = strings.TrimPrefix(operationID, "delete")
      verb = "delete"
    case strings.HasPrefix(strings.ToLower(operationID), "create" + tag + "with"):
      context = operationID
      if strings.HasSuffix(context, "input") {
        context = strings.TrimSuffix(context, "input")
      }
      
      if strings.HasSuffix(context, "Input") {
        context = strings.TrimSuffix(context, "Input")
      }
        
      verb = "post"
    case strings.HasPrefix(operationID, "create"):
      context = strings.TrimPrefix(operationID, "create")
      verb = "post"
    }
    glog.V(1).Info("Verb is : " + verb)
    glog.V(1).Info("Tag is : " + tag)

    context = strings.Replace(context, strings.Title(tag), "", 1)
    glog.V(1).Info("Context is : " + context)
    if context != "" {
      context = "/" + context
    }

    if strings.Contains(context, "//") {
      context = strings.Replace(context, "//", "/", -1)
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

func removeWhiteSpace(stringWithSpaces string) (stringWithoutSpaces string) {
  stringWithoutSpaces = strings.Replace(stringWithSpaces, "\n", "", -1)
  stringWithoutSpaces = strings.Replace(stringWithoutSpaces, "\r", "", -1)
  stringWithoutSpaces = strings.Replace(stringWithoutSpaces, "\t", "", -1)
  return stringWithoutSpaces
}

