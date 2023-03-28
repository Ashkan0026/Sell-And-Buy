package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Ashkan0026/sell-the-old/db"
	"github.com/Ashkan0026/sell-the-old/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
	//Initialize the database
	db.Init()

	err = db.CreateUserTable()

	if err != nil {
		log.Printf("Error : %s", err.Error())
		return
	}

	db.InitSessions(os.Getenv("SESSION_KEY"))

	err = db.CreateProductTable()

	if err != nil {
		log.Printf("Error : %s", err.Error())
		return
	}

	err = db.ReadProductsFromDB()

	if err != nil {
		log.Printf("Error : %s", err.Error())
		return
	}

	//Build the server with its handler
	handlr := Handler()
	handlr.PathPrefix("/resources/").Handler(http.StripPrefix("/resources/", http.FileServer(http.Dir("./resources"))))

	server := &http.Server{
		Addr:         os.Getenv("PORT"),
		ReadTimeout:  12 * time.Second,
		WriteTimeout: 20 * time.Second,
		Handler:      handlr,
	}

	log.Printf("Server is running on PORT %s\n", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}

}

func Handler() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/home", handlers.HomePage).Methods("GET")
	router.HandleFunc("/auth/login", handlers.VisitLoginPage).Methods("GET")
	router.HandleFunc("/auth/signup", handlers.VisitSignupPage).Methods("GET")
	router.HandleFunc("/auth/login", handlers.HandlerLogin).Methods("POST")
	router.HandleFunc("/auth/signup", handlers.SingUpHandler).Methods("POST")
	router.HandleFunc("/productInsert", handlers.HandleProductInsertPage).Methods("GET")
	router.HandleFunc("/productInsert", handlers.GetAndSaveProduct).Methods("POST")
	router.HandleFunc("/products/{id}", handlers.EachProductPage).Methods("GET")

	return router
}
