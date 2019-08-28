package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	r, err := http.Get("https://www.bankmega.com/promolainnya.php")
	if err != nil {
		log.Fatalf("could not fetch URL")
	}
	doc, err := goquery.NewDocumentFromResponse(r)
	if err != nil {
		log.Fatalf("Failed to parse Document")
	}
	ids := findIds(doc)
	fmt.Println(ids)
	getJS(doc)
}

func findIds(doc *goquery.Document) []string {
	ids := []string{}
	promoSelector := doc.Find("#subcatpromo").Find("img")
	promoSelector.Each(func(i int, s *goquery.Selection) {
		attr, exist := s.Attr("id")
		if exist {
			ids = append(ids, attr)
		}
	})
	return ids
}

func getJS(doc *goquery.Document) string {
	js := doc.Find("#contentpromolain2").Find("script").Text()
	fmt.Println(js)
	return js
}

func getURL(idPromo string) string {
	return ""
}
