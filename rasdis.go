package main

import (
  "strings"
  "github.com/golang/glog"
  "fmt"
  "errors"
  "encoding/json"
)

func dealWithTags(template *SwaggerTemplate, contentPolicy contentPolicy, policyType string) error  {
  if strings.Contains(contentPolicy.Name, "_" + policyType + "_content_policy_") {
    tags := template.Tags

    tag := new(TagStruct)

    tag.Name = strings.Split(contentPolicy.Name, "_policy_")[1]
    tag.Description = contentPolicy.Description

    tags = append(tags, *tag)

    template.Tags = tags
    return nil
  } else {
    glog.Warning("The name of content policy " + contentPolicy.Name + " is not fomatted correctly for service discovery, please update the content policy name to include it in your swagger UI")
    return errors.New("Content policy format error")
  }
}

func dealWithContentPolicyList(template *SwaggerTemplate, policyList policyList, policyType string, config Config) {
  //paths := template.Paths
  
  for i := range policyList.Policy {
    contentPolicy := getContentPolicy(policyList.Policy[i].Name, "json", config)
    dealWithContentPolicy(template, contentPolicy, policyType, config)
    glog.Flush()
  }
}

func dealWithContentPolicy(template *SwaggerTemplate, contentPolicy contentPolicy, policyType string, config Config) {
  err := dealWithTags(template, contentPolicy, policyType)

  if err == nil {
    virtualDirectoryList := getVirtualDirectoryList(contentPolicy.Name, "json", config)
      
    dealWithVirtualDirectoryList(template, &contentPolicy, virtualDirectoryList, policyType, config)
  }
}

func dealWithVirtualDirectoryList(template *SwaggerTemplate, contentPolicy *contentPolicy, virtualDirectoryList policyList, policyType string, config Config) {
 for j := range virtualDirectoryList.Policy {

    virtualDirectory := getVirtualDirectory(contentPolicy.Name, virtualDirectoryList.Policy[j].Name, "json", config)
  
    dealWithVirtualDirectory(template, contentPolicy, &virtualDirectory, policyType, config)
    glog.Flush()
  }
}

func dealWithVirtualDirectory(template *SwaggerTemplate, contentPolicy *contentPolicy, virtualDirectory *virtualDirectory, policyType string, config Config) {
  path := new(PathStruct)
  verb := new(VerbStruct)
  definition := new(Definition)
  
  listenerPolicy := getListenerPolicy(getNetworkProtocol(virtualDirectory.ListenerPolicy), virtualDirectory.ListenerPolicy, config)

  contentPolicySplit := strings.Split(contentPolicy.Name, "_" + policyType + "_content_policy_") 

  //grab any documents for this verb
  document, parameterName, docExists := getDocument(contentPolicySplit[0], contentPolicySplit[1], config)

    glog.V(1).Info(docExists)
    if docExists {
      definition.Type = "object"
      definition.Properties = document
      definition.Xml.Name = contentPolicySplit[1]
      template.Definitions.definitions = append(template.Definitions.definitions, *definition)

      parameters := new(VerbParameters)

      parameters.In = "body"
      parameters.Name = parameterName
      parameters.Description = ""

      listenerPort,err := json.Marshal(listenerPolicy["port"])
      if err != nil {
        glog.Warning("Unable to parse port from listener policy setting to 80")
        listenerPort = []byte("80")
      }
      fmt.Println("Port from listenerPolicy is : " + string(listenerPort))
      parameters.Port = string(listenerPort)
      parameters.Required = false
      parameters.Schema.Ref = "#/definitions/" + parameterName 

      verb.Parameters = append(verb.Parameters, * parameters)
    } else {
      parameters := new(VerbParameters)
      parameters.Port = "20001"

      verb.Parameters = append(verb.Parameters, * parameters)
    } 
    
    verb.Tags = append(verb.Tags, contentPolicySplit[1])
    verb.Summary, verb.OperationID = processVirtualDirectory(virtualDirectory.VirtualPath, (strings.Split(contentPolicy.Name, "_policy_")[1]), virtualDirectory.Description)

  if verb.Summary != "" && verb.OperationID != "" {
    verb.Description = virtualDirectory.Description
    
    //Set produces and consumes based on policy type
    switch policyType {
    case "json":
      verb.Consumes = []string{"application/json"}
      verb.Produces = []string{"application/json"}
    case "xml":
      verb.Consumes = []string{"application/xml"}
      verb.Produces = []string{"application/xml"}
    }

    glog.V(2).Info(verb)
    path.verbs = append(path.verbs, *verb)
  
    template.Paths.paths = append(template.Paths.paths, *path)
  }
}

func rasdis(user string) string {

  fmt.Println("Authorised user is: " + user)
  template := getSwaggerTemplate()
  policyType := "json"

  config := ReadConfig("rasdis.cfg")

  glog.V(1).Info("forum url from config is: " + config.ForumURL)

  if config.ForumURL == "" {
    glog.Fatal("Forum URL is empty please populate or check your config is valid JSON")
  }
  policyList := getContentPolicyList(policyType, config)  
  
  dealWithContentPolicyList(&template, policyList, policyType, config)

  glog.V(2).Info(template)
  glog.V(2).Info(toJson(&template))

  glog.Flush()

  return toJson(&template)
}

