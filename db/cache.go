package db

import (
	"errors"

	"github.com/Ashkan0026/sell-the-old/models"
	"github.com/gorilla/sessions"
)

var cookies *sessions.CookieStore
var products []*models.Product

func InitSessions(key string) {
	cookies = sessions.NewCookieStore([]byte(key))
	cookies.Options.Domain = "localhost"
	cookies.Options.Secure = true
	cookies.Options.Path = "/"
}

func GetCookies() *sessions.CookieStore {
	return cookies
}

func EmptyProducts() {
	products = nil
}

func GetProducts() []*models.Product {
	return products
}

func GetProduct(id int64) (*models.Product, error) {
	if products == nil {
		return nil, errors.New("products is empty")
	}
	for _, pr := range products {
		if pr.ID == id {
			return pr, nil
		}
	}
	return nil, errors.New("There isn't such a product")
}
