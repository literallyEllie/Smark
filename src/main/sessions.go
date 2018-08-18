package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

const (
	// FlashTypeInfo is to declare info flash cookies
	FlashTypeInfo = 0
	// FlashTypeErr is to declare error flash cookies
	FlashTypeErr = 1
)

// FlashCookie contains flash data of a session
type FlashCookie struct {
	Type    int
	Message string
}

// Cookies is where the cookies are stored.
var cookies *sessions.CookieStore

// SessionData is a map of all users with a valid cookie indexed by their session key
var SessionData = map[string]*User{}

func sessionsInit() {
	// Load up hash for passwords
	key, err := ioutil.ReadFile("sess_key.txt")
	if err != nil {
		log.Fatal(err)
		return
	}
	cookies = sessions.NewCookieStore(key)

	gob.Register(FlashCookie{})
}

// Client recv handles

func loginHandle(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {

		// get credentials
		username := req.FormValue("username")
		password := req.FormValue("password")

		// get from db
		u := GetAccount(username)

		// tell them to go away
		if u == nil {
			CreateFlashCookie(req, w, FlashTypeErr, "User doesn't exist")
			http.Redirect(w, req, "/login", http.StatusSeeOther)
			return
		}

		// check credentials
		if (u.Username == username || u.Email == username) && passMatch(u.Password, []byte(password)) {
			createCookie(u, req, w)
			http.Redirect(w, req, "/dashboard", http.StatusSeeOther)
			return
		}

		// go away
		CreateFlashCookie(req, w, FlashTypeErr, "Invalid credentials")
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}

	// Get their user and make an instance of view data
	user, _, err := GetSessionedUser(req, w)
	viewData := &ViewData{Viewer: user}

	// Get any flash cookies from previous loadings
	LoadFlashCookies(req, w, viewData)

	if err != nil {
		viewData.Viewer = &User{Username: ""}
	}

	// Load template
	err = templates.ExecuteTemplate(w, "login.html", viewData)
	if err != nil {
		log.Println("Error exectuing login template ", err)
	}

}

func signupHandle(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {

		// get credentials
		email := req.FormValue("email")
		username := req.FormValue("username")
		password := req.FormValue("password")

		u, err := createUser(email, username, password)
		if err != nil {
			CreateFlashCookie(req, w, FlashTypeErr, err.Error())
			http.Redirect(w, req, "/signup", http.StatusSeeOther)
			return
		}

		// create session + redirect
		createCookie(u, req, w)
		http.Redirect(w, req, "/dashboard", http.StatusSeeOther)
		return
	}

	// Get their session and create an instance of view data
	user, _, err := GetSessionedUser(req, w)
	viewData := &ViewData{Viewer: user}

	// Get their flash data from previous sessions
	LoadFlashCookies(req, w, viewData)

	// Execute the template.
	err = templates.ExecuteTemplate(w, "signup.html", viewData)
	if err != nil {
		log.Println("Error exectuing signup template ", err)
	}

}

func logoutHandle(w http.ResponseWriter, req *http.Request) {
	user, _, err := GetSessionedUser(req, w)
	if err != nil {
		CreateFlashCookie(req, w, FlashTypeErr, err.Error())
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}

	// Clean them up
	deleteCookie(user, req, w)
	CreateFlashCookie(req, w, FlashTypeInfo, "You have been logged out.")
	http.Redirect(w, req, "/login", http.StatusSeeOther)
}

// GetSessionedUser gets user data, their session id and if an error occurs, that too
func GetSessionedUser(req *http.Request, w http.ResponseWriter) (*User, string, error) {
	session, err := cookies.Get(req, "session-id")
	user := &User{Username: ""}

	if err != nil {
		return user, "", err
	}

	// Get their session key id thing
	sessionKeyRaw := session.Values["id"]
	if sessionKeyRaw == nil {
		// I know this goes against error conventions but it shows to the end user
		return user, "", errors.New(" You were not logged in. ")
	}
	sessionKey := sessionKeyRaw.(string)

	user, ok := SessionData[sessionKey]

	// If they aren't logged in
	if !ok {
		user = &User{Username: ""}
		return user, "", errors.New(" You were not logged in. ")
	}

	return user, sessionKey, nil
}

