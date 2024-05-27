package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/cors"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Resource struct {
	ID     int    `json:"id"`
	UserId int    `json:"userId"`
	Name   string `json:"name"`
	Data   string `json:"data"`
}

var (
	sessionStore = NewRAMSessionStore()
)

func main() {

	db, err := OpenDB("file:db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/login", login)
	mux.HandleFunc("/resources", getResources)
	mux.HandleFunc("/resource/{id}", updateResource)
	mux.HandleFunc("/resource", setResource)

	fmt.Println("Server is running on port 8080")

	allowedOrigins := []string{"http://localhost:8080", "http://localhost:4000"}
	handler := cors.New(
		cors.Options{
			AllowedOrigins:   allowedOrigins,
			AllowCredentials: true,
			AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS"},
			Debug:            true,
		}).Handler(mux)
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var userReq User
	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userId, err := IsAccountCorrect(userReq)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if userId > 0 {
		userReq.ID = userId
		session, err := sessionStore.CreateSession(userId)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:  "session",
			Value: session.ID,
			Path:  "/",
		})
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Error(w, "Invalid username or password", http.StatusUnauthorized)
}

func isAuthorized(r *http.Request) (bool, int) {
	sessionID, err := r.Cookie("session")
	if err != nil {
		return false, 0
	}
	session, err := sessionStore.GetSession(sessionID.Value)
	return session != nil && err == nil, session.UserId
}

func returnJson(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func getResources(w http.ResponseWriter, r *http.Request) {
	_, userId := isAuthorized(r)
	if !(userId > 0) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	resources := GetUserResources(User{ID: userId})

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		limit = len(resources)
	}
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		offset = 0
	}

	end := offset + limit
	if end > len(resources) {
		end = len(resources)
	}
	result := resources[offset:end]

	returnJson(w, result)
}

func setResource(w http.ResponseWriter, r *http.Request) {
	_, userId := isAuthorized(r)
	if !(userId > 0) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Only POST allowed", http.StatusBadRequest)
		return
	}

	postResource(w, r, userId)
}

func updateResource(w http.ResponseWriter, r *http.Request) {
	_, userId := isAuthorized(r)
	if !(userId > 0) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(r.URL.Path[len("/resource/"):])
	if err != nil {
		http.Error(w, "Invalid resource ID", http.StatusBadRequest)
		return
	}

	putResource(w, r, id)
	getResource(w, id)
}

func getResource(w http.ResponseWriter, id int) {
	var result Resource
	resources := GetUserResources(User{ID: 1})
	for _, res := range resources {
		if res.ID == id {
			result = res
			break
		}
	}

	if result.ID == 0 {
		http.Error(w, "Resource not found", http.StatusNotFound)
		return
	}
	returnJson(w, result)
}

func putResource(w http.ResponseWriter, r *http.Request, id int) {
	var tmp Resource
	if err := json.NewDecoder(r.Body).Decode(&tmp); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if tmp.ID <= 0 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	success, err := UpdateResource(tmp)
	if !success {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	returnJson(w, tmp)
}

func postResource(w http.ResponseWriter, r *http.Request, userId int) {
	var res Resource
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	res.UserId = userId
	success, err := InsertResource(res)
	if !success {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	returnJson(w, res)
}
