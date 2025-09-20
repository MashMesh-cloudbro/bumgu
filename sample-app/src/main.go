package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// User 구조체 정의
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

var db *sql.DB

func main() {
	dsn := fmt.Sprintf("root:%s@tcp(mysql.sample:3306)/testdb?parseTime=true", "asdf")

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}
	log.Println("Successfully connected to MySQL database!")

	// --- users 테이블 및 샘플 데이터 생성 ---
	createTable()

	// --- HTTP 라우터 설정 ---
	// 각 URL 경로에 맞는 핸들러 함수를 등록합니다.
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/users", usersHandler)
	http.HandleFunc("/users/", userHandler) // /users/1, /users/2, ... 와 같은 경로 처리

	// --- 서버 시작 (기존과 동일) ---
	log.Println("Starting server on port 8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}

// 루트 경로 핸들러
func rootHandler(w http.ResponseWriter, r *http.Request) {
	var currentTime string
	err := db.QueryRow("SELECT NOW()").Scan(&currentTime)
	if err != nil {
		http.Error(w, "Failed to query database", http.StatusInternalServerError)
		log.Printf("query failed: %v", err)
		return
	}
	fmt.Fprintf(w, "Hello from Go App!\nMySQL current time: %s", currentTime)
}

// '/users' 경로에 대한 핸들러
func usersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet: // 모든 사용자 조회 (Read All)
		getUsers(w, r)
	case http.MethodPost: // 새 사용자 생성 (Create)
		createUser(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// '/users/{id}' 경로에 대한 핸들러
func userHandler(w http.ResponseWriter, r *http.Request) {
	// URL 경로에서 ID 추출
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 || parts[2] == "" {
		http.Error(w, "Invalid URL, user ID is missing", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet: // 특정 사용자 조회 (Read One)
		getUser(w, r, id)
	case http.MethodPut: // 사용자 정보 수정 (Update)
		updateUser(w, r, id)
	case http.MethodDelete: // 사용자 삭제 (Delete)
		deleteUser(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// C: 새 사용자 생성
func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Requsted User: {+%v}", user)

	result, err := db.Exec("INSERT INTO users (name, age) VALUES (?, ?)", user.Name, user.Age)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		log.Printf("db.Exec failed: %v", err)
		return
	}

	id, _ := result.LastInsertId()
	user.ID = int(id)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// R: 모든 사용자 조회
func getUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, age FROM users")
	if err != nil {
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		log.Printf("db.Query failed: %v", err)
		return
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Age); err != nil {
			http.Error(w, "Failed to scan user data", http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}
	json.NewEncoder(w).Encode(users)
}

// R: 특정 사용자 조회
func getUser(w http.ResponseWriter, r *http.Request, id int) {
	var user User
	err := db.QueryRow("SELECT id, name, age FROM users WHERE id = ?", id).Scan(&user.ID, &user.Name, &user.Age)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get user", http.StatusInternalServerError)
		}
		log.Printf("db.QueryRow failed: %v", err)
		return
	}
	json.NewEncoder(w).Encode(user)
}

// U: 사용자 정보 수정
func updateUser(w http.ResponseWriter, r *http.Request, id int) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE users SET name = ?, age = ? WHERE id = ?", user.Name, user.Age, id)
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		log.Printf("db.Exec failed: %v", err)
		return
	}

	user.ID = id
	json.NewEncoder(w).Encode(user)
}

// D: 사용자 삭제
func deleteUser(w http.ResponseWriter, r *http.Request, id int) {
	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		log.Printf("db.Exec failed: %v", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DB 테이블 생성 함수
func createTable() {
	query := `
    CREATE DATABASE IF NOT EXISTS testdb;
    USE testdb;
    CREATE TABLE IF NOT EXISTS users (
        id INT AUTO_INCREMENT,
        name VARCHAR(50),
        age INT,
        PRIMARY KEY (id)
    );`

	// 여러 쿼리를 한 번에 실행
	if _, err := db.Exec("CREATE DATABASE IF NOT EXISTS testdb"); err != nil {
		log.Fatalf("could not create database: %v", err)
	}

	if _, err := db.Exec("USE testdb"); err != nil {
		log.Fatalf("could not select database: %v", err)
	}

	if _, err := db.Exec(strings.Split(query, ";")[2]); err != nil {
		log.Fatalf("could not create table: %v", err)
	}

	log.Println("Database and table checked/created successfully.")
}