// Session assignment

// Generates a random session key from 32 bytes then encoding to Base64
func generateSessionKey() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// Creates a cookie based on user data and sets to a  response writer
func createCookie(u *User, req *http.Request, w http.ResponseWriter) {
	session, err := cookies.Get(req, "session-id")
	if err != nil {
		log.Printf("[!!] Failed to create get cookie info from Cookies for %s", u.Username)
	}

	// make new key
	newKey := generateSessionKey()
	// Mark client with key
	session.Values["id"] = newKey
	cookies.Save(req, w, session)
	// Map session key to user
	SessionData[newKey] = u
}

// Deletes a cookie by a user
func deleteCookie(u *User, req *http.Request, w http.ResponseWriter) {
	// Can't get data via GetSessionedUser as we need to get Session and expire it.
	session, err := cookies.Get(req, "session-id")
	if err != nil {
		UserDB[u.Username] = u
		log.Printf("[!!] Failed to delete get cookie info from Cookies for %s", u.Username)
		// TODO try and get session id from looking SessionData

		cookie, err := req.Cookie("session-id")
		if err == nil {
			cookie.MaxAge = -1
			http.SetCookie(w, cookie)
		}

		return
	}

	// Get their session ID
	sessionKeyRaw := session.Values["id"]
	if sessionKeyRaw == nil {
		return
	}
	sessionKey := sessionKeyRaw.(string)

	// Remove from session data
	delete(SessionData, sessionKey)
	// Push to database
	UserDB[u.Username] = u

	// Expire cookie
	session.Options.MaxAge = -1
	cookies.Save(req, w, session)

	// remove it manually in case their session was corrupted
	cookie, err := req.Cookie("session-id")
	if err == nil {
		cookie.MaxAge = -1
		http.SetCookie(w, cookie)
	}

}

// CreateFlashCookie creates a temporary cookie which is used to show tempoary notifications to them for the next reload
func CreateFlashCookie(req *http.Request, w http.ResponseWriter, flashType int, contents string) {
	session, _ := cookies.Get(req, "flash-data")
	flashData := FlashCookie{
		Type:    flashType,
		Message: contents,
	}
	session.AddFlash(flashData)
	err := session.Save(req, w)
	if err != nil {
		log.Println("[!!] Failed to save session after adding flash data: ", err)
	}
}

// LoadFlashCookies loads in any created flash cookies and appends them to an instance of ViewData
func LoadFlashCookies(req *http.Request, w http.ResponseWriter, viewData *ViewData) *ViewData {
	session, _ := cookies.Get(req, "flash-data")
	flashCookies := session.Flashes()
	err := session.Save(req, w)
	if err != nil {
		log.Println("[!!] Failed to save session after reading flash data: ", err)
	}

	fCookies := make([]FlashCookie, len(flashCookies))

	for index, flashCookie := range flashCookies {
		fCookies[index] = flashCookie.(FlashCookie)
	}

	viewData.FlashData = fCookies

	return viewData
}

// CheckAccess checks access of a requester ensuring they have rights to visit
func CheckAccess(w http.ResponseWriter, req *http.Request, reqPage string) (*User, error) {
	var user *User

	user, _, err := GetSessionedUser(req, w)

	// If they're not logged in (i.e in SessionData) and they're not already trying to login, tell them to go away.
	if err != nil && !(reqPage == "login" || reqPage == "signup" || reqPage == "404") {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return user, err
	}

	// Return their data and its all gucci
	w.WriteHeader(http.StatusOK)
	return user, nil
}
