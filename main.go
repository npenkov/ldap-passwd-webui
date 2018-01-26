package main

import (
	"fmt"
	"github.com/npenkov/ldap-passwd-webui/app"
	"net/http"
)

func main() {
	reHandler := new(app.RegexpHandler)

	reHandler.HandleFunc(".*.[js|css|png|eof|svg|ttf|woff]", "GET", app.ServeAssets)
	reHandler.HandleFunc("/", "GET", app.ServeIndex)
	reHandler.HandleFunc("/", "POST", app.ChangePassword)

	fmt.Println("Starting server on port 8080")
	http.ListenAndServe(":8080", reHandler)
}
