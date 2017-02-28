package main

type SwaggerTemplate struct {
  Swagger string `json:"swagger"`
  Info struct {
    Description string `json:"description"`
    Version string `json:"version"`
    Title string `json:"title"`
    Contact struct {
      Email string `json:"email"`
    } `json:"contact"`
    License struct {
      Name string `json:"name"`
      URL string `json:"url"`
    } `json:"license"`
  } `json:"info"`
  Host string `json:"host"`
  BasePath string `json:"basePath"`
  Tags []TagStruct `json:"tags"`
  Schemes []string `json:"schemes"`
  //Paths map[string]map[string]VerbStruct `json:"paths"`
  Paths PathsStruct `json:"paths"`
  //SecurityDefinitions string `json:"securityDefinitions"`
  Definitions DefinitionStruct `json:"definitions"`
}

type DocumentStruct struct {
  Name string `json:"name"`
  Document map[string]string
}

type Definition struct {
  Type string `json:"type"`
  Properties string `json:"properties"`
  Xml struct {
    Name string `json:"name"`
  } `json:"xml"`
} 

type DefinitionStruct struct {
  definitions []Definition
}

type TagStruct struct {
  Name string `json:"name"`
  Description string `json:"description"`
}

type PathsStruct struct {
  paths []PathStruct `json:paths`
}

type PathStruct struct {
  verbs []VerbStruct `json:path`
}

type VerbStruct struct {
  Tags []string `json:"tags"`
  Summary string `json:"summary"`
  Description string `json:"description"`
  OperationID string `json:"operationId"`
  Consumes []string `json:"consumes"`
  Produces []string `json:"produces"`
  //Port string `json:"port"`
  Connection ConnectionStruct `json:"connection"`
  //Parameters []VerbParameters `json:"parameters"`
  Parameters string `json:"parameters"`
}

type ConnectionStruct struct {
  Port int `json:"port"`
  Type string `json:"type"`
}

type VerbParameters struct {
  In string `json:"in"`
    Name string `json:"name"`
    Description string `json:"description"`
    Required bool `json:"required"`
    Schema struct {
      Ref string `json:"$ref"`
    } `json:"schema"`
}

type policyList struct {
  Policy []struct {
    Name string `json:"name"`
    URL string `json:"url"`
  } `json:"policy"`
}

type contentPolicy struct {
  Description string `json:"description"`
  Name string `json:"name"`
  IdpGroup string `json:"idpGroup"`
  RequestProcess string `json:"requestProcess"`
  RequestProcessType string `json:"requestProcessType"`
  ResponseProcess string `json:"responseProcess"`
  ResponseProcessType string `json:"responseProcessType"`
}

type virtualDirectory struct {
  Description string `json:"description"`
  AclPolicy string `json:"aclPolicy"`
  Name string `json:"name"`
  Enabled bool `json:"enabled"`
  ErrorTemplate string `json:"errorTemplate"`
  ListenerPolicy string `json:"listenerPolicy"`
  VirtualPath string `json:"virtualPath"`
  RemotePath string `json:"remotePath"`
  RemotePolicy string `json:"remotePolicy"`
  RequestFilterPolicy string `json:"requestFilterPolicy"`
  RequestProcess string `json:"requestProcess"`
  RequestProcessType string `json:"requestProcessType"`
  ResponseProcess string `json:"responseProcess"`
  ResponseProcessType string `json:"responseProcessType"`
  UseRemotePolicy bool `json:"useRemotePolicy"`
  VirtualHost string `json:"virtualHost"`
}

type httpListener struct {
  Name string `json:"name"`
  Enabled bool `json:"enabled"`
  ErrorTemplate string `json:"errorTemplate"`
  Interface string `json:"interface"`
  Port int `json:"port"`
  ReadTimeoutMillis int `json:"readTimeoutMillis"`
  UseDeviceIP bool `json:"useDeviceIp"`
  IPAclPolicy string `json:"ipAclPolicy"`
  ListenerSSLEnabled bool `json:"listenerSSLEnabled"`
  ListenerSSLPolicy string `json:"listenerSSLPolicy"`
  PasswordAuthenticationRealm string `json:"passwordAuthenticationRealm"`
  RequirePasswordAuthentication bool `json:"requirePasswordAuthentication"`
  UseBasicAuthentication bool `json:"useBasicAuthentication"`
  UseChunking bool `json:"useChunking"`
  UseCookieAuthentication bool `json:"useCookieAuthentication"`
  UseDigestAuthentication bool `json:"useDigestAuthentication"`
  UseFormPostAuthentication bool `json:"useFormPostAuthentication"`
  UseKerberosAuthentication bool `json:"useKerberosAuthentication"`
}
type rabbitMqRemote struct {
  Name string `json:"name"`
  Description string `json:"description"`
  Enabled bool `json:"enabled"`
}
