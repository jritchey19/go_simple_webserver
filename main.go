package main

import (
	"fmt"
	"errors"
	"net/http"
	"encoding/json"
	"strconv"
	"sync/atomic"
	"sync"
)

type User struct {
	Id    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ResponseData struct {
	Message string `json:"message"`
}

var users = make(map[int64]User)
var counter atomic.Int64
var mu sync.RWMutex

var (
	ErrBadId        = errors.New("bad or missing id")
	ErrNotFound     = errors.New("user not found")
	ErrMalformed    = errors.New("malformed data")
	ErrMissingInfo  = errors.New("missing input")
)

func getID(r string) (int64, error) {

	id,err := strconv.ParseInt(r,10,64)
	if err != nil {
		return 0, errors.New("Bad request.")
	}

	return id, nil
}

func getRecord(r *http.Request) (User, error) {
		var record User
		err := json.NewDecoder(r.Body).Decode(&record)

		if err != nil {
			return record, ErrMalformed
		}

		if len(record.Name) == 0 || len(record.Email) == 0 {
			return record, errors.New("Missing Input")
		} 

		return record, nil
}

func getUser(id int64) (User, error) {
	mu.RLock()
	defer mu.RUnlock()

	user, ok := users[id]
	if !ok {
		return User{}, ErrNotFound
	}

	return user, nil
}

func returnUser(r *http.Request) (User, error) {
  id, err := getID(r.PathValue("id"))
	if err != nil {
		return User{}, ErrBadId
	}

	user, err := getUser(id)
	if err != nil {
		return User{}, ErrNotFound
	}

	return user, nil
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		fmt.Fprintln(w, "This is a GET request to the home page")
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
	fmt.Fprintln(w, "Method not allowed")
}

func errorReturn(w http.ResponseWriter, err error, status int, headerStatus string, errorMessage string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Error-Reason", headerStatus)
	w.WriteHeader(status)
	resp := ErrorResponse{Error: errorMessage}

	json.NewEncoder(w).Encode(resp)
	if err == nil {
		fmt.Println(errorMessage)
		return
	}
	fmt.Println(errorMessage + ": ", err.Error())
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var record User
		err := json.NewDecoder(r.Body).Decode(&record)

		if err != nil {
			errorReturn(w, err, http.StatusBadRequest, "Invalid Input", "Malformed payload.")
			return
		}

		if len(record.Name) == 0 || len(record.Email) == 0 {
			errorReturn(w, nil, http.StatusBadRequest, "Missing Input", "Name and Email are required.")
			return
		} 

		mu.Lock()
		defer mu.Unlock()

		id := counter.Add(1)
		record.Id = id
		users[id] = record

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(users[id])

		fmt.Println("Successfully added: ", users[id])
		return
	} else if r.Method == http.MethodGet {
		mu.RLock()
		defer mu.RUnlock()
		values := make([]User, 0, len(users))
		for _, v := range users {
			values = append(values, v)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(values)

		fmt.Println("Returned all users.")

		return
	}
	errorReturn(w, nil, http.StatusMethodNotAllowed, "Invalid Method", "Method not allowed.")
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	user, err := returnUser(r)
	if err != nil {
		if errors.Is(err, ErrBadId) {
		  errorReturn(w, err, http.StatusBadRequest, "Bad Request", "Bad or missing ID.")
		  fmt.Println("Got bad request.")
		} else if errors.Is(err, ErrNotFound) {
		  errorReturn(w, err, http.StatusNotFound, "User not found", "User not found.")
		  fmt.Println("User not found.")
		} else {
			errorReturn(w, err, http.StatusInternalServerError, "Server Error", "Internal server error.")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
	fmt.Println("Returned user:",user)
}

func replaceUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r.PathValue("id"))
	if err != nil {
		errorReturn(w, err, http.StatusBadRequest, "Bad Request", "Bad or missing ID")
		fmt.Println("Got bad request.")
		return
	}

	record, err := getRecord(r)
	if err != nil {
		errorReturn(w, err, http.StatusBadRequest, "Malformed record", "Malformed record.")
		fmt.Println("Malformed record.")
		return
	}
	record.Id = id

	_, err = getUser(id)
	if err != nil {
		errorReturn(w, nil, http.StatusNotFound, "User not found", "User not found.")
		fmt.Println("User not found.")
		return
	}

	mu.Lock()
	defer mu.Unlock()
	users[id] = record

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users[id])
	fmt.Println("Replaced user: ", users[id])
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	user, err := returnUser(r)
	if err != nil {
		if errors.Is(err, ErrBadId) {
		  errorReturn(w, err, http.StatusBadRequest, "Bad Request", "Bad or missing ID.")
		  fmt.Println("Got bad request.")
		} else if errors.Is(err, ErrNotFound) {
		  errorReturn(w, err, http.StatusNotFound, "User not found", "User not found.")
		  fmt.Println("User not found.")
		} else {
			errorReturn(w, err, http.StatusInternalServerError, "Server Error", "Internal server error.")
		}
		return
	}

	mu.Lock()
	defer mu.Unlock()
	delete(users,user.Id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	fmt.Println("Deleted user: ", user.Name)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintln(w, "<html><body><h1>Welcome to my shop!</h1><h2>Let me cut your Mop!</h2><br><b>This is the about page of my terrible web server</b></body></html>")
}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("GET /home", homeHandler)
	mux.HandleFunc("GET /about", aboutHandler)
	mux.HandleFunc("/api/data", apiHandler)
	mux.HandleFunc("GET /api/data/{id}", getUserHandler)
	mux.HandleFunc("PUT /api/data/{id}", replaceUserHandler)
	mux.HandleFunc("DELETE /api/data/{id}", deleteUserHandler)

	fmt.Println("Server starting on http://localhost:8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
