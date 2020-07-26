package main

import (
	"os"
	"fmt"
	nt "./network"
	bc "./blockchain"
	"net/http"
	"strconv"
	"strings"
	"html/template"
	"encoding/json"
)

const (
	STTC_PATH = "static/"
	TMPL_PATH = "templates/"
	ADDR_FILE = "addr.json"
)

func init() {
	err := json.Unmarshal([]byte(readFile(ADDR_FILE)), &Addresses)
	if err != nil {
		panic("failed: load addresses")
	}
	if len(Addresses) == 0 {
		panic("failed: len(Addresses) == 0")
	}
}

func main() {
	fmt.Println("Server is running ...")

	http.Handle("/static/", http.StripPrefix(
		"/static/",
		handleFileServer(http.Dir(STTC_PATH))),
	)

	http.HandleFunc("/", indexPage)
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/signup", signupPage)
	http.HandleFunc("/logout", logoutPage)
	http.HandleFunc("/account", accountPage)
	http.HandleFunc("/transaction", transactionPage)
	http.HandleFunc("/blockchain", blockchainPage)
	http.HandleFunc("/blockchain/", blockchainXPage)

	http.ListenAndServe(":7545", nil)
}

func handleFileServer(fs http.FileSystem) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := fs.Open(r.URL.Path); os.IsNotExist(err) {
			indexPage(w, r)
			return
		}
		http.FileServer(fs).ServeHTTP(w, r)
	})
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		TMPL_PATH+"base.html",
		TMPL_PATH+"index.html",
	)
	if err != nil {
		panic("can't load hmtl files")
	}
	var data struct{
		User *bc.User
	}
	data.User = User
	t.Execute(w, data)
}

func accountPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		TMPL_PATH+"base.html",
		TMPL_PATH+"account.html",
	)
	if err != nil {
		panic("can't load hmtl files")
	}
	var data struct{
		User *bc.User
		Address string
		Balance string
	}
	data.User = User
	if data.User != nil {
		data.Address = User.Address()
		res := nt.Send(Addresses[0], &nt.Package{
			Option: GET_BLNCE,
			Data:   data.Address,
		})
		if res != nil {
			data.Balance = res.Data
		}
	} else {
		http.Redirect(w, r, "/", 302)
		return
	}
	t.Execute(w, data)
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		TMPL_PATH+"base.html",
		TMPL_PATH+"login.html",
	)
	if err != nil {
		panic("can't load hmtl files")
	}
	var data struct{
		User *bc.User
		Error string
	}
	if r.Method == "POST" {
		r.ParseForm()
		User = bc.LoadUser(r.FormValue("private"))
		if User == nil {
			data.Error = "Load Private Key Error"
		} else {
			http.Redirect(w, r, "/", 302)
			return
		}
	}
	data.User = User
	t.Execute(w, data)
}

func logoutPage(w http.ResponseWriter, r *http.Request) {
	User = nil
	http.Redirect(w, r, "/", 302)
}

func signupPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		TMPL_PATH+"base.html",
		TMPL_PATH+"signup.html",
	)
	if err != nil {
		panic("can't load hmtl files")
	}
	var data struct{
		User *bc.User
		PrivateKey string
	}
	data.User = User
	if r.Method == "POST" {
		data.PrivateKey = bc.NewUser().Purse()
	}
	t.Execute(w, data)
}

func transactionPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		TMPL_PATH+"base.html",
		TMPL_PATH+"transaction.html",
	)
	if err != nil {
		panic("can't load hmtl files")
	}
	var data struct{
		User *bc.User
		Error string
	}
	data.User = User
	if r.Method == "POST" {
		r.ParseForm()
		if data.User == nil {
			data.Error = "User not authorizated"
			t.Execute(w, data)
			return
		} 
		receiver := r.FormValue("receiver")
		num, err := strconv.Atoi(r.FormValue("value"))
		if err != nil {
			data.Error = "strconv.Atoi error"
			t.Execute(w, data)
			return
		}
		flag := false 
		for _, addr := range Addresses {
			res := nt.Send(addr, &nt.Package{
				Option: GET_LHASH,
			})
			if res == nil {
				continue
			}
			tx := bc.NewTransaction(User, bc.Base64Decode(res.Data), receiver, uint64(num))
			res = nt.Send(addr, &nt.Package{
				Option: ADD_TRNSX,
				Data:   bc.SerializeTX(tx),
			})
			if res == nil || res.Data != "ok" {
				continue
			}
			flag = true
		}
		if !flag {
			data.Error = "TX failed"
		} else {
			data.Error = "TX success"
		}
	}
	t.Execute(w, data)
}

func blockchainPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		TMPL_PATH+"base.html",
		TMPL_PATH+"blockchain.html",
	)
	if err != nil {
		panic("can't load hmtl files")
	}
	var data struct{
		Error string
		Size []bool
		Address string
		Balance string 
		User *bc.User
	}
	data.User = User
	if r.Method == "POST" {
		data.Address = r.FormValue("address")
		res := nt.Send(Addresses[0], &nt.Package{
			Option: GET_BLNCE,
			Data: data.Address,
		})
		if res != nil {
			data.Balance = res.Data
		}
	}
	res := nt.Send(Addresses[0], &nt.Package{
		Option: GET_CSIZE,
	})
	if res == nil || res.Data == "" {
		data.Error = "Receive error"
		t.Execute(w, data)
		return 
	}
	num, err := strconv.Atoi(res.Data)
	if err != nil {
		data.Error = "strconv.Atoi error"
		t.Execute(w, data)
		return 
	}
	data.Size = make([]bool, num)
	t.Execute(w, data)
}

func blockchainXPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		TMPL_PATH+"base.html",
		TMPL_PATH+"blockchainX.html",
	)
	if err != nil {
		panic("can't load hmtl files")
	}
	var data struct{
		Error string
		Block *bc.Block
		User *bc.User
	}
	data.User = User
	res := nt.Send(Addresses[0], &nt.Package{
		Option: GET_BLOCK,
		Data: strings.Replace(r.URL.Path, "/blockchain/", "", 1),
	})
	if res == nil || res.Data == "" {
		data.Error = "Receive error"
		t.Execute(w, data)
		return 
	}
	data.Block = bc.DeserializeBlock(res.Data)
	if data.Block == nil {
		data.Error = "Block is nil"
		t.Execute(w, data)
		return 
	}
	t.Execute(w, data)
}
