package main

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
)

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

var router = mux.NewRouter()

func indexPage(w http.ResponseWriter, r *http.Request) {
	msg := getMsg(w, r, "message")
	var u = &User{}
	u.Errors = make(map[string]string)
	if msg != "" {
		u.Errors["message"] = msg
		render(w, "signin", u)
	} else {
		u := &User{}
		render(w, "signin", u)
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("uname")
	pass := r.FormValue("password")
	u := &User{Username: name, Password: pass}
	redirect := "/"
	if name != "" && pass != "" {
		if b, uuid := userExists(u); b == true {
			setSession(&User{UUID: uuid}, w)
			redirect = "/example"
		} else {
			setMsg(w, "message", "Kesalahan memasukkan username dan password anda")
		}
	} else {
		setMsg(w, "message", "username dan password tidak boleh kosong")
	}
	http.Redirect(w, r, redirect, 302)
}

func logout(w http.ResponseWriter, r *http.Request) {
	clearSession(w, "session")
	http.Redirect(w, r, "/", 302)
}

func examplePage(w http.ResponseWriter, r *http.Request) {
	uuid := getUUID(r)
	u := getUserFromUUID(uuid)
	if uuid != "" {
		render(w, "internal", u)
	} else {
		setMsg(w, "message", "Silahkan Login dengan Akun Anda")
		http.Redirect(w, r, "/", 302)
	}
}

func signup(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		u := &User{}
		u.Errors = make(map[string]string)
		u.Errors["fname"] = getMsg(w, r, "fname")
		u.Errors["lname"] = getMsg(w, r, "lname")
		u.Errors["username"] = getMsg(w, r, "username")
		u.Errors["email"] = getMsg(w, r, "email")
		u.Errors["password"] = getMsg(w, r, "password")
		render(w, "signup", u)
	case "POST":
		if n := checkUser(r.FormValue("uSername")); n == true {
			setMsg(w, "username", "User Sudah ada. Silahkan masukkan username yang unik!")
			http.Redirect(w, r, "/signup", 302)
			return
		}
		u := &User{
			UUID:     UUID(),
			Fname:    r.FormValue("fName"),
			Lname:    r.FormValue("lName"),
			Username: r.FormValue("uSername"),
			Email:    r.FormValue("eMail"),
			Password: r.FormValue("password"),
		}
		result, err := govalidator.ValidateStruct(u)
		if err != nil {
			e := err.Error()
			if re := strings.Contains(e, "Username"); re == true {
				setMsg(w, "username", "Masukkan Username dengan benar")
			}
			if re := strings.Contains(e, "Fname"); re == true {
				setMsg(w, "fname", "Masukkan Firstname dengan benar")
			}
			if re := strings.Contains(e, "Lname"); re == true {
				setMsg(w, "lname", "Masukkan Lastname dengan benar")
			}
			if re := strings.Contains(e, "Email"); re == true {
				setMsg(w, "email", "Masukkan Email dengan benar")
			}
			if re := strings.Contains(e, "Password"); re == true {
				setMsg(w, "password", "Masukkan Password!")
			}
		}
		if r.FormValue("password") != r.FormValue("cpassword") {
			setMsg(w, "password", "Password Yang anda Masukkan tidak sama")
			http.Redirect(w, r, "/signup", 302)
			return
		}
		if result == true {
			u.Password = enyptPass(u.Password)
			saveData(u)
			http.Redirect(w, r, "/", 302)
			return
		}
		http.Redirect(w, r, "/signup", 302)
	}
}

func render(w http.ResponseWriter, name string, data interface{}) {
	tmpl, err := template.ParseGlob("*.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	tmpl.ExecuteTemplate(w, name, data)
}

func main() {
	govalidator.SetFieldsRequiredByDefault(true)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	router.HandleFunc("/", indexPage)
	router.HandleFunc("/login", login).Methods("POST")
	router.HandleFunc("/logout", logout).Methods("POST")
	router.HandleFunc("/example", examplePage)
	router.HandleFunc("/signup", signup).Methods("POST", "GET")
	http.Handle("/", router)
	http.ListenAndServe(":8000", nil)
}
