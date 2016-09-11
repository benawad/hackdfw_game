package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"html/template"
	"net/http"
)

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32),
)

var templates = template.Must(template.ParseFiles("templates/loginForm.html", "templates/dashboard.html", "templates/profile.html"))

func renderTemplate(res http.ResponseWriter, template string, obj interface{}) {
	err := templates.ExecuteTemplate(res, template+".html", obj)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

var router = mux.NewRouter()

func getUsername(req *http.Request) (username string) {
	if cookie, err := req.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			username = cookieValue["name"]
		}
	}
	return username
}

func authenticate(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		username := getUsername(req)
		if username != "" {
			fn(res, req, username)
		} else {
			http.Redirect(res, req, "/", 302)
		}
	}
}

func main() {
	router.HandleFunc("/", indexHandler)
	router.HandleFunc("/dashboard", authenticate(dashboardHandler))
	router.HandleFunc("/profile", authenticate(profileHandler))
	router.HandleFunc("/login", loginHandler).Methods("POST")
	router.HandleFunc("/logout", logoutHandler).Methods("POST")

	http.Handle("/", router)
	http.ListenAndServe(":8000", nil)
}

func indexHandler(res http.ResponseWriter, req *http.Request) {
	renderTemplate(res, "loginForm", nil)
}

func dashboardHandler(res http.ResponseWriter, req *http.Request, username string) {
	user := map[string]string{
		"username": username,
	}
	renderTemplate(res, "dashboard", user)
}

func profileHandler(res http.ResponseWriter, req *http.Request, username string) {
	user := map[string]string{
		"username": username,
	}
	renderTemplate(res, "profile", user)
}

func setSession(username string, res http.ResponseWriter) {
	value := map[string]string{
		"name": username,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(res, cookie)
	}
}

func loginHandler(res http.ResponseWriter, req *http.Request) {
	name := req.FormValue("name")
	password := req.FormValue("password")
	redirectTarget := "/"
	if name != "" && password != "" {

		if name == "bob" && password == "password" {
			setSession(name, res)
			redirectTarget = "/dashboard"
		}

	}

	http.Redirect(res, req, redirectTarget, 302)
}

func clearSession(res http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(res, cookie)
}

func logoutHandler(res http.ResponseWriter, req *http.Request) {
	clearSession(res)
	http.Redirect(res, req, "/", 302)
}
