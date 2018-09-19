package main

import (
	"github.com/andanhm/go-prettytime"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User contains data about a user
type User struct {
	// Credentials
	Email    string `bson:"email"`
	Username string `bson:"username"`
	Password []byte `bson:"password"`

	// Activity
	Online   bool
	LastSeen time.Time `bson:"lastseen"`

	// Misc
	IsAdmin bool   `bson:"isadmin"`
	Locale  string `bson:"locale"`
}

// QualifiedName returns their qualified name including their prefix tag, if applicable.
func (user User) QualifiedName() string {
	if user.IsAdmin {
		return user.Prefix() + " " + user.Username
	}

	return user.Username
}

// Prefix returns a prefix for the user
func (user User) Prefix() string {
	if user.IsAdmin {
		return "[ADMIN]"
	}
	return ""
}

// DisplayLastSeen formats the user's last seen timestamp.
func (user User) DisplayLastSeen() string {

	log.Print(user.Online)

	// Not seen
	if user.LastSeen.Unix() == -62135596800 {
		return "Unknown"
	}

	return prettytime.Format(user.LastSeen)
}

// UserDB is a temp map containing user data, an effective database
// var UserDB = map[string]*User{}

// GetAccount gets a user from the database with the given query and returns them
func GetAccount(query string) *User {

	/*
		for _, u := range UserDB {
			if strings.ToLower(u.Email) == strings.ToLower(query) {
				return u
			}
			if strings.ToLower(u.Username) == strings.ToLower(query) {
				return u
			}
		}
	*/

	return GetUserByEmailUsername(query)
}

// SaveAccount saves a userdata to db
func SaveAccount(user *User) {
	// UserDB[strings.ToLower(user.Username)] = user
	UpdateUserDB(user)
}

func createUser(locale string, email string, username string, password string) (*User, string) {
	// Validation checks
	if email == "" || !regexEmail.MatchString(email) {
		return nil, string(T(locale, "error.email-invalid"))
	}
	if username == "" {
		return nil, string(T(locale, "username.username-invalid"))
	}
	if len(password) < 6 {
		return nil, string(T(locale, "error.password-invalid"))
	}

	// Check if in use

	similarEmail := GetUserByEmail(email)
	if similarEmail != nil {
		return nil, string(T(locale, "error.email-used"))
	}

	similarUserName := GetUserByName(username)
	if similarUserName != nil {
		return nil, string(T(locale, "error.username-used"))
	}

	/*
		for _, u := range UserDB {
			if strings.ToLower(u.Email) == strings.ToLower(email) {
				return nil, errors.New("Email in-use")
			}
			if strings.ToLower(u.Username) == strings.ToLower(username) {
				return nil, errors.New("Username in-use")
			}
		}
	*/

	securePass := hashSaltPassword([]byte(password))
	if string(securePass) == password {
		return nil, string(T(locale, "error.cannot-hash"))
	}

	user := &User{
		Email:    email,
		Username: username,
		Password: securePass,
		IsAdmin:  false,
		Online:   true,
	}
	// UserDB[strings.ToLower(username)] = user

	// insert to db
	InsertUserDB(user)

	return user, ""
}

func hashSaltPassword(password []byte) []byte {
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.MinCost)
	if err != nil {
		log.Println("[!!] Error hashing password. ", err)
		return password
	}

	return hash
}

func passMatch(hashed []byte, input []byte) bool {
	err := bcrypt.CompareHashAndPassword(hashed, input)
	if err != nil {
		if !strings.Contains(err.Error(), "is not the hash of the given password") {
			log.Println("[!!] Error comparing hashed password. ", err)
		}
		return false
	}

	return true
}
