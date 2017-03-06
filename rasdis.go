package main

import (
  "strings"
  "github.com/golang/glog"
  "fmt"
  "errors"
  "strconv"
)

var config Config
var template SwaggerTemplate

func dealWithTags(contentPolicy contentPolicy, policyType string) error  {
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

func dealWithContentPolicyList(policyList policyList, policyType string) {
  //paths := template.Paths
  
  for i := range policyList.Policy {
    contentPolicy := getContentPolicy(policyList.Policy[i].Name, "json", config)
    dealWithContentPolicy(contentPolicy, policyType)
    glog.Flush()
  }
}

func dealWithContentPolicy(contentPolicy contentPolicy, policyType string) {
  err := dealWithTags(contentPolicy, policyType)

  if err == nil {
    virtualDirectoryList := getVirtualDirectoryList(contentPolicy.Name, "json", config)
      
    dealWithVirtualDirectoryList(&contentPolicy, virtualDirectoryList, policyType)
  }
}

func dealWithVirtualDirectoryList(contentPolicy *contentPolicy, virtualDirectoryList policyList, policyType string) {
 for j := range virtualDirectoryList.Policy {

    virtualDirectory := getVirtualDirectory(contentPolicy.Name, virtualDirectoryList.Policy[j].Name, "json", config)
  
    dealWithVirtualDirectory(contentPolicy, &virtualDirectory, policyType)
    glog.Flush()
  }
}

func dealWithVirtualDirectory(contentPolicy *contentPolicy, virtualDirectory *virtualDirectory, policyType string) {
  path := new(PathStruct)
  verb := new(VerbStruct)
  
  listenerType := getNetworkProtocol(virtualDirectory.ListenerPolicy)
  listenerPolicy := getListenerPolicy(listenerType, virtualDirectory.ListenerPolicy, config)

  contentPolicySplit := strings.Split(contentPolicy.Name, "_" + policyType + "_content_policy_") 
  virtualDirectorySplit := strings.Split(virtualDirectory.Name, "_virtual_directory_") 
  parameterDocument, parameterDocExists := getDocument(virtualDirectorySplit[0], virtualDirectorySplit[1], "parameters", config)

  var pathVariables []string
  if parameterDocExists {
    glog.V(1).Info("Parameter doc is: " + parameterDocument)
    verb.Parameters = parameterDocument 

    checkForDefinitions(virtualDirectorySplit[0], parameterDocument)
    pathVariables = checkForPathVariables(parameterDocument)
    for _,path := range pathVariables {
      glog.V(1).Info("Path variable is: " + path)
      glog.V(1).Info("Virtual Path is: " + virtualDirectory.VirtualPath)
      path = removeQuotes(path)
      if strings.HasSuffix(virtualDirectory.VirtualPath, "/") {
        virtualDirectory.VirtualPath = virtualDirectory.VirtualPath + "{" + path + "}"
      } else {
        virtualDirectory.VirtualPath = virtualDirectory.VirtualPath + "/{" + path + "}"
      }
    }
  } 
    
    verb.Tags = append(verb.Tags, contentPolicySplit[1])
    verb.Summary, verb.OperationID = processVirtualDirectory(virtualDirectorySplit[1], virtualDirectory.VirtualPath, contentPolicySplit[1], virtualDirectory.Description)

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
   
    listenerPort := listenerPolicy["port"]
    fmt.Println("Port from listenerPolicy is : " + string(listenerPort))
    port, err := strconv.Atoi(removeQuotes(string(listenerPort)))
    if err != nil {
      glog.Warning("Unable to convert port string " + removeQuotes(string(listenerPort)) + " to a number")
    }
    verb.Connection.Port = port
    verb.Connection.Type = listenerType

    glog.V(2).Info(verb)
    path.verbs = append(path.verbs, *verb)
  
    template.Paths.paths = append(template.Paths.paths, *path)
  }
}

func checkForPathVariables(parameterDocument string) (pathVariables []string) {
  if strings.Contains(parameterDocument, "\"in\": \"path\"") {
    pathVariables = make([]string, 0)
    if jsonDoc, ok := isJsonArray(parameterDocument); ok {
      for _,jsonEntry := range jsonDoc {
        if jsonEntry["in"] == "path" {
          glog.Info("Adding path variable: " + jsonEntry["name"].(string))
          pathVariables = append(pathVariables, jsonEntry["name"].(string))
        }
      }
    }
    return pathVariables
  }
  return nil
}

func checkForDefinitions(projectName string, checkInHere string) {
  if strings.Contains(checkInHere, "#/definitions/") {
    checkInHere = strings.SplitAfterN(checkInHere, "#/definitions/", 2)[1]
    checkForDefinitions(projectName, checkInHere)
    definitionName := strings.SplitAfterN(checkInHere, "\"", 2)[0]
    definitionName = removeQuotes(definitionName)
    glog.Info("Found a definition: #/definitions/" + definitionName)
    grabDefinitionDoc(projectName, definitionName)
  }
}

func grabDefinitionDoc(projectName string, definitionName string) {
  definition := new(Definition)

  definitionDocument, docExists := getDocument(projectName, definitionName, "definition", config)

  glog.V(1).Info("Definition document from forum is: " + definitionDocument)
  glog.V(1).Info(docExists)
  if docExists {
    definition.Type = "object"
    definition.Properties = definitionDocument
    definition.Xml.Name = definitionName

    alreadyExists := false
    for _,def := range template.Definitions.definitions {
      if def == *definition {
        alreadyExists = true
      }
    }

    if !alreadyExists {
      template.Definitions.definitions = append(template.Definitions.definitions, *definition)
    }

    checkForDefinitions(projectName, definitionDocument)
  }
}

