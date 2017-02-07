package main

import (
  "fmt"
  "net/http"
  "github.com/golang/glog"
  "os/exec"
  "strings"
)

type HasHandleFunc interface { //this is just so it would work for gorilla and http.ServerMux
    HandleFunc(pattern string, handler func(w http.ResponseWriter, req *http.Request))
}
type Handler struct {
    http.HandlerFunc
    Enabled bool
}

type Handlers map[string]*Handler

func (h Handlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path
    if handler, ok := h[path]; ok && handler.Enabled {
        handler.ServeHTTP(w, r)
    } else {
        http.Error(w, "Swagger UI Not Found", http.StatusNotFound)
    }
}

func (h Handlers) HandleFunc(mux HasHandleFunc, pattern string, handler http.HandlerFunc) {
    h[pattern] = &Handler{handler, true}
    mux.HandleFunc(pattern, h.ServeHTTP)
}

//func generateSwagger(fn http.HandlerFunc, mux HasHandleFunc, handlers Handlers, contextMap map[string]string) http.HandlerFunc {
func generateSwagger(fn http.HandlerFunc, contextMap map[string]string) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    uuid, err := exec.Command("uuidgen").Output()
    if err != nil {
      glog.Fatal(err)
    }
    fmt.Printf("%s", uuid)

    fmt.Println("Lauching generator")

    user := r.Header.Get("user")
    swaggerJson := rasdis(user)

    //handlers.HandleFunc(mux, "/ui/" + string(uuid), swaggerGenerator)
    contextMap[strings.TrimRight(string(uuid), "\n")] = swaggerJson

    //fmt.Println(handlers)
    fmt.Println(contextMap)

    //http.Redirect(w, r, "/ui/" + string(uuid), 302)
    http.Redirect(w, r, "http://" + r.Host + "/?url=http://" + r.Host + "/ui/" + string(uuid), 302)
    
    //fn(w, r)
  }
}

func serveSwaggerJson(fn http.HandlerFunc, contextMap map[string]string) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path

    fmt.Println("Path is: " + path + ".")
    fmt.Println("Split is: \"" + strings.Split(path, "/")[2] + "\"")
    fmt.Println(contextMap)
    fmt.Println("Entry is: " + contextMap["\"" + strings.Split(path, "/")[2] + "\""])
 
    if swaggerJson, ok := contextMap[strings.Split(path, "/")[2]]; ok {
      fmt.Fprintf(w, swaggerJson)    
      //http.Redirect(w, r, "http://" + r.Host + "/?url=http://" + r.Host + r.URL.Path, 302)
    } else {
      http.Error(w, "Swagger UI Not Found", http.StatusNotFound)
    }
  }
}

func swaggerGenerator(w http.ResponseWriter, r *http.Request) {
  /**fmt.Fprintf(w, "<h1>Loading Swagger UI</h1>")
  if f, ok := w.(http.Flusher); ok {
    f.Flush()
  } else {
    log.Println("Damn, no flush");
  }**/

  fmt.Println("Lauching generator")

  user := r.Header.Get("user")
  swaggerJson := rasdis(user)

  fmt.Fprintf(w, swaggerJson)
}

func swaggerUI(w http.ResponseWriter, r *http.Request) {
  /**fmt.Fprintf(w, "<h1>Loading Swagger UI</h1>")
  if f, ok := w.(http.Flusher); ok {
     f.Flush()
  } else {
     log.Println("Damn, no flush");
  }**/

}

func startServer() {
  //mux := http.NewServeMux()
  //handlers := Handlers{}
  /**handlers.HandleFunc(mux, "/ui", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("this will show once"))
    handlers["/ui"].Enabled = false
  })**/
  
  contextMap := make(map[string]string)
  
  //http.HandleFunc("/generate", generateSwagger(swaggerGenerator, mux, handlers, contextMap))
  http.HandleFunc("/generate", generateSwagger(swaggerGenerator, contextMap))
  http.HandleFunc("/ui/", serveSwaggerJson(swaggerUI, contextMap))
  fs := http.FileServer(http.Dir("./swagger/dist"))
  http.Handle("/", http.StripPrefix("/", fs))
  http.ListenAndServe(":8080", nil)
  glog.Flush()
}
