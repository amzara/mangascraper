package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type body struct {
	Url string
}

type reqCookies struct {
	Url     string            `json:"url"`
	Success bool              `json:"success"`
	Cookies map[string]string `json:cookies:`
}

func GetCookies() error {

	// body := `{"url": "https://www.mangakakalot.gg/official"}`

	jsonBOdy := map[string]string{"url": "https://www.mangakakalot.gg/official"} //go syntax for hash map / dictionary

	// jsonBody2 := body{
	// url: "https://www.mangakakalot.gg/official",
	// }

	jsonData, err := json.Marshal(jsonBOdy)

	if err != nil {
		panic(err)
	}

	resp, err := http.Post("http://localhost:8000/getCfCookies", "application/json", bytes.NewReader(jsonData))

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println(resp.Body)
		// responseBody, err := io.ReadAll(resp.Body)
		// if err != nil {
		// 	fmt.Println(err)
		// }
		fmt.Println("This is response body")

		var result reqCookies

		err := json.NewDecoder(resp.Body).Decode(&result)

		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(result.Url)
		fmt.Println(result.Success)
		fmt.Println(result.Cookies)

		//maybe need to declare a struct for it
	}
	return nil
}

//what i've learned? how to handle http requests in golang
//strings, structs, byte, encode, decode

//for sending post request, normally you have to throw in request body
//http.post require a io.reader, so you can do 2 way
//1) whack a string in the form of `{}` into String package reader
//2) use map[string][string]{json} to form a byte slice and put into bytes.newReader

//to decode json request(from byte -> meaningful data you can use in form of struct)
//if you know the struct type, just create a struct object of that type, and use json.NewDEcoder(response.body).Decode(&struct object), the pointer is so you can modify the data of that struct
