package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/context"
)

// UserDB is a temp map containing user data, an effective database
var UserDB = map[string]*User{}

const templatePath = "templates"

// User contains data about a user
type User struct {
	Username  string
	Password  string
	IsAdmin   bool
	LoggedOut bool
}

var templates *template.Template

func main() {
	templates = populateTemplates()

	testUser := &User{Username: "Test", Password: "jkdgf", IsAdmin: false}
	UserDB[testUser.Username] = testUser

	// userTwo := &user{fName: "Dave", lName: "Cool", admin: true}

	//http.Handle("/jim", userOne)
	//	http.Handle("/dave", userTwo)
	// http.Handle("/profile", new(user))
	// http.Handle("/", new(fileHandle))
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		requestedFile := req.URL.Path[1:]

		// Get the page
		template := templates.Lookup(requestedFile + ".html")

		if template != nil {
			log.Println("not null template")
			// Check user has access to page
			user, err := CheckAccess(w, req, requestedFile)

			// they have rights, if not, they will have been redirected.
			if err == nil {
				// Execute page with user template.
				template.Execute(w, user)
			}

		} else {
			w.WriteHeader(404)
		}

	})
	http.HandleFunc("/login", loginHandle)
	http.HandleFunc("/logout", logoutHandle)

	http.ListenAndServe(":8080", context.ClearHandler(http.DefaultServeMux))
}

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
	log.Println(err)
	return result
}
