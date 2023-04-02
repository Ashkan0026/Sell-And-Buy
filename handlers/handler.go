package handlers

import (
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/Ashkan0026/sell-the-old/db"
	"github.com/Ashkan0026/sell-the-old/models"
	"github.com/Ashkan0026/sell-the-old/utils"
	"github.com/gorilla/mux"
)

func HomePage(w http.ResponseWriter, r *http.Request) {
	cookies := db.GetCookies()
	session, err := cookies.Get(r, "go-sessions")
	if err != nil {
		log.Printf("Error in session %s", err.Error())
		return
	}
	authenticated := session.Values["authenticated"]
	if authenticated != nil {
		isAuth := session.Values["authenticated"].(bool)
		if !isAuth {
			utils.ExecFile("home", w, nil)
		} else {
			username, email := session.Values["username"].(string), session.Values["email"].(string)
			res := &models.HomeData{Usr: &models.User{Username: username, Email: email}}
			res.Products = db.GetProducts()
			utils.ExecFile("home", w, res)
		}
	} else {
		utils.ExecFile("home", w, nil)
	}
}

func VisitLoginPage(w http.ResponseWriter, r *http.Request) {
	utils.ExecFile("login", w, nil)
}

func VisitSignupPage(w http.ResponseWriter, r *http.Request) {
	utils.ExecFile("signup", w, nil)
}

func HandlerLogin(w http.ResponseWriter, r *http.Request) {
	usr := &models.User{}
	r.ParseForm()
	usr.Username = r.Form.Get("username")
	usr.Passwrod = r.Form.Get("password")
	usr, err := db.UserExists(usr.Username, usr.Passwrod)
	if err != nil {
		log.Printf("%s", err.Error())
		errSt := models.Error{StatusCode: 404, Message: err.Error()}
		utils.ExecFile("login", w, errSt)
		return
	}

	cookies := db.GetCookies()
	session, err := cookies.Get(r, "go-sessions")
	if err != nil {
		log.Printf("Error %s", err.Error())
		return
	}

	session.Values["authenticated"] = true
	session.Values["username"] = usr.Username
	session.Values["email"] = usr.Email
	session.Values["userid"] = usr.ID
	err = session.Save(r, w)
	if err != nil {
		log.Printf("Error happend %s", err.Error())
		return
	}
	utils.ExecFile("login", w, &models.Error{StatusCode: 200, Message: "Successful login"})
}

func SingUpHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username, password, email, phonenumber := r.Form.Get("username"), r.Form.Get("password"), r.Form.Get("email"), r.Form.Get("phonenumber")
	err := db.UserExistsWithUsername(username)
	if err != nil {
		utils.ExecFile("signup", w, err)
		return
	}
	err = db.UserExistsWithPassword(password)
	if err != nil {
		utils.ExecFile("signup", w, err)
		return
	}
	err = db.UserExistsWithEmail(email)
	if err != nil {
		utils.ExecFile("signup", w, err)
		return
	}
	err = EmailValidation(email)
	if err != nil {
		utils.ExecFile("signup", w, err)
		return
	}
	err = PhonenumberValidation(phonenumber)
	if err != nil {
		utils.ExecFile("signup", w, err)
		return
	}
	usr, err := db.InsertUserToTable(username, password, email, phonenumber)
	if err != nil {
		utils.ExecFile("signup", w, err)
		return
	}
	cookies := db.GetCookies()
	session, errs := cookies.Get(r, "go-sessions")
	if errs != nil {
		log.Printf("Error %s", errs.Error())
		return
	}
	session.Values["authenticated"] = true
	session.Values["username"] = usr.Username
	session.Values["email"] = usr.Email
	session.Values["userid"] = usr.ID
	errs = session.Save(r, w)
	if errs != nil {
		log.Printf("Error happend %s", errs.Error())
		return
	}
	utils.ExecFile("signup", w, &models.Error{StatusCode: 200, Message: "Successful signup"})
}

func EmailValidation(email string) *models.Error {
	res1, err := regexp.Match("[a-zA-Z0-9]{4,}@[g|e]mail.com", []byte(email))
	errSt := &models.Error{StatusCode: 403, Message: "Failed at email validation"}
	if err != nil {
		log.Printf("%s\n", err.Error())
		return errSt
	}
	if res1 {
		return nil
	}
	res1, err = regexp.Match("[a-zA-Z0-9]{4,}.com@outlook.com", []byte(email))
	if err != nil {
		log.Printf("%s\n", err.Error())
		return errSt
	}
	if res1 {
		return nil
	}
	return errSt
}

func PhonenumberValidation(phonenumber string) *models.Error {
	res, err := regexp.Match("[0][9][1|3|9|0][0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]", []byte(phonenumber))
	errSt := &models.Error{StatusCode: 403, Message: "Failed in phonenumber validation"}
	if err != nil {
		log.Printf("%s\n", err.Error())
		return errSt
	}
	if !res {
		return errSt
	}
	return nil
}

