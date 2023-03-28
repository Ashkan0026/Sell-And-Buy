package utils

import (
	"log"
	"net/http"
	"text/template"
)

func ExecFile(name string, w http.ResponseWriter, data interface{}) {
	t, err := template.ParseFiles("./views/"+name+".html", "./views/temps/header.html", "./views/temps/style.html")
	if err != nil {
		log.Printf("%v", err)
		return
	}
	t.Execute(w, data)
}
