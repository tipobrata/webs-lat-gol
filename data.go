package main

import (
	"database/sql"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type Page struct {
	Title string
	Body  []byte
}

// this table users model

type User struct {
	UUID     string            `valid:"required,uuidv4"`
	Username string            `valid:"required,alphanum"`
	Password string            `valid:"required"`
	Fname    string            `valid:"required,alpha"`
	Lname    string            `valid:"required,alpha"`
	Email    string            `valid:"required,email"`
	Errors   map[string]string `valid:"-"`
}

func (p *Page) saveCache() error {
	var db, _ = sql.Open("sqlite3", "cache/db.sqlite3")
	defer db.Close()
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	if strings.Contains(p.Title, " ") {
		p.Title = strings.Replace(p.Title, " ", "_", -1)
	}
	f := "cache/" + p.Title + ".txt"
	db.Exec("create table if not exists pages(title text, body blob, timestamp text)")
	tx, _ := db.Begin()
	stmt, _ := tx.Prepare("insert into pages (title, body, timestamp) values (?, ?, ?)")
	_, err := stmt.Exec(p.Title, p.Body, timestamp)
	tx.Commit()
	ioutil.WriteFile(f, p.Body, 0600)
	return err
}

func load(title string) (*Page, error) {
	var db, _ = sql.Open("sqlite3", "cache/db.sqlite3")
	defer db.Close()
	f := "cache/" + title + ".txt"
	body, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func loadSource(title string) (*Page, error) {
	var db, _ = sql.Open("sqlite3", "cache/db.sqlite3")
	defer db.Close()
	var name string
	var body []byte
	q, err := db.Query("select title, body from pages where title = '" + title + "' order by timestamp Desc limit 1")
	if err != nil {
		return nil, err
	}
	for q.Next() {
		q.Scan(&name, &body)
	}
	return &Page{Title: title, Body: body}, nil
}

// Fungsi create database dan simpan data dan buat DB SQLite username signup ke sqlite3

func saveData(u *User) error {
	var db, _ = sql.Open("sqlite3", "cache/db.sqlite3")
	defer db.Close()
	db.Exec("create table if not exists users (uuid text not null unique, firstname text not null, lastname text not null, username text not null unique, email text not null, password text not null, primary key(uuid))")
	tx, _ := db.Begin()
	stmt, _ := tx.Prepare("insert into users(uuid, firstname, lastname, username, email, password) values (?, ?, ?, ?, ?, ?)")
	_, err := stmt.Exec(u.UUID, u.Fname, u.Lname, u.Username, u.Email, u.Password)
	tx.Commit()
	return err
}

// fungsi login
func userExists(u *User) (bool, string) {
	var db, _ = sql.Open("sqlite3", "cache/db.sqlite3")
	defer db.Close()
	var ps, uu string
	q, err := db.Query("select uuid, password from users where username = '" + u.Username + "'")
	if err != nil {
		return false, ""
	}
	for q.Next() {
		q.Scan(&uu, &ps)
	}
	pw := bcrypt.CompareHashAndPassword([]byte(ps), []byte(u.Password))
	if uu != "" && pw == nil {
		return true, uu
	}
	return false, ""
}

func pageExists(title string) (bool, error) {
	var db, _ = sql.Open("sqlite3", "cache/db.sqlite3")
	defer db.Close()
	var pt string
	var pb []byte
	q, err := db.Query("select title, body from pages where title = '" + title + "' order by timestamp Desc limit 1")
	if err != nil {
		return false, err
	}
	for q.Next() {
		q.Scan(&pt, &pb)
	}
	if pt != "" && pb != nil {
		return true, nil
	}
	return false, nil
}

func checkUser(user string) bool {
	var db, _ = sql.Open("sqlite3", "cache/db.sqlite3")
	defer db.Close()
	var un string
	q, err := db.Query("select username where users where username = '" + user + "'")
	if err != nil {
		return false
	}
	for q.Next() {
		q.Scan(&un)
	}
	if un == user {
		return true
	}
	return false
}

func getUserFromUUID(uuid string) *User {
	var db, _ = sql.Open("sqlite3", "cache/db.sqlite3")
	defer db.Close()
	var uu, fn, ln, un, em, pass string
	q, err := db.Query("select * from users where uuid = '" + uuid + "'")
	if err != nil {
		return &User{}
	}
	for q.Next() {
		q.Scan(&uu, &fn, &ln, &un, &em, &pass)
	}
	return &User{Username: un, Fname: fn, Lname: ln, Email: em, Password: pass}
}

func enyptPass(password string) string {
	pass := []byte(password)
	hashpw, _ := bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)
	return string(hashpw)
}

// ini fungsi UUID

func UUID() string {
	id, err := uuid.NewV4()
	if err != nil {
		return id.String()
	}
	return id.String()
}