func rasdis(user string) string {

  fmt.Println("Authorised user is: " + user)
  template = getSwaggerTemplate()
  policyType := "json"

  config = ReadConfig("rasdis.cfg")

  glog.V(1).Info("forum url from config is: " + config.ForumURL)

  if config.ForumURL == "" {
    glog.Fatal("Forum URL is empty please populate or check your config is valid JSON")
  }
  policyList := getContentPolicyList(policyType, config)  
  
  dealWithContentPolicyList(policyList, policyType)

  glog.V(2).Info(template)
  glog.V(2).Info(swaggerToJson(&template))

  glog.Flush()

  return swaggerToJson(&template)
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

func processVirtualDirectory(virtualName string, virtualPath string, tag string, virtualDescription string) (summary string, operationId string)  {

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
    case strings.HasPrefix(context, "delete"):
      operationId = strings.Replace(context, "delete", "delete" + strings.Title(tag), 1)
      summary = "Delete a " + tag
    case strings.HasPrefix(context, "create"):
      operationId = strings.Replace(context, "create", "create" + strings.Title(tag), 1)
      summary = "Create a " + tag
    default:
      glog.Warningf("The virtual path %s doesn't contain a currently supported action. Trying to get the info from the virtual directory name.", virtualPath)
      switch true {
      case strings.HasPrefix(strings.ToLower(virtualName), "add"):
        operationId = "add" + strings.Title(tag)
        summary = "Add a " + tag
      case strings.HasPrefix(strings.ToLower(virtualName), "post"):
        operationId = "post" + strings.Title(tag)
        summary = "Post a " + tag
      case strings.HasPrefix(strings.ToLower(virtualName), "upload"):
        operationId = "upload" + strings.Title(tag)
        summary = "Upload a " + tag
      case strings.HasPrefix(strings.ToLower(virtualName), "update"):
        operationId = "update" + strings.Title(tag)
        summary = "Update an existing " + tag
      case strings.HasPrefix(strings.ToLower(virtualName), "get"):
        operationId = "get" + strings.Title(tag)
        summary = "Get " + tag
      case strings.HasPrefix(strings.ToLower(virtualName), "find"):
        by := strings.Split(virtualDescription, "by")
        operationId = "find" + strings.Title(tag) + "By" + by[1]
        summary = "Finds " + tag + " by " + by[1]
      case strings.HasPrefix(strings.ToLower(virtualName), "delete"):
        operationId = "delete" + strings.Title(tag)
        summary = "Delete a " + tag
      case strings.HasPrefix(strings.ToLower(virtualName), "create"):
        operationId = "create" + strings.Title(tag)
        summary = "Create a " + tag
      default:
        glog.Warningf("Please add supported action to virtual directories name, description or virtual path. Setting as default")
        if strings.Contains(context, "by") {
          operationId = "get" + tag + context
          summary = "get user by name"
        } else {
          operationId = context
          summary = context
        }
      }
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
    case strings.HasPrefix(strings.ToLower(virtualDescription), "delete"):
      operationId = "delete" + strings.Title(tag)
      summary = "Delete a " + tag
    case strings.HasPrefix(strings.ToLower(virtualDescription), "create"):
      operationId = "create" + strings.Title(tag)
      summary = "Create a " + tag
    default:
      glog.Warningf("Unable to determine action from description of virtual directory with virtual path of %s. Trying to get the info from the virtual directory name.", virtualPath)
      switch true {
      case strings.HasPrefix(strings.ToLower(virtualName), "add"):
        operationId = "add" + strings.Title(tag)
        summary = "Add a " + tag
      case strings.HasPrefix(strings.ToLower(virtualName), "post"):
        operationId = "post" + strings.Title(tag)
        summary = "Post a " + tag
      case strings.HasPrefix(strings.ToLower(virtualName), "upload"):
        operationId = "upload" + strings.Title(tag)
        summary = "Upload a " + tag
      case strings.HasPrefix(strings.ToLower(virtualName), "update"):
        operationId = "update" + strings.Title(tag)
        summary = "Update an existing " + tag
      case strings.HasPrefix(strings.ToLower(virtualName), "get"):
        operationId = "get" + strings.Title(tag)
        summary = "Get " + tag
      case strings.HasPrefix(strings.ToLower(virtualName), "find"):
        by := strings.Split(virtualDescription, "by")
        operationId = "find" + strings.Title(tag) + "By" + by[1]
        summary = "Finds " + tag + " by " + by[1]
      case strings.HasPrefix(strings.ToLower(virtualName), "delete"):
        operationId = "delete" + strings.Title(tag)
        summary = "Delete a " + tag
      case strings.HasPrefix(strings.ToLower(virtualName), "create"):
        operationId = "create" + strings.Title(tag)
        summary = "Create a " + tag
      default:
        glog.Warningf("Please add supported action to virtual directories name, description or virtual path. Setting as default")
        if strings.Contains(context, "by") {
          operationId = "get" + tag + context
          summary = "get user by name"
        } else {
          operationId = context
          summary = context
        }
      }
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
    case strings.HasPrefix(context, "delete"):
      verb = "delete"
    case strings.HasPrefix(context, "create"):
      verb = "create"
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
    case strings.HasPrefix(strings.ToLower(virtualDescription), "delete"):
      verb = "delete"
    case strings.HasPrefix(strings.ToLower(virtualDescription), "create"):
      verb = "create"
    }
  }
  return verb
}
