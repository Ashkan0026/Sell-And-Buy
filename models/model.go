package models

type User struct {
	ID          int64   `json:"id"`
	Username    string  `json:"username"`
	Passwrod    string  `json:"password"`
	Email       string  `json:"email"`
	Phonenumber string  `json:"phonenumber"`
	Putted      []uint8 `json:"putted"`
	Bought      []uint8 `json:"bought"`
	Role        string  `json:"role"`
}

type Product struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"userId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Address     string `json:"address"`
	Price       int32  `json:"price"`
	Bought      bool   `json:"bought"`
	ImgURL      string `json:"imgUrl"`
}

type HomeData struct {
	Usr      *User      `json:"usr"`
	Products []*Product `json:"products"`
}

type Error struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

type Response struct {
	Usr *User `json:"user"`
}
