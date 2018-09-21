package main

import (
	"fmt"
	"net/http"

	"github.com/dchest/captcha"
	"github.com/npenkov/ldap-passwd-webui/app"
)

func main() {
	reHandler := new(app.RegexpHandler)

	reHandler.HandleFunc(".*.[js|css|png|eof|svg|ttf|woff]", "GET", app.ServeAssets)
	reHandler.HandleFunc("/", "GET", app.ServeIndex)
	reHandler.HandleFunc("/", "POST", app.ChangePassword)
	http.Handle("/captcha/", captcha.Server(captcha.StdWidth, captcha.StdHeight))
	http.Handle("/", reHandler)
	fmt.Println("Starting server on port 8080")
	http.ListenAndServe(":8080", nil)
}
