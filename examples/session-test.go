package main

import (
	"fmt"
	"github.com/wangluozhe/requests"
)

func main() {
	session := requests.NewSession()
	response, err := session.Get("https://ipinfo.io", nil)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(response.Text)
}
