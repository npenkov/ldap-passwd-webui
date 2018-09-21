package app

import (
	"fmt"
	"log"
	"path"
	"strings"

	"html/template"

	"github.com/dchest/captcha"

	"regexp"

	"net/http"
)

type route struct {
	pattern *regexp.Regexp
	verb    string
	handler http.Handler
}

// RegexpHandler is used for http handler to bind using regular expressions
type RegexpHandler struct {
	routes []*route
}

// Handler binds http handler on RegexpHandler
func (h *RegexpHandler) Handler(pattern *regexp.Regexp, verb string, handler http.Handler) {
	h.routes = append(h.routes, &route{pattern, verb, handler})
}

// HandleFunc binds http handler function on RegexpHandler
func (h *RegexpHandler) HandleFunc(r string, v string, handler func(http.ResponseWriter, *http.Request)) {
	re := regexp.MustCompile(r)
	h.routes = append(h.routes, &route{re, v, http.HandlerFunc(handler)})
}

func (h *RegexpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range h.routes {
		if route.pattern.MatchString(r.URL.Path) && route.verb == r.Method {
			route.handler.ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}

type pageData struct {
	Title       string
	Pattern     string
	PatternInfo string
	Username    string
	Alerts      map[string]string
	CaptchaId   string
}

// ServeAssets : Serves the static assets
func ServeAssets(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, path.Join("static", req.URL.Path[1:]))
}

// ServeIndex : Serves index page on GET request
func ServeIndex(w http.ResponseWriter, req *http.Request) {
	p := &pageData{Title: getTitle(), CaptchaId: captcha.New(), Pattern: getPattern(), PatternInfo: getPatternInfo()}
	t, e := template.ParseFiles(path.Join("templates", "index.html"))
	if e != nil {
		log.Printf("Error parsing file %v\n", e)
	} else {
		t.Execute(w, p)
	}
}

// ChangePassword : Serves index page on POST request - executes the change
func ChangePassword(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	un := ""
	username := req.Form["username"]
	oldPassword := req.Form["old-password"]
	newPassword := req.Form["new-password"]
	confirmPassword := req.Form["confirm-password"]
	captchaID := req.Form["captchaId"]
	captchaSolution := req.Form["captchaSolution"]

	alerts := map[string]string{}

	if len(username) < 1 || username[0] == "" {
		alerts["error"] = "Username not specified."
	} else {
		un = username[0]
	}
	if len(oldPassword) < 1 || oldPassword[0] == "" {
		alerts["error"] = alerts["error"] + "Old password not specified."
	}
	if len(newPassword) < 1 || newPassword[0] == "" {
		alerts["error"] = alerts["error"] + "New password not specified."
	}
	if len(confirmPassword) < 1 || confirmPassword[0] == "" {
		alerts["error"] = alerts["error"] + "Confirmation password not specified."
	}

	if len(confirmPassword) >= 1 && len(newPassword) >= 1 && strings.Compare(newPassword[0], confirmPassword[0]) != 0 {
		alerts["error"] = alerts["error"] + "New and confirmation passwords does not match."
	}

	if m, _ := regexp.MatchString(getPattern(), newPassword[0]); !m {
		alerts["error"] = alerts["error"] + fmt.Sprintf("%s", getPatternInfo())
	}

	if len(captchaID) < 1 || captchaID[0] == "" ||
		len(captchaSolution) < 1 || captchaSolution[0] == "" ||
		!captcha.VerifyString(captchaID[0], captchaSolution[0]) {
		alerts["error"] = "Wrong captcha."
	}

	if len(alerts) == 0 {
		client := NewLDAPClient()
		if err := client.ModifyPassword(un, oldPassword[0], newPassword[0]); err != nil {
			alerts["error"] = fmt.Sprintf("%v", err)
		} else {
			alerts["success"] = "Password successfuly changed"
		}
	}

	p := &pageData{Title: getTitle(), Alerts: alerts, Username: un, CaptchaId: captcha.New()}

	t, e := template.ParseFiles(path.Join("templates", "index.html"))
	if e != nil {
		log.Printf("Error parsing file %v\n", e)
	} else {
		t.Execute(w, p)
	}
}
