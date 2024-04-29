package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/rs/cors"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Resource struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Data string `json:"data"`
}

var (
	users        []User
	resources    []Resource
	sessionStore = NewRAMSessionStore()
)

func main() {

	users = []User{
		{ID: 1, Username: "admin", Password: "admin"},
		{ID: 2, Username: "user", Password: "password"},
	}

	resources = []Resource{
		{ID: 1, Name: "Resource 1"},
		{ID: 2, Name: "Resource 2"},
		{ID: 3, Name: "Resource 3"},
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/login", login)
	mux.HandleFunc("/resources", getResources)
	mux.HandleFunc("/resource", resource)

	fmt.Println("Server is running on port 8080")

	allowedOrigins := []string{"http://localhost:8080", "http://localhost:4000"}
	handler := cors.New(cors.Options{AllowedOrigins: allowedOrigins, AllowCredentials: true, Debug: true}).Handler(mux)
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

	for _, user := range users {
		if user.Username == userReq.Username && user.Password == userReq.Password {
			session, err := sessionStore.CreateSession()
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
	}

	http.Error(w, "Invalid username or password", http.StatusUnauthorized)
}

func isAuthorized(r *http.Request) bool {
	sessionID, err := r.Cookie("session")
	if err != nil {
		return false
	}
	session, err := sessionStore.GetSession(sessionID.Value)
	return session != nil && err == nil
}

func returnJson(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func getResources(w http.ResponseWriter, r *http.Request) {
	if !isAuthorized(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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

func resource(w http.ResponseWriter, r *http.Request) {
	if !isAuthorized(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == "POST" {
		postResource(w, r)
		return
	}

	id, err := strconv.Atoi(r.URL.Path[len("/resource/"):])
	if err != nil {
		http.Error(w, "Invalid resource ID", http.StatusBadRequest)
		return
	}

	if r.Method == "PUT" {
		putResource(w, r, id)
		return
	}

	getResource(w, id)
}

func getResource(w http.ResponseWriter, id int) {
	var result Resource
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
	for i, res := range resources {
		if res.ID == id {
			resources[i] = tmp
			returnJson(w, res)
			return
		}
	}
	http.Error(w, "Not found", http.StatusNotFound)
}

func postResource(w http.ResponseWriter, r *http.Request) {
	var res Resource
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	res.ID = resources[len(resources)-1].ID + 1
	resources = append(resources, res)
	returnJson(w, res)
}
