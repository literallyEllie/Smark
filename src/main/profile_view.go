package main

import (
	"log"
	"net/http"
	"strings"
)

// ProfileView contains the data needed when viewing another profile
type ProfileView struct {
	Owner *User
}

func profileLoadHandle(w http.ResponseWriter, req *http.Request) {
	user, _, err := GetSessionedUser(req, w)
	if err != "" {
		CreateFlashCookie(req, w, FlashTypeErr, err)
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}

	requestedProfile := req.URL.Path[len("/profile/"):]

	log.Printf("Requested profile %s", requestedProfile)

	if requestedProfile == "" || strings.EqualFold(user.Username, requestedProfile) {
		log.Printf("%s attempted to view own profile, redircting (from %s).", user.Username, requestedProfile)
		http.Redirect(w, req, "/dashboard", http.StatusSeeOther)
		return
	}

	targetProfile := GetUserByName(requestedProfile)
	if targetProfile == nil {
		http.Redirect(w, req, "/404", http.StatusSeeOther)
		return
	}

	viewData := &ViewData{
		Viewer: user,
		ProfileView: ProfileView{
			Owner: targetProfile,
		},
	}

	templateErr := templates.ExecuteTemplate(w, "profile.html", viewData)
	if templateErr != nil {
		log.Println("Error exectuing profile template ", templateErr)
	}

}
