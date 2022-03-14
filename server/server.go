package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/go-session/session"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"
)

func main() {
	manager := manage.NewDefaultManager()
	manager.MustTokenStorage(store.NewMemoryTokenStore())
	clientStore := store.NewClientStore()
	clientStore.Set("999999", &models.Client{
		ID:     "999999",
		Secret: "super-secret",
		Domain: "http://localhost:9098",
	})
	manager.MapClientStorage(clientStore)

	srv := server.NewServer(server.NewConfig(), manager)
	srv.SetUserAuthorizationHandler(UserAuthorizeHandler)
	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Internal Error:", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Response Error:", re.Error.Error())
	})

	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/auth", AuthHandler)
	http.HandleFunc("/token", TokenHandler)
	http.HandleFunc("/authorize", AuthorizeHandler)

	log.Println("Server is running at 9098 port.")
	log.Fatal(http.ListenAndServe(":9098", nil))
}

func outputHTML(w http.ResponseWriter, req *http.Request, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer file.Close()
	fi, _ := file.Stat()
	http.ServeContent(w, req, file.Name(), fi.ModTime(), file)
}

func UserAuthorizeHandler(w http.ResponseWriter, r *http.Request) (userID string, err error) {
	store, err := session.Start(context.TODO(), w, r)
	if err != nil {
		return
	}
	uid, ok := store.Get("UserID")
	if !ok {
		if r.Form == nil {
			r.ParseForm()
		}
		store.Set("ReturnUri", r.Form)
		store.Save()
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}
	userID = uid.(string)
	store.Delete("UserID")
	store.Save()
	return
}
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	store, err := session.Start(context.TODO(), w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" {
		store.Set("LoggedInUserID", "000000")
		store.Save()
		w.Header().Set("Location", "/auth")
		w.WriteHeader(http.StatusFound)
		return
	}
	outputHTML(w, r, "static/login.html")

}
func AuthHandler(w http.ResponseWriter, r *http.Request) {
	store, err := session.Start(context.TODO(), w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, ok := store.Get("LoggedInUserID"); !ok {
		w.Header().Set("Location", "/login")
	}
	if r.Method == "POST" {
		var form url.Values
		if v, ok := store.Get("ReturnUri"); ok {
			form = v.(url.Values)
		}
		u := new(url.URL)
		u.Path = "/authorize"
		u.RawQuery = form.Encode()
		w.Header().Set("Location", u.String())
		w.WriteHeader(http.StatusFound)
		store.Delete("Form")

		if v, ok := store.Get("LoggedInUserID"); ok {
			store.Set("UserID", v)
		}
		store.Save()

		return
	}
	outputHTML(w, r, "static/auth.html")
}

func TokenHandler(w http.ResponseWriter, r *http.Request) {

}
func AuthorizeHandler(w http.ResponseWriter, r *http.Request) {

}
