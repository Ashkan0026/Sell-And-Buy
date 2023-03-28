package utils

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func ParseRequest(r *http.Request, x interface{}) error {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, x)
	if err != nil {
		return err
	}
	return nil
}