func HandleProductInsertPage(w http.ResponseWriter, r *http.Request) {
	utils.ExecFile("productInsertion", w, nil)
}

func GetAndSaveProduct(w http.ResponseWriter, r *http.Request) {
	cookies := db.GetCookies()
	session, err := cookies.Get(r, "go-sessions")
	if err != nil {
		log.Printf("Error : " + err.Error())
		return
	}
	authenticated := session.Values["authenticated"]
	if authenticated == nil {
		err := &models.Error{StatusCode: 403, Message: "You haven't logged in yet"}
		utils.ExecFile("productInsertion", w, err)
		return
	}
	isAuth := session.Values["authenticated"].(bool)
	if !isAuth {
		err := &models.Error{StatusCode: 403, Message: "You haven't logged in yet"}
		utils.ExecFile("productInsertion", w, err)
		return
	}
	userId := session.Values["userid"].(int64)
	r.ParseMultipartForm(10 << 20)
	title, address, description, priceStr := r.MultipartForm.Value["title"][0], r.MultipartForm.Value["address"][0], r.MultipartForm.Value["description"][0], r.MultipartForm.Value["price"][0]
	price, err := strconv.Atoi(priceStr)
	if err != nil {
		errSt := &models.Error{StatusCode: 403, Message: "Error in parsing the price"}
		utils.ExecFile("productInsertion", w, errSt)
		return
	}
	file, handler, err := r.FormFile("image")

	if file == nil {
		errSt := &models.Error{StatusCode: 404, Message: "No file uploaded"}
		utils.ExecFile("productInsertion", w, errSt)
		return
	}
	if err != nil {
		errSt := &models.Error{StatusCode: 404, Message: "Error in file"}
		utils.ExecFile("productInsertion", w, errSt)
		return
	}
	defer file.Close()
	imgUrl := "./resources/" + handler.Filename
	dst, err := os.Create(imgUrl)
	if err != nil {
		log.Printf("Error in creating image file")
		return
	}
	_, err = io.Copy(dst, file)
	if err != nil {
		log.Printf("Error in copying image file")
		return
	}
	if title == "" || address == "" || description == "" {
		errSt := &models.Error{StatusCode: 403, Message: "Some fields are empty"}
		utils.ExecFile("productInsertion", w, errSt)
		return
	}
	errSt := db.InsertProductIntoTable(int32(price), userId, address, title, description, imgUrl)
	if errSt != nil {
		utils.ExecFile("productInsertion", w, errSt)
		return
	}
	db.EmptyProducts()
	_ = db.ReadProductsFromDB()
	success := &models.Error{StatusCode: 200, Message: "Product putted here successfully"}
	utils.ExecFile("productInsertion", w, success)
}

func EachProductPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, exists := vars["id"]
	if !exists {
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("%s", err.Error())
		return
	}
	pr, err := db.GetProduct(int64(id))
	if err != nil {
		err = db.ReadProductsFromDB()
		if err != nil {
			log.Printf("Server error")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		pr, _ = db.GetProduct(int64(id))
	}
	prAndError := &models.ProductAndError{Prd: pr}
	utils.ExecFile("product", w, prAndError)
}

func BuyAProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, exists := vars["id"]
	if !exists {
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Error : %s", err.Error())
		return
	}
	product, err := db.GetProduct(int64(id))
	if err != nil {
		log.Printf("Error : %s\n", err.Error())
		return
	}
	cookies := db.GetCookies()
	session, err := cookies.Get(r, "go-sessions")
	if err != nil {
		prAndErr := &models.ProductAndError{Err: &models.Error{StatusCode: 404, Message: err.Error()}, Prd: product}
		utils.ExecFile("product", w, prAndErr)
		return
	}
	authen := session.Values["authenticated"]
	if authen == nil {
		prAndErr := &models.ProductAndError{Err: &models.Error{StatusCode: 403, Message: "You haven't logged in yet"}, Prd: product}
		utils.ExecFile("product", w, prAndErr)
		return
	}
	isAuthen := session.Values["authenticated"].(bool)
	if !isAuthen {
		prAndErr := &models.ProductAndError{Err: &models.Error{StatusCode: 403, Message: "User has logged out\n"}, Prd: product}
		utils.ExecFile("product", w, prAndErr)
		return
	}
	userId := session.Values["userid"].(int64)
	err = db.AddBoughtToUserList(int64(id), userId)
	if err != nil {
		prAndErr := &models.ProductAndError{Err: &models.Error{StatusCode: 500, Message: err.Error()}, Prd: product}
		utils.ExecFile("product", w, prAndErr)
		return
	}
	pr, err := db.GetProduct(int64(id))
	if err != nil {
		log.Printf("Error : %s", err.Error())
	}
	prAndError := &models.ProductAndError{Prd: pr, Err: &models.Error{StatusCode: 200, Message: "Product added to your bought list"}}
	utils.ExecFile("product", w, prAndError)
}
