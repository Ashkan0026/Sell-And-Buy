package db

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Ashkan0026/sell-the-old/models"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

var db *sql.DB

func Init() error {
	err := godotenv.Load(".env")
	if err != nil {
		return err
	}
	host, user, password, dbname, port := os.Getenv("host"), os.Getenv("user"), os.Getenv("password"), os.Getenv("dbname"), 5432
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err = sql.Open("postgres", psqlconn)
	if err != nil {
		return err
	}
	return nil
}

func CreateUserTable() error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS users(" +
		"id bigserial not null primary key," +
		"username VARCHAR(50) not null," +
		"password VARCHAR(90) not null," +
		"email VARCHAR(50) not null," +
		"phonenumber VARCHAR(20) not null," +
		"putted integer ARRAY," +
		"bought integer ARRAY," +
		"role VARCHAR(10) not null" +
		")")
	if err != nil {
		return err
	}
	return nil
}

func CreateProductTable() error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS products(" +
		"id bigserial not null primary key," +
		"userId bigserial not null," +
		"title VARCHAR(30) not null," +
		"description VARCHAR(100) not null," +
		"address VARCHAR(50) not null," +
		"price integer not null," +
		"bought boolean not null," +
		"imgURL VARCHAR(100) not null" +
		")")
	if err != nil {
		return err
	}
	return nil
}

func GetDB() *sql.DB {
	return db
}

func ReadUsersRows(rows *sql.Rows) []models.User {
	defer rows.Close()
	usrs := []models.User{}
	for rows.Next() {
		usr := models.User{}
		err := rows.Scan(&usr.ID, &usr.Username, &usr.Passwrod, &usr.Email, &usr.Phonenumber, &usr.Putted, &usr.Bought, &usr.Role)
		if err != nil {
			panic(err)
		}
		usrs = append(usrs, usr)
	}
	return usrs
}

func UserExists(username string, password string) (*models.User, error) {
	newPassword := base64.StdEncoding.EncodeToString([]byte(password))
	rows, err := db.Query("SELECT * FROM users WHERE username = $1 AND password = $2", username, newPassword)
	if err != nil {
		return nil, err
	}
	usrs := ReadUsersRows(rows)
	if len(usrs) == 0 {
		return nil, fmt.Errorf("Such user doesn't exist")
	}
	return &usrs[0], nil
}

func UserExistsWithUsername(username string) *models.Error {
	rows, err := db.Query("SELECT * FROM users WHERE username = $1", username)
	if err != nil {
		log.Printf("%s\n", err.Error())
		return &models.Error{StatusCode: 500, Message: "Server Error"}
	}
	usrs := ReadUsersRows(rows)
	if len(usrs) > 0 {
		return &models.Error{StatusCode: 403, Message: "There is a User with this username"}
	}
	return nil
}

func UserExistsWithPassword(password string) *models.Error {
	rows, err := db.Query("SELECT * FROM users WHERE password = $1", password)
	if err != nil {
		log.Printf("%s\n", err.Error())
		return &models.Error{StatusCode: 500, Message: "Server Error"}
	}
	usrs := ReadUsersRows(rows)
	if len(usrs) > 0 {
		return &models.Error{StatusCode: 403, Message: "There is a User with this password"}
	}
	return nil
}

func UserExistsWithEmail(email string) *models.Error {
	rows, err := db.Query("SELECT * FROM users WHERE email = $1", email)
	if err != nil {
		log.Printf("%s\n", err.Error())
		return &models.Error{StatusCode: 500, Message: "Server Error"}
	}
	usrs := ReadUsersRows(rows)
	if len(usrs) > 0 {
		return &models.Error{StatusCode: 403, Message: "There is a User with this email"}
	}
	return nil
}

func InsertUserToTable(username, password, email, phonenumber string) (*models.User, *models.Error) {
	newPassword := base64.StdEncoding.EncodeToString([]byte(password))
	rows, err := db.Query("INSERT INTO users (username, password, email, phonenumber, putted, bought, role) VALUES ($1, $2, $3, $4, '{}', '{}', 'user') RETURNING *", username, newPassword, email, phonenumber)
	if err != nil {
		return nil, &models.Error{StatusCode: 500, Message: err.Error()}
	}
	usrs := ReadUsersRows(rows)
	if len(usrs) == 0 {
		return nil, &models.Error{StatusCode: 500, Message: "User didn't add"}
	}
	return &usrs[0], nil
}

func InsertProductIntoTable(price int32, userId int64, address, title, decription, imgUrl string) *models.Error {
	errSt := &models.Error{}
	rows, err := db.Query("INSERT INTO products (userid, title, description, address, price, bought,imgurl) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id", userId, title, decription, address, price, false, imgUrl)
	if err != nil {
		errSt.Message = "Error in writing in the database"
		errSt.StatusCode = 403
		return errSt
	}
	prdId := 0
	for rows.Next() {
		err = rows.Scan(&prdId)
		if err != nil {
			log.Printf("Error in Scanning, %s\n", err.Error())
			errSt.Message = err.Error()
			errSt.StatusCode = 404
			return errSt
		}
	}
	_, err = db.Exec("UPDATE users SET putted = array_append(putted,$1) WHERE id = $2", prdId, userId)
	if err != nil {
		log.Printf("Error in adding product to user ")
		errSt.Message = err.Error()
		errSt.StatusCode = 403
		return errSt
	}
	return nil
}

func ReadProductsFromDB() error {
	rows, err := db.Query("SELECT * FROM products")
	if err != nil {
		return err
	}
	for rows.Next() {
		product := &models.Product{}
		err = rows.Scan(&product.ID, &product.UserID, &product.Title, &product.Description, &product.Address, &product.Price, &product.Bought, &product.ImgURL)
		if err != nil {
			return err
		}
		product.ImgURL = product.ImgURL[1:]
		products = append(products, product)
	}
	return nil
}

func AddBoughtToUserList(productId, userId int64) error {
	var putteds pq.Int64Array
	err := db.QueryRow("SELECT putted FROM users WHERE id=$1", userId).Scan(&putteds)
	for _, putt := range putteds {
		if putt == productId {
			return errors.New("You can not buy your own product")
		}
	}
	_, err = db.Exec("UPDATE users SET bought = array_append(bought, $1) WHERE id=$2", productId, userId)
	if err != nil {
		return err
	}
	SetProductAsBought(productId)
	return nil
}

func SetProductAsBought(productId int64) {
	_, err := db.Exec("UPDATE products SET bought = $1 WHERE id=$2", true, productId)
	if err != nil {
		log.Printf("Error : %s\n", err.Error())
		return
	}
}
