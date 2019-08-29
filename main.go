package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type item struct {
	Title   string `json:"title"`
	Area    string `json:"areaPromo"`
	Periode string `json:"periodePromo"`
	Gambar  string `json:"imgUrl"`
	URL     string `json:"urlPromo"`
	cat     category
}

type category struct {
	title, id, url string
}

type itemURL struct {
	cat category
	url string
}

var baseURL = "https://www.bankmega.com/"

func main() {
	var (
		numOfConcurency = flag.Int("n", 5, "Number of concurency.")
		output          = flag.String("o", "solution.json", "Output Filename.")
		stdoutput       = flag.Bool("stdout", false, "write to stdout instead of file")
	)
	flag.Parse()

	r, err := http.Get(baseURL + "promolainnya.php")
	if err != nil {
		log.Fatalf("could not fetch URL promolainnya.php :%s", err.Error())
	}

	doc, err := goquery.NewDocumentFromResponse(r)
	if err != nil {
		log.Fatalf("Failed to parse Document : %s", err.Error())
	}

	// meow
	cats := getCategories(doc)
	js := getJS(doc)

	for i := range cats {
		cats[i].url = baseURL + getURL(js, cats[i].id)
	}

	itemurls := make(chan itemURL)
	go func() {
		for _, c := range cats {
			fetchItemURL(itemurls, c)
		}
		close(itemurls)
	}()

	var wg sync.WaitGroup
	items := make(chan item)

	wg.Add(*numOfConcurency)
	go func() {
		wg.Wait()
		close(items)
	}()

	for i := 0; i < *numOfConcurency; i++ {
		go func() {
			defer wg.Done()
			for u := range itemurls {
				url := u.url
				itm, err := fetchItem(url)
				itm.cat = u.cat
				if err != nil {
					log.Println("could not fetch items", err.Error())
					return
				}
				items <- itm
			}
		}()
	}

	res := make(map[string][]item)
	for i := range items {
		res[i.cat.title] = append(res[i.cat.title], i)
	}
	if err != nil {
		log.Fatalf("failed to write json file : %s", err.Error())
	}
	e := json.NewEncoder(os.Stdout)
	if !*stdoutput {
		outputFile, err := os.OpenFile(*output, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			log.Fatalf("could not open file %s", *output)
		}
		defer outputFile.Close()
		e = json.NewEncoder(outputFile)
	}
	e.SetIndent("", "    ")
	if e.Encode(res) != nil {
		log.Fatalf("failed to write JSON File")
	}
}

func getCategories(doc *goquery.Document) []category {
	cats := []category{}
	promoSelector := doc.Find("#subcatpromo").Find("img")
	promoSelector.Each(func(i int, s *goquery.Selection) {
		attr, exist := s.Attr("id")
		if title, ok := s.Attr("title"); exist && ok {
			cats = append(cats, category{title: title, id: attr})
		}
	})
	return cats
}

func getJS(doc *goquery.Document) string {
	js := doc.Find("#contentpromolain2").Find("script").Text()
	return js
}

func getURL(js, idPromo string) string {
	search := fmt.Sprintf("$(\"#%s\").click(function(){", idPromo)
	searchIdx := strings.Index(js, search)
	foundFunc := js[searchIdx:]
	openBracket := strings.Index(foundFunc, "{")
	closeBracket := strings.Index(foundFunc, "}")
	funcBody := foundFunc[openBracket:closeBracket]
	urlIdx := strings.Index(funcBody, ".load") + 5
	url := strings.TrimSpace(funcBody[urlIdx:])
	return strings.Trim(url, "()\";")
}

func fetchItemURL(u chan itemURL, c category) {
	r, err := http.Get(c.url)
	if err != nil {
		log.Print("could not fetch category URL", err.Error())
		return
	}
	doc, err := goquery.NewDocumentFromResponse(r)
	if err != nil {
		log.Printf("failed to parse Document")
		return
	}
	p, err := getTotalPage(doc)
	if err != nil {
		log.Printf("failed to get total page : %s", err.Error())
		return
	}
	for i := 1; i <= p; i++ {
		fetchPage(u, c, i)
	}
}

func getTotalPage(d *goquery.Document) (int, error) {
	p := 0
	a, err := d.Find(".tablepaging").Html()
	if err != nil {
		return 0, fmt.Errorf("could net get page information")
	}
	idx := strings.Index(a, "Page 1 of ")
	pageTitle := a[idx:]
	quoteidx := strings.Index(pageTitle, "\"")
	pageInfo := strings.Split(pageTitle[:quoteidx], " ")
	p, err = strconv.Atoi(pageInfo[3])
	return p, err
}

func fetchPage(u chan itemURL, c category, page int) {
	pageParam := "&page=" + strconv.Itoa(page)
	r, err := http.Get(c.url + pageParam)
	if err != nil {
		log.Print("could not fetch category URL")
		return
	}
	doc, err := goquery.NewDocumentFromResponse(r)
	if err != nil {
		log.Printf("Failed to parse Document")
		return
	}
	s := doc.Find("#promolain").Find("li")
	s.Each(func(i int, s *goquery.Selection) {
		s = s.Find("a")
		href, ok := s.Attr("href")
		if !ok {
			log.Printf("Could not get href")
		}
		u <- itemURL{
			cat: c,
			url: href,
		}
	})
}

func fetchItem(url string) (item, error) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = baseURL + url
	}
	r, err := http.Get(url)
	if err != nil {
		log.Print("could not fetch category URL", url)
		return item{}, nil
	}
	doc, err := goquery.NewDocumentFromResponse(r)
	if err != nil {
		log.Printf("Failed to parse Document")
		return item{}, nil
	}
	title := doc.Find(".titleinside").Find("h3").Text()
	area := doc.Find(".area").Find("b").Text()
	p := []string{}
	doc.Find(".periode").Find("b").Each(func(i int, s *goquery.Selection) {
		p = append(p, s.Text())
	})
	periode := strings.Join(p, "")
	gambar, _ := doc.Find(".keteranganinside").Find("img").Attr("src")
	return item{
		Title:   title,
		Area:    area,
		Periode: periode,
		Gambar:  gambar,
		URL:     url,
	}, nil
}
