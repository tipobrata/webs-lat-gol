package main

import (
	"html/template"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/securecookie"
	_ "github.com/mattn/go-sqlite3"
)

var (
	validPath     = regexp.MustCompile(`^/(edit|save|test)/([:\w+:]+)$`)
	cookieHandler = securecookie.New(
		securecookie.GenerateRandomKey(64),
		securecookie.GenerateRandomKey(32))
)

func view(w http.ResponseWriter, r *http.Request, title string) {
	title = strings.Title(title)
	p, err := loadSource(title)
	if err != nil {
		p, _ = load(title)
	}
	if p.Title == "" {
		p, _ = load(title)
	}
	if strings.Contains(p.Title, "_") {
		p.Title = strings.Replace(p.Title, "_", " ", -1)
	}
	render(w, "test", p)
}

func edit(w http.ResponseWriter, r *http.Request, title string) {
	title = strings.Title(title)
	p, err := loadSource(title)
	if err != nil {
		p, _ = load(title)
	}
	if p.Title == "" {
		p, _ = load(title)
	}
	if strings.Contains(p.Title, "_") {
		p.Title = strings.Replace(p.Title, "_", " ", -1)
	}
	render(w, "edit", p)
}

func save(w http.ResponseWriter, r *http.Request, title string) {
	title = strings.Replace(title, " ", "_", -1)
	body := r.FormValue("body")
	p := &Page{Title: strings.Title(title), Body: []byte(body)}
	p.saveCache()
	http.Redirect(w, r, "/test/"+title, http.StatusFound)
}

func upload(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":

		title := "Upload"
		p := &Page{Title: title}
		render(w, "upload", p)

	case "POST":
		err := r.ParseMultipartForm(100000)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		m := r.MultipartForm
		files := m.File["myfiles"]
		for i := range files {
			file, err := files[i].Open()
			defer file.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			f, err := os.Create("./files/" + files[i].Filename)
			defer f.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if _, err := io.Copy(f, file); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/files/"+files[i].Filename, http.StatusFound)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func indexPage(w http.ResponseWriter, r *http.Request) {
	uuid := getUUID(r)
	if uuid != "" {
		http.Redirect(w, r, "/example", 302)
		return
	}
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

func create(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		p := &Page{}
		render(w, "create", p)
	case "POST":
		title := r.FormValue("title")
		body := r.FormValue("body")
		if strings.Contains(title, " ") {
			title = strings.Replace(title, " ", "_", -1)
			//http.Redirect(w, r, "/create", 302)
			//return
		}
		p := &Page{Title: strings.Title(title), Body: []byte(body)}
		err := p.saveCache()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		http.Redirect(w, r, "/test/"+title, 302)
		return
	}
}

func search(w http.ResponseWriter, r *http.Request) {
	sValue := r.FormValue("search")
	sValue = strings.Title(sValue)
	sValue = strings.Replace(sValue, " ", "_", -1)
	if b, _ := pageExists(sValue); b == true {
		http.Redirect(w, r, "/test/"+sValue, 302)
		return
	}
	render(w, "search", &Page{Title: strings.Title(sValue)})
}

func render(w http.ResponseWriter, name string, data interface{}) {
	funcMap := template.FuncMap{

		"urlize": func(s string) string {
			return strings.Replace(s, " ", "_", -1)
		},
	}
	tmpl, err := template.New(name).Funcs(funcMap).ParseGlob("templates/*.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	tmpl.ExecuteTemplate(w, name, data)
}

func checkUUID(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := getUUID(r)
		if uuid != "" {
			fn(w, r)
			return
		}
		http.Redirect(w, r, "/", 302)
	}
}

func checkPath(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := validPath.FindStringSubmatch(r.URL.Path)
		if path == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, path[2])
	}
}

func main() {
	govalidator.SetFieldsRequiredByDefault(true)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir("files"))))
	http.HandleFunc("/", indexPage)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/example", examplePage)
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/test/", checkUUID(checkPath(view)))
	http.HandleFunc("/edit/", checkUUID(checkPath(edit)))
	http.HandleFunc("/save/", checkUUID(checkPath(save)))
	http.HandleFunc("/create/", checkUUID(create))
	http.HandleFunc("/upload/", checkUUID(upload))
	http.HandleFunc("/search", checkUUID(search))
	http.ListenAndServe(":8000", nil)
}
