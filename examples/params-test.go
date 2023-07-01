package main

import (
	"fmt"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
)

func main() {
	req := url.NewRequest()
	//params := map[string]string{
	//	"page":  "1",
	//	"limit": "20",
	//	"skip":  "5",
	//}
	//params := map[string]int{
	//	"page":  1,
	//	"limit": 20,
	//	"skip":  5,
	//}
	//params := map[string][]string{
	//	"page":  []string{"1", "2"},
	//	"limit": []string{"20"},
	//	"skip":  []string{"5"},
	//}
	//params := map[string][]int{
	//	"page":  []int{1, 2},
	//	"limit": []int{20},
	//	"skip":  []int{5},
	//}
	params := map[string]interface{}{
		"page":  []interface{}{"1", 2},
		"limit": []string{"20"},
		"skip":  []int{5},
	}
	req.Params = url.ParseParams(params)
	r, err := requests.Get("https://httpbin.org/get", req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(r.Text)
}
