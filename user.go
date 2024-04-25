package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strconv"
)

type CreateUserRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type User struct {
	ID       int    `json:"id"`
	Name     string `json"name"`
	Password string `json:"-"`
}

type session struct {
	UserID int `json:"userID"`
}

func (s *session) String() string {
	fmt.Println(s.UserID)
	user, ok := FindUserByID(s.UserID)
	if !ok {
		return "Unknown"
	}
	return user.Name
}

var (
	users    = make(map[int]User)
	sessions = make(map[int]*session)
)

func FindUserByID(id int) (User, bool) {
	user, ok := users[id]
	return user, ok
}

func FindUserBySessionID(id int) (int, bool) {
	session, ok := sessions[id]
	return session.UserID, ok
}

func FindUserByName(name string) (User, bool) {
	for _, user := range users {
		if user.Name == name {
			return user, true
		}
	}
	return User{}, false
}

func handleSignup(w http.ResponseWriter, r *http.Request) {
	var u CreateUserRequest
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, user := range users {
		if user.Name == u.Name {
			http.Error(w, "User already exists", http.StatusBadRequest)
			return
		}
	}

	user := User{
		ID:       rand.Int(),
		Name:     u.Name,
		Password: u.Password,
	}
	users[user.ID] = user
	fmt.Println(users)
	w.WriteHeader(http.StatusCreated)
}

func handleSignin(w http.ResponseWriter, r *http.Request) {
	var u CreateUserRequest
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, ok := FindUserByName(u.Name)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if user.Password != u.Password {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	s, id := NewSession(user.ID)
	sessions[id] = s
	fmt.Println(sessions)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(strconv.Itoa(id)))
}

func handleListUsers(w http.ResponseWriter, r *http.Request) {
	var userList []User
	for _, user := range users {
		userList = append(userList, user)
	}
	json.NewEncoder(w).Encode(userList)
}
