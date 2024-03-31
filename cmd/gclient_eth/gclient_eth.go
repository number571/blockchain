package main

import (
	"os"
	"fmt"
	"strings"
	"context"
	"net/http"
	"math/big"
	"html/template"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

const (
	STTC_PATH = "static/"
	TMPL_PATH = "templates_eth/"
)

func init() {
	if ClientETH == nil {
		panic("failed: connect to ETH")
	}
	if Instance == nil {
		panic("failed: instance is nil")
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
	http.HandleFunc("/logout", logoutPage)
	http.HandleFunc("/account", accountPage)

	http.HandleFunc("/blockchain", blockchainPage)
	http.HandleFunc("/blockchain/estates", blockchainEstatesPage)
	http.HandleFunc("/blockchain/presents", blockchainPresentsPage)
	http.HandleFunc("/blockchain/sales", blockchainSalesPage)
	http.HandleFunc("/blockchain/rents", blockchainRentsPage)

	http.HandleFunc("/blockchain/estates/", blockchainEstatesXPage)
	http.HandleFunc("/blockchain/presents/", blockchainPresentsXPage)
	http.HandleFunc("/blockchain/sales/", blockchainSalesXPage)
	http.HandleFunc("/blockchain/rents/", blockchainRentsXPage)

	http.HandleFunc("/blockchain/presents/do/", blockchainPresentsDoPage)
	http.HandleFunc("/blockchain/sales/do/", blockchainSalesDoPage)
	http.HandleFunc("/blockchain/rents/do/", blockchainRentsDoPage)

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
		User *UserType
	}
	data.User = User
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
		User *UserType
		Error string
	}
	if r.Method == "POST" {
		r.ParseForm()
		User = loadUser(r.FormValue("private"))
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

func accountPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		TMPL_PATH+"base.html",
		TMPL_PATH+"account.html",
	)
	if err != nil {
		panic("can't load hmtl files")
	}
	var data struct{
		User *UserType
		Address string
		Balance string
	}
	data.User = User
	if data.User != nil {
		data.Address = User.AddressHex
		balance, err := ClientETH.BalanceAt(context.Background(), User.AddressEth, nil)
		if err == nil {
			data.Balance = balance.String()
		}
	} else {
		http.Redirect(w, r, "/", 302)
		return
	}
	t.Execute(w, data)
}

func blockchainPresentsDoPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		TMPL_PATH+"base.html",
		TMPL_PATH+"presentsDo.html",
	)
	if err != nil {
		panic("can't load hmtl files")
	}
	if User == nil {
		http.Redirect(w, r, "/login", 302)
		return
	}
	var data struct{
		User *UserType
		Block *EstateStr
		Error string
	}
	data.User = User
	var (
		index = new(big.Int)
		ok bool
	)
	num := strings.Replace(r.URL.Path, "/blockchain/presents/do/", "", 1)
	index, ok = index.SetString(num, 10)
	if !ok {
		data.Error = "strconv error"
		t.Execute(w, data)
		return
	}
	estate := getEstates(index)
	if estate == nil {
		data.Error = "estate is nil"
		t.Execute(w, data)
		return
	}
	data.Block = estatesToString(estate)
	if r.Method == "POST" {
		r.ParseForm()
		_, err := Instance.CreatePresent(
			resetAuth(User), 
			index, 
			common.HexToAddress(r.FormValue("address")),
		)
		if err != nil {
			data.Error = err.Error()
			t.Execute(w, data)
			return
		}
		data.Error = "Success created"
	}
	t.Execute(w, data)
}

func blockchainSalesDoPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		TMPL_PATH+"base.html",
		TMPL_PATH+"salesDo.html",
	)
	if err != nil {
		panic("can't load hmtl files")
	}
	if User == nil {
		http.Redirect(w, r, "/login", 302)
		return
	}
	var data struct{
		User *UserType
		Block *EstateStr
		Error string
	}
	data.User = User
	var (
		index = new(big.Int)
		ok bool
	)
	num := strings.Replace(r.URL.Path, "/blockchain/sales/do/", "", 1)
	index, ok = index.SetString(num, 10)
	if !ok {
		data.Error = "strconv error"
		t.Execute(w, data)
		return
	}
	estate := getEstates(index)
	if estate == nil {
		data.Error = "estate is nil"
		t.Execute(w, data)
		return
	}
	data.Block = estatesToString(estate)
	if r.Method == "POST" {
		r.ParseForm()
		var (
			price = new(big.Int)
			ok bool
		)
		price, ok = price.SetString(r.FormValue("price"), 10)
		if !ok {
			data.Error = "strconv error"
			t.Execute(w, data)
			return
		}
		_, err := Instance.CreateSale(
			resetAuth(User), 
			index, 
			price,
		)
		if err != nil {
			data.Error = err.Error()
			t.Execute(w, data)
			return
		}
		data.Error = "Success created"
	}
	t.Execute(w, data)
}

func blockchainRentsDoPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		TMPL_PATH+"base.html",
		TMPL_PATH+"rentsDo.html",
	)
	if err != nil {
		panic("can't load hmtl files")
	}
	if User == nil {
		http.Redirect(w, r, "/login", 302)
		return
	}
	var data struct{
		User *UserType
		Block *EstateStr
		Error string
	}
	data.User = User
	var (
		index = new(big.Int)
		ok bool
	)
	num := strings.Replace(r.URL.Path, "/blockchain/rents/do/", "", 1)
	index, ok = index.SetString(num, 10)
	if !ok {
		data.Error = "strconv error"
		t.Execute(w, data)
		return
	}
	rent := getEstates(index)
	if rent == nil {
		data.Error = "rent is nil"
		t.Execute(w, data)
		return
	}
	data.Block = estatesToString(rent)
	if r.Method == "POST" {
		r.ParseForm()
		var (
			days = new(big.Int)
			price = new(big.Int)
			ok bool
		)
		days, ok = days.SetString(r.FormValue("days"), 10)
		if !ok {
			data.Error = "strconv error"
			t.Execute(w, data)
			return
		}
		price, ok = price.SetString(r.FormValue("price"), 10)
		if !ok {
			data.Error = "strconv error"
			t.Execute(w, data)
			return
		}
		_, err := Instance.CreateRent(
			resetAuth(User), 
			index, 
			days,
			price,
		)
		if err != nil {
			data.Error = err.Error()
			t.Execute(w, data)
			return
		}
		data.Error = "Success created"
	}
	t.Execute(w, data)
}

func blockchainEstatesXPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		TMPL_PATH+"base.html",
		TMPL_PATH+"estatesX.html",
	)
	if err != nil {
		panic("can't load hmtl files")
	}
	if User == nil {
		http.Redirect(w, r, "/login", 302)
		return
	}
	var data struct{
		User *UserType
		Block *EstateStr
		Error string
	}
	data.User = User
	var (
		index = new(big.Int)
		ok bool
	)
	num := strings.Replace(r.URL.Path, "/blockchain/estates/", "", 1)
	index, ok = index.SetString(num, 10)
	if !ok {
		data.Error = "strconv error"
		t.Execute(w, data)
		return
	}
	estate := getEstates(index)
	if estate == nil {
		data.Error = "estate is nil"
		t.Execute(w, data)
		return
	}
	data.Block = estatesToString(estate)
	t.Execute(w, data)
}

func blockchainPresentsXPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		TMPL_PATH+"base.html",
		TMPL_PATH+"presentsX.html",
	)
	if err != nil {
		panic("can't load hmtl files")
	}
	if User == nil {
		http.Redirect(w, r, "/login", 302)
		return
	}
	var data struct{
		User *UserType
		Block *PresentStr
		Error string
	}
	data.User = User
	var (
		index = new(big.Int)
		ok bool
	)
	num := strings.Replace(r.URL.Path, "/blockchain/presents/", "", 1)
	index, ok = index.SetString(num, 10)
	if !ok {
		data.Error = "strconv error"
		t.Execute(w, data)
		return
	}
	present := getPresents(index)
	if present == nil {
		data.Error = "present is nil"
		t.Execute(w, data)
		return
	}
	data.Block = presentsToString(present)
	if r.Method == "POST" {
		r.ParseForm()
		if r.FormValue("cancel") != "" {
			_, err := Instance.CancelPresent(
				resetAuth(User), 
				index,
			)
			if err != nil {
				data.Error = err.Error()
				t.Execute(w, data)
				return
			}
			data.Error = "Success cancel"
		}
		if r.FormValue("confirm") != "" {
			_, err := Instance.ConfirmPresent(
				resetAuth(User), 
				index,
			)
			if err != nil {
				data.Error = err.Error()
				t.Execute(w, data)
				return
			}
			data.Error = "Success confirm"
		}
	}
	t.Execute(w, data)
}

func blockchainSalesXPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		TMPL_PATH+"base.html",
		TMPL_PATH+"salesX.html",
	)
	if err != nil {
		panic("can't load hmtl files")
	}
	if User == nil {
		http.Redirect(w, r, "/login", 302)
		return
	}
	var data struct{
		User *UserType
		Block *SaleStr
		Error string
		InCustomers bool
	}
	data.User = User
	var (
		index = new(big.Int)
		ok bool
	)
	num := strings.Replace(r.URL.Path, "/blockchain/sales/", "", 1)
	index, ok = index.SetString(num, 10)
	if !ok {
		data.Error = "strconv error"
		t.Execute(w, data)
		return
	}
	sale := getSales(index)
	if sale == nil {
		data.Error = "sale is nil"
		t.Execute(w, data)
		return
	}
	data.Block = salesToString(sale)
	for _, cust := range data.Block.Customers {
		if cust == User.AddressHex {
			data.InCustomers = true
			break
		}
	}
	if r.Method == "POST" {
		r.ParseForm()
		if r.FormValue("cancel") != "" {
			_, err := Instance.CancelSale(
				resetAuth(User), 
				index,
			)
			if err != nil {
				data.Error = err.Error()
				t.Execute(w, data)
				return
			}
			data.Error = "Success cancel"
		}
		if r.FormValue("confirm") != "" {
			var (
				saleTo = new(big.Int)
				ok bool
			)
			saleTo, ok = saleTo.SetString(r.FormValue("addrnum"), 10)
			if !ok {
				data.Error = "strconv error"
				t.Execute(w, data)
				return
			}
			_, err := Instance.ConfirmSale(
				resetAuth(User), 
				index,
				saleTo,
			)
			if err != nil {
				data.Error = err.Error()
				t.Execute(w, data)
				return
			}
			data.Error = "Success confirm"
		}
		if r.FormValue("buy") != "" {
			var (
				price = new(big.Int)
				ok bool
			)
			price, ok = price.SetString(r.FormValue("price"), 10)
			if !ok {
				data.Error = "strconv error"
				t.Execute(w, data)
				return
			}
			auth := resetAuth(User)
			if auth == nil {
				data.Error = "failed: auth is nil"
				t.Execute(w, data)
				return
			}
			auth.Value = price
			_, err := Instance.CheckToBuy(
				auth, 
				index,
			)
			if err != nil {
				data.Error = err.Error()
				t.Execute(w, data)
				return
			}
			data.Error = "Success buy request"
		}
		if r.FormValue("close") != "" {
			_, err := Instance.CancelToBuy(
				resetAuth(User), 
				index,
			)
			if err != nil {
				data.Error = err.Error()
				t.Execute(w, data)
				return
			}
			data.Error = "Success close buy request"
		}
	}
	t.Execute(w, data)
}

func blockchainRentsXPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		TMPL_PATH+"base.html",
		TMPL_PATH+"rentsX.html",
	)
	if err != nil {
		panic("can't load hmtl files")
	}
	if User == nil {
		http.Redirect(w, r, "/login", 302)
		return
	}
	var data struct{
		User *UserType
		Block *RentStr
		Error string
		Confirmed bool
	}
	data.User = User
	var (
		index = new(big.Int)
		ok bool
	)
	num := strings.Replace(r.URL.Path, "/blockchain/rents/", "", 1)
	index, ok = index.SetString(num, 10)
	if !ok {
		data.Error = "strconv error"
		t.Execute(w, data)
		return
	}
	rent := getRents(index)
	if rent == nil {
		data.Error = "rent is nil"
		t.Execute(w, data)
		return
	}
	data.Block = rentsToString(rent)
	if data.Block.Renter != "0x0000000000000000000000000000000000000000" {
		data.Confirmed = true
	}
	if r.Method == "POST" {
		r.ParseForm()
		if r.FormValue("cancel") != "" {
			_, err := Instance.CancelRent(
				resetAuth(User), 
				index,
			)
			if err != nil {
				data.Error = err.Error()
				t.Execute(w, data)
				return
			}
			data.Error = "Success cancel"
		}
		if r.FormValue("confirm") != "" {
			auth := resetAuth(User)
			if auth == nil {
				data.Error = "failed: auth is nil"
				t.Execute(w, data)
				return
			}
			auth.Value = rent.Money
			_, err := Instance.ToRent(
				auth, 
				index,
			)
			if err != nil {
				data.Error = err.Error()
				t.Execute(w, data)
				return
			}
			data.Error = "Success confirm"
		}
		if r.FormValue("finish") != "" {
			_, err := Instance.FinishRent(
				resetAuth(User), 
				index,
			)
			if err != nil {
				data.Error = err.Error()
				t.Execute(w, data)
				return
			}
			data.Error = "Success finish"
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
	if User == nil {
		http.Redirect(w, r, "/login", 302)
		return
	}
	var data struct{
		User *UserType
		IsAdmin bool
		Error string
	}
	data.User = User
	iamAdmin, err := Instance.IamAdmin(&bind.CallOpts{From: User.AddressEth})
	if err != nil {
		data.Error = err.Error()
		t.Execute(w, data)
		return
	}
	if iamAdmin {
		data.IsAdmin = true
	}
	if r.Method == "POST" {
		r.ParseForm()
		var (
			squere = new(big.Int)
			usefulSquere = new(big.Int)
			ok bool
		)
		squere, ok = squere.SetString(r.FormValue("squere"), 10)
		if !ok {
			data.Error = "strconv error 1"
			t.Execute(w, data)
			return
		}
		usefulSquere, ok = usefulSquere.SetString(r.FormValue("usefulsquere"), 10)
		if !ok {
			data.Error = "strconv error 2"
			t.Execute(w, data)
			return
		}
		_, err := Instance.CreateEstate(
			resetAuth(User), 
			User.AddressEth, 
			r.FormValue("info"),
			squere,
			usefulSquere,
		)
		if err != nil {
			data.Error = err.Error()
			t.Execute(w, data)
			return
		}
		data.Error = "Success created"
	}
	t.Execute(w, data)
}

func blockchainEstatesPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		TMPL_PATH+"base.html",
		TMPL_PATH+"estates.html",
	)
	if err != nil {
		panic("can't load hmtl files")
	}
	if User == nil {
		http.Redirect(w, r, "/login", 302)
		return
	}
	var data struct{
		Error string
		Blocks []uint64
		Address string
		User *UserType
	}
	data.User = User
	data.Address = User.AddressHex
	if r.Method == "POST" {
		data.Address = r.FormValue("address")
	}
	var inc = big.NewInt(1)
	num, err := Instance.GetEstatesNumber(&bind.CallOpts{From: User.AddressEth})
	if err != nil {
		data.Error = err.Error()
		t.Execute(w, data)
		return
	}
	for index := big.NewInt(0); index.Cmp(num) == -1; index.Add(index, inc) {
		block := getEstates(index)
		if block == nil {
			data.Error = "data is nil"
			t.Execute(w, data)
			return
		}
		if data.Address != "all" && strings.ToLower(data.Address) != strings.ToLower(block.Owner.Hex()) {
			continue
		}
		data.Blocks = append(data.Blocks, index.Uint64())
	}
	t.Execute(w, data)
}

func blockchainPresentsPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		TMPL_PATH+"base.html",
		TMPL_PATH+"presents.html",
	)
	if err != nil {
		panic("can't load hmtl files")
	}
	if User == nil {
		http.Redirect(w, r, "/login", 302)
		return
	}
	var data struct{
		Error string
		Blocks []uint64
		Address string
		User *UserType
	}
	data.User = User
	data.Address = User.AddressHex
	if r.Method == "POST" {
		data.Address = r.FormValue("address")
	}
	var inc = big.NewInt(1)
	num, err := Instance.GetPresentsNumber(&bind.CallOpts{From: User.AddressEth})
	if err != nil {
		data.Error = err.Error()
		t.Execute(w, data)
		return
	}
	for index := big.NewInt(0); index.Cmp(num) == -1; index.Add(index, inc) {
		block := getPresents(index)
		if block == nil {
			data.Error = "data is nil"
			t.Execute(w, data)
			return
		}
		if block.Finished {
			continue
		}
		if data.Address != "all" && 
			strings.ToLower(data.Address) != strings.ToLower(block.AddressFrom.Hex()) && 
			strings.ToLower(data.Address) != strings.ToLower(block.AddressTo.Hex()) {
			continue
		}
		data.Blocks = append(data.Blocks, index.Uint64())
	}
	t.Execute(w, data)
}

func blockchainSalesPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		TMPL_PATH+"base.html",
		TMPL_PATH+"sales.html",
	)
	if err != nil {
		panic("can't load hmtl files")
	}
	if User == nil {
		http.Redirect(w, r, "/login", 302)
		return
	}
	var data struct{
		Error string
		Blocks []uint64
		Address string
		User *UserType
	}
	data.User = User
	data.Address = User.AddressHex
	if r.Method == "POST" {
		data.Address = r.FormValue("address")
	}
	var inc = big.NewInt(1)
	num, err := Instance.GetSalesNumber(&bind.CallOpts{From: User.AddressEth})
	if err != nil {
		data.Error = err.Error()
		t.Execute(w, data)
		return
	}
	for index := big.NewInt(0); index.Cmp(num) == -1; index.Add(index, inc) {
		block := getSales(index)
		if block == nil {
			data.Error = "data is nil"
			t.Execute(w, data)
			return
		}
		if block.Finished {
			continue
		}
		if data.Address != "all" && 
			strings.ToLower(data.Address) != strings.ToLower(block.Owner.Hex()) {
			continue
		}
		data.Blocks = append(data.Blocks, index.Uint64())
	}
	t.Execute(w, data)
}

func blockchainRentsPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		TMPL_PATH+"base.html",
		TMPL_PATH+"rents.html",
	)
	if err != nil {
		panic("can't load hmtl files")
	}
	if User == nil {
		http.Redirect(w, r, "/login", 302)
		return
	}
	var data struct{
		Error string
		Blocks []uint64
		Address string
		User *UserType
	}
	data.User = User
	data.Address = User.AddressHex
	if r.Method == "POST" {
		data.Address = r.FormValue("address")
	}
	var inc = big.NewInt(1)
	num, err := Instance.GetRentsNumber(&bind.CallOpts{From: User.AddressEth})
	if err != nil {
		data.Error = err.Error()
		t.Execute(w, data)
		return
	}
	for index := big.NewInt(0); index.Cmp(num) == -1; index.Add(index, inc) {
		block := getRents(index)
		if block == nil {
			data.Error = "data is nil"
			t.Execute(w, data)
			return
		}
		if block.Finished {
			continue
		}
		if data.Address != "all" && 
			strings.ToLower(data.Address) != strings.ToLower(block.Owner.Hex()) && 
			strings.ToLower(data.Address) != strings.ToLower(block.Renter.Hex()) {
			continue
		}
		data.Blocks = append(data.Blocks, index.Uint64())
	}
	
	t.Execute(w, data)
}
