package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	gContext "github.com/gorilla/context"
)

const templatePath = "templates"

// ViewData is the data passed to the templates when a page is loaded.
type ViewData struct {
	Viewer    *User
	FlashData []FlashCookie
}

// User contains data about a user
type User struct {
	Email    string `bson:"email"`
	Username string `bson:"username"`
	Password []byte `bson:"password"`
	IsAdmin  bool   `bson:"isadmin"`
}

var templates *template.Template
var regexEmail *regexp.Regexp

func main() {
	templates = populateTemplates()
	regexEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	// Init modules
	sessionsInit()
	dbInit()

	// Main handle
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		requestedPath := req.URL.Path[1:]
		if requestedPath == "/" || requestedPath == "" {
			requestedPath += "dashboard"
		}

		// Get the page
		template := templates.Lookup(requestedPath + ".html")

		if template != nil {
			// Check user has access to page
			user, err := CheckAccess(w, req, requestedPath)

			// they have rights, if not, they will have been redirected.
			if err == nil {
				// Execute page with viewer data.
				viewData := &ViewData{Viewer: user}
				LoadFlashCookies(req, w, viewData)

				err = template.Execute(w, viewData)
				if err != nil {
					log.Println("[!!] Failed to exectute template ", err)
				}
			}

		} else {
			http.Redirect(w, req, "/404", http.StatusSeeOther)
		}

	})
	http.HandleFunc("/login", loginHandle)
	http.HandleFunc("/signup", signupHandle)
	http.HandleFunc("/logout", logoutHandle)
	http.HandleFunc("/res/", handleResourceRequest)

	http.ListenAndServe(":8080", gContext.ClearHandler(http.DefaultServeMux))
}

// Method to handle requests to the resources folder
func handleResourceRequest(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path[len("/res"):]
	data, err := ioutil.ReadFile("templates/res/" + string(path))

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var contentType string

	if strings.HasSuffix(path, ".css") {
		contentType = "text/css"
	} else if strings.HasSuffix(path, ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(path, ".jpg") {
		contentType = "image/jpg"
	} else {
		contentType = "text/plain"
	}

	w.Header().Add("Content Type", contentType)
	w.Write(data)
}

// Method to get templates
func populateTemplates() *template.Template {
	result := template.New("templates")

	templateFolder, _ := os.Open(templatePath)
	defer templateFolder.Close()

	templatePathsRaw, _ := templateFolder.Readdir(-1)
	templatePaths := new([]string)
	for _, pathInfo := range templatePathsRaw {
		if !pathInfo.IsDir() {
			*templatePaths = append(*templatePaths, templatePath+"/"+pathInfo.Name())
		}
	}

	_, err := result.ParseFiles(*templatePaths...)
	if err != nil {
		log.Println("[!!] Failed to load templates ", err)
	}

	return result
}
