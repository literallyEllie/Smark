package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// DBCredentials contains database credential data
type DBCredentials struct {
	Username string `json:"username"`
	Password string `json:"pwd"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
}

var session *mgo.Session

func dbInit() {

	// Load credentials
	jsonFile, err := os.Open("db.json")
	if err != nil {
		log.Fatal(err)
		return
	}

	defer jsonFile.Close()

	credBytes, _ := ioutil.ReadAll(jsonFile)
	var c DBCredentials
	json.Unmarshal(credBytes, &c)

	// Start up db session
	session, err = mgo.Dial(fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", c.Username, c.Password, c.Host, c.Port, c.Database))
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Println("Connected to DB")
}

func userCollection() *mgo.Collection {
	return session.DB("smark").C("users")
}

// GetUserByEmail queries the database and gets a user matching the email.
func GetUserByEmail(email string) *User {
	var rUser *User
	err := userCollection().Find(bson.M{"email": cIQuery(email)}).One(&rUser)
	if err != nil {
		return nil
	}

	return rUser
}

// GetUserByName queries the database and gets a user matching the username.
func GetUserByName(username string) *User {
	var rUser *User
	err := userCollection().Find(bson.M{"username": cIQuery(username)}).One(&rUser)
	if err != nil {
		return nil
	}

	return rUser
}

// GetUserByEmailUsername attemps to get a user by their username or email
func GetUserByEmailUsername(field string) *User {
	var rUser *User
	err := userCollection().Find(bson.M{"$or": []bson.M{{"username": cIQuery(field)}, {"email": cIQuery(field)}}}).One(&rUser)
	if err != nil {
		return nil
	}

	return rUser
}

// InsertUserDB inserts a user object into the database
func InsertUserDB(user *User) {
	err := userCollection().Insert(&user)
	if err != nil {
		return
	}

	log.Printf("Created user %s", user.Username)
}

// UpdateUserDB updates an existing user object into the database
func UpdateUserDB(user *User) {
	err := userCollection().Update(bson.M{"email": cIQuery(user.Email)}, &user)
	if err != nil {
		log.Printf("[!!] Failed to update user %s : %e", user.Email, err)
		return
	}
}

// Makes a query- case insensitive
func cIQuery(in string) map[string]interface{} {
	return bson.M{"$regex": bson.RegEx{Pattern: "^" + in + "$", Options: "i"}}
}
