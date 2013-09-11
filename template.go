package main

import (
  "html/template"
  "path/filepath"
  "fmt"
  "net/http"
  "thegoods.biz/httpbuf"
)

func reverse(name string, things ...interface{}) string {
  //convert the things to strings
  strs := make([]string, len(things))
  for i, th := range things {
    strs[i] = fmt.Sprint(th)
  }
  //grab the route
  u, err := router.Get(name).URL(strs...)
  if err != nil {
    panic(err)
  }
  return u.Path
}

var funcs = template.FuncMap{
  "reverse": reverse,
}

var cachedTemplates = map[string]*template.Template{}

func T(name string) *template.Template {
  if t, ok := cachedTemplates[name]; ok {
    return t
  }

  t := template.New("_layout.html").Funcs(funcs)

  t = template.Must(t.ParseFiles(
    "templates/_layout.html",
    filepath.Join("templates", name),
  ))

  cachedTemplates[name] = t

  return t
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, *Context) error) http.HandlerFunc {
  return func(w http.ResponseWriter, req *http.Request) {
      //create the context

      ctx, err := NewContext(req)

      if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
      }
      defer ctx.Close()

      //run the handler and grab the error, and report it
      buf := new(httpbuf.Buffer)
      err = fn(buf, req, ctx)
      if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
      }
    
      //save the session
      if err = ctx.Session.Save(req, buf); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
      }
    
      //apply the buffered response to the writer
      buf.Apply(w)
  }
}




