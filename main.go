package main

import (
	"database/sql"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32),
)

var templates = template.Must(template.ParseFiles("templates/login.html", "templates/dashboard.html", "templates/profile.html", "templates/register.html"))

func renderTemplate(res http.ResponseWriter, template string, obj interface{}) {
	err := templates.ExecuteTemplate(res, template+".html", obj)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

var router = mux.NewRouter()

var db, dbErr = sql.Open("sqlite3", "users.sqlite3")

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

	if dbErr != nil {
		panic(dbErr)
	}

	defer db.Close()

	router.HandleFunc("/", indexHandler)
	router.HandleFunc("/dashboard", authenticate(dashboardHandler))
	router.HandleFunc("/profile", authenticate(profileHandler))
	router.HandleFunc("/login", loginHandler).Methods("GET", "POST")
	router.HandleFunc("/logout", logoutHandler).Methods("POST")
	router.HandleFunc("/register", registerHandler).Methods("GET", "POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	http.Handle("/", router)
	http.ListenAndServe(":"+port, nil)
}

func indexHandler(res http.ResponseWriter, req *http.Request) {
	http.Redirect(res, req, "/login", 302)
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

func verify(username string, password string) bool {
	username = strings.TrimSpace(username)
	rows, err := db.Query("select password from users where username='" + username + "'")
	if err != nil {
		log.Fatal(err)
		return false
	}
	defer rows.Close()
	rows.Next()
	var hashedPassword string
	err = rows.Scan(&hashedPassword)
	if err != nil {
		log.Fatal(err)
		return false
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func loginHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		renderTemplate(res, "login", nil)
	} else {
		name := req.FormValue("name")
		password := req.FormValue("password")
		redirectTarget := "/"
		if name != "" && password != "" {

			if verify(name, password) {
				setSession(name, res)
				redirectTarget = "/dashboard"
			}

		}

		http.Redirect(res, req, redirectTarget, 302)
	}
}

func register(username string, password string) bool {
	username = strings.TrimSpace(username)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
		return false
	}

	// insert into db
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
		return false
	}
	stmt, err := tx.Prepare("insert into users(username, password) values(?, ?)")
	if err != nil {
		log.Fatal(err)
		return false
	}
	defer stmt.Close()
	_, err = stmt.Exec(username, hashedPassword)
	if err != nil {
		log.Fatal(err)
		return false
	}
	tx.Commit()
	return true
}

func registerHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		renderTemplate(res, "register", nil)
	} else {
		name := req.FormValue("name")
		password := req.FormValue("password")
		redirectTarget := "/register"
		if name != "" && password != "" {
			if register(name, password) {
				setSession(name, res)
				redirectTarget = "/dashboard"
			}
		}
		http.Redirect(res, req, redirectTarget, 302)
	}
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