func getNetworkProtocol(networkName string) string {
  switch true {
  case strings.Contains(networkName, "_http_"):
   return "http"
  case strings.Contains(networkName, "_amqp10_"):
   return "amqp10"
  }
  return ""
}

func getVerbs(requestFilter string) []string {
  verbs := make([]string, 1)

  switch true {
  case strings.Contains(requestFilter, "POST"):
    verbs = append(verbs, "POST")
  case strings.Contains(requestFilter, "GET"):
    verbs = append(verbs, "GET")
  case strings.Contains(requestFilter, "PUT"):
    verbs = append(verbs, "PUT")
  }

  return verbs
}

func processVirtualDirectory(virtualPath string, tag string, virtualDescription string) (summary string, operationId string)  {

  var context string
  if strings.HasPrefix(virtualPath, "/" + tag + "/") {
    context = strings.TrimPrefix(virtualPath, "/" + tag + "/")
  }

  if context != "" {
    switch true {
    case strings.HasPrefix(context, "add"): 
      operationId = strings.Replace(context, "add", "add" + strings.Title(tag), 1)
      summary = "Add a new " + tag
    case strings.HasPrefix(context, "post"):
      operationId = strings.Replace(context, "post", "post" + strings.Title(tag), 1)
      summary = "Post a " + tag
    case strings.HasPrefix(context, "upload"):
      operationId = strings.Replace(context, "upload", "upload" + strings.Title(tag), 1)
      summary = "Upload a " + tag
    case strings.HasPrefix(context, "update"):
      operationId = strings.Replace(context, "update", "update" + strings.Title(tag), 1)
      summary = "Update an existing " + tag
    case strings.HasPrefix(context, "get"): 
      operationId = strings.Replace(context, "get", "get" + strings.Title(tag), 1)
      summary = "Get " + tag
    case strings.HasPrefix(context, "find"):
      operationId = strings.Replace(context, "find", "find" + strings.Title(tag), 1)
      var by = ""
      if strings.Contains(context, "By") {
        by = (strings.Split(context, "By"))[1]
      } else {
        by = (strings.Split(context, "by"))[1]
      }
      summary = "Finds " + tag + " by " + by
    default:
      glog.Warningf("The virtual path %s doesn't contain a currently supported action", virtualPath)
      return "", ""
    }
  } else {
    switch true {
    case strings.HasPrefix(strings.ToLower(virtualDescription), "add"):
      operationId = "add" + strings.Title(tag)
      summary = "Add a " + tag
    case strings.HasPrefix(strings.ToLower(virtualDescription), "post"):
      operationId = "post" + strings.Title(tag)
      summary = "Post a " + tag
    case strings.HasPrefix(strings.ToLower(virtualDescription), "upload"):
      operationId = "upload" + strings.Title(tag)
      summary = "Upload a " + tag
    case strings.HasPrefix(strings.ToLower(virtualDescription), "update"):
      operationId = "update" + strings.Title(tag)
      summary = "Update an existing " + tag
    case strings.HasPrefix(strings.ToLower(virtualDescription), "get"):
      operationId = "get" + strings.Title(tag)
      summary = "Get " + tag
    case strings.HasPrefix(strings.ToLower(virtualDescription), "find"):
      by := strings.Split(virtualDescription, "by")
      operationId = "find" + strings.Title(tag) + "By" + by[1]
      summary = "Finds " + tag + " by " + by[1]
    default:
      glog.Warningf("Unable to determine action from description of virtual directory with virtual path of %s. Please add supported action to virtual directories description or virtual path", virtualPath)
      return "", ""
    }
  }
  return summary, operationId
}

func getVerb(virtualPath string, tag string, virtualDescription string) (verb string) {
  var context string
  if strings.HasPrefix(virtualPath, tag) {
    context = strings.TrimPrefix(virtualPath, "/" + tag + "/")
  }

  if context != "" {
    switch true {
    case strings.HasPrefix(context, "add"):
      verb = "post"
    case strings.HasPrefix(context, "post"):
      verb = "post"
    case strings.HasPrefix(context, "upload"):
      verb = "post"
    case strings.HasPrefix(context, "update"):
      verb = "put"
    case strings.HasPrefix(context, "get"):
      verb = "get"
    case strings.HasPrefix(context, "find"):
      verb = "get"
    }
  } else {
    switch true {
    case strings.HasPrefix(strings.ToLower(virtualDescription), "add"):
      verb = "post"
    case strings.HasPrefix(strings.ToLower(virtualDescription), "post"):
      verb = "post"
    case strings.HasPrefix(strings.ToLower(virtualDescription), "upload"):
      verb = "post"
    case strings.HasPrefix(strings.ToLower(virtualDescription), "update"):
      verb = "put"
    case strings.HasPrefix(strings.ToLower(virtualDescription), "get"):
      verb = "get"
    case strings.HasPrefix(strings.ToLower(virtualDescription), "find"):
      verb = "get"
    }
  }
  return verb
}
