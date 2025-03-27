package main

import (
    "database/sql"
    "encoding/json"
    "log"
    "math/rand"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"
    _ "github.com/lib/pq" // импортируем драйвер PostgreSQL
)

type Car struct {
    ID      string  `json:"id"`
    Model   string  `json:"model"`
    Client  *Client `json:"client"`
}

type Client struct {
    Firstname string `json:"firstname"`
    Lastname  string `json:"lastname"`
}

var db *sql.DB

func initDB() {
    var err error
    connStr := "user=postgres password=rootroot dbname=college sslmode=disable"
    db, err = sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }
}

func getCars(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    rows, err := db.Query("SELECT id, model, firstname, lastname FROM cars")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var cars []Car
    for rows.Next() {
        var car Car
        var client Client
        if err := rows.Scan(&car.ID, &car.Model, &client.Firstname, &client.Lastname); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        car.Client = &client
        cars = append(cars, car)
    }
    json.NewEncoder(w).Encode(cars)
}

func getCar(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    params := mux.Vars(r)
    
    var car Car
    var client Client
    err := db.QueryRow("SELECT id, model, firstname, lastname FROM cars WHERE id=$1", params["id"]).Scan(&car.ID, &car.Model, &client.Firstname, &client.Lastname)
    
    if err != nil {
        if err == sql.ErrNoRows {
            json.NewEncoder(w).Encode(&Car{})
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    car.Client = &client
    json.NewEncoder(w).Encode(car)
}

func createCar(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    var car Car
    _ = json.NewDecoder(r.Body).Decode(&car)

    // Генерация ID для нового автомобиля (в реальном приложении лучше использовать автоинкремент в БД)
    car.ID = strconv.Itoa(rand.Intn(1000000))

    _, err := db.Exec("INSERT INTO cars (id, model, firstname, lastname) VALUES ($1, $2, $3, $4)",
        car.ID, car.Model, car.Client.Firstname, car.Client.Lastname)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(car)
}

func updateCar(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    params := mux.Vars(r)

    var car Car
    _ = json.NewDecoder(r.Body).Decode(&car)

    _, err := db.Exec("UPDATE cars SET model=$1, firstname=$2, lastname=$3 WHERE id=$4",
        car.Model, car.Client.Firstname, car.Client.Lastname, params["id"])

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    car.ID = params["id"]
    json.NewEncoder(w).Encode(car)
}

func deleteCar(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    params := mux.Vars(r)

    _, err := db.Exec("DELETE FROM cars WHERE id=$1", params["id"])
    
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent) // Возвращаем статус 204 No Content
}

func main() {
    initDB()
    defer db.Close()

    r := mux.NewRouter()
    
    r.HandleFunc("/cars", getCars).Methods("GET")
    r.HandleFunc("/cars/{id}", getCar).Methods("GET")
    r.HandleFunc("/cars", createCar).Methods("POST")
    r.HandleFunc("/cars/{id}", updateCar).Methods("PUT")
    r.HandleFunc("/cars/{id}", deleteCar).Methods("DELETE")
    
    log.Fatal(http.ListenAndServe(":8000", r))
}
