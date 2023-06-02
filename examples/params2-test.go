package main

import (
	"fmt"
	"github.com/wangluozhe/requests/url"
)

func main() {
	h := url.NewValues()
	h.Set("a", "a")
	h.Set("a", "b")
	h.Set("b", "c")
	h.Add("a", "c")
	h.Set("c", "d")
	fmt.Println("value:", h.Get("a"))
	fmt.Println("values:", h.Values())
	fmt.Println("keys:", h.Keys())
	//h.Set("a", "d")
	//fmt.Println("value:", h.Get("a"))
	//fmt.Println("values:", h.Values())
	//fmt.Println("keys:", h.Keys())
	fmt.Println("encode:", h.Encode())
	params := url.NewParams()
	params.Add("ip", "127.0.0.1")
	params.Add("ip1", "127.0.0.1")
	params.Add("ip", "127.0.0.2")
	params.Add("ip2", "127.0.0.2")
	fmt.Println(params.Encode())
}
