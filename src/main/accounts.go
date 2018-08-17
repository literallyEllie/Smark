package main

import (
	"errors"
	"strings"
)

// UserDB is a temp map containing user data, an effective database
var UserDB = map[string]*User{}

// GetAccount gets a user from the database with the given query and returns them
func GetAccount(query string) *User {
	for _, u := range UserDB {
		if strings.ToLower(u.Email) == strings.ToLower(query) {
			return u
		}
		if strings.ToLower(u.Username) == strings.ToLower(query) {
			return u
		}
	}

	return nil
}

// SaveAccount saves a userdata to db
func SaveAccount(user *User) {
	UserDB[strings.ToLower(user.Username)] = user
}

func createUser(email string, username string, password string) (*User, error) {
	// Validation checks
	if email == "" || !regexEmail.MatchString(email) {
		return nil, errors.New("Email invalid")
	}
	if username == "" {
		return nil, errors.New("Username invalid")
	}
	if len(password) < 6 {
		return nil, errors.New("Password must be at least 6 characters")
	}

	// Check if in use
	for _, u := range UserDB {
		if strings.ToLower(u.Email) == strings.ToLower(email) {
			return nil, errors.New("Email in-use")
		}
		if strings.ToLower(u.Username) == strings.ToLower(username) {
			return nil, errors.New("Username in-use")
		}
	}

	user := &User{
		Email:    email,
		Username: username,
		Password: password,
		IsAdmin:  false,
	}
	UserDB[strings.ToLower(username)] = user

	return user, nil
}
