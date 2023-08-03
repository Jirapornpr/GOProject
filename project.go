package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql" //import db sql package
)

var db *sql.DB

const path = "student"

type Student struct {
	StudentID int    `json:"studentID"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Gender    string `json:"gender"`
	Country   string `json:"country"`
}

func connectDB() {
	var err error
	db, err = sql.Open("mysql", "springstudent:springstudent@tcp(127.0.0.1:3306)/coursedb")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(db)
}

func getStudent() ([]Student, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, "SELECT id, firstname, lastname, gender, country FROM student")
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer rows.Close()
	students := make([]Student, 0)
	for rows.Next() {
		var student Student
		rows.Scan(&student.StudentID,
			&student.Firstname,
			&student.Lastname,
			&student.Gender,
			&student.Country)

		students = append(students, student)
	}
	return students, nil
}

func getById(studentID int) (*Student, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows := db.QueryRowContext(ctx, `SELECT id, firstname, lastname, gender, country FROM student WHERE id = ?`, studentID)
	student := &Student{}
	err := rows.Scan(
		&student.StudentID,
		&student.Firstname,
		&student.Lastname,
		&student.Gender,
		&student.Country)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		log.Println(err)
		return nil, err
	}
	return student, nil
}

func studentsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		studentList, err := getStudent()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json, err := json.Marshal(studentList)
		if err != nil {
			log.Fatal(err)
		}
		_, err = w.Write(json)
		if err != nil {
			log.Fatal(err)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func studentByidHandler(w http.ResponseWriter, r *http.Request) {
	urlPathSegments := strings.Split(r.URL.Path, fmt.Sprintf("%s/", path))
	if len(urlPathSegments[1:]) > 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	studentID, err := strconv.Atoi(urlPathSegments[len(urlPathSegments)-1])
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	switch r.Method {
	case http.MethodGet:
		student, err := getById(studentID)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if student == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json, err := json.Marshal(student)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, err = w.Write(json)
		if err != nil {
			log.Fatal(err)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func corsMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Access-Control-Allow-Methods", "GET")
		w.Header().Add("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encodeing, Authorization, X-CSRF-Token")
		handler.ServeHTTP(w, r)
	})

}

func main() {
	connectDB()
	studentsHandler := http.HandlerFunc(studentsHandler)
	studentByidHandler := http.HandlerFunc(studentByidHandler)
	http.Handle(fmt.Sprintf("/%s", path), corsMiddleware(studentsHandler))
	http.Handle(fmt.Sprintf("/%s/", path), corsMiddleware(studentByidHandler))

	http.ListenAndServe(":5000", nil)
}
