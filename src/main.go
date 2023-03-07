package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	//sleep time in sec
	sleepTime = 30
)

type Scraper struct {
	session   *http.Client
	addresses []string
	proxies   []string
}

func NewScraper() *Scraper {
	scraper := &Scraper{}
	scraper.session = &http.Client{}
	scraper.addresses = readLines("./addresses.txt")
	scraper.proxies = readLines("./proxies.txt")
	return scraper
}

func (scraper *Scraper) Scrape() {
	for _, address := range scraper.addresses {
		var proxyURL, err = url.Parse("http://" + scraper.getRandomProxy())
		if err != nil {
			fmt.Printf("\nError parsing proxy url: %v", err)
			continue
		}

		scraper.session = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
			Timeout: sleepTime * time.Second,
		}

		response, err := scraper.session.Get(fmt.Sprintf("https://opensea.io/%s", address))
		if err != nil {
			fmt.Printf("\nError fetching %s: %v", address, err)
			continue
		}

		if response.StatusCode == 200 {
			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				fmt.Printf("\nError reading response body: %v", err)
				continue
			}

			if strings.Contains(string(body), "https://twitter.com/") {
				doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
				if err != nil {
					fmt.Printf("\nError parsing HTML: %v", err)
					continue
				}

				doc.Find("a.sc-1f719d57-0.fKAlPV").Each(func(_ int, s *goquery.Selection) {
					twitter := strings.Replace(s.AttrOr("href", ""), "https://twitter.com/", "", -1)
					if !strings.Contains(twitter, "@opensea") {
						fmt.Printf("[+] %s -> %s\n", address, twitter)
						writeToFile("./twitters.txt", twitter+"\n")
					} else {
						fmt.Printf("[-] %s -> Twitter Not Found\n", address)
					}
				})
			} else {
				fmt.Printf("[-] %s -> Twitter Not Found\n", address)
			}
		} else {
			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				fmt.Printf("Error reading response body: %v", err)
				continue
			}

			fmt.Printf("Error fetching %s: %s", address, body)
		}
	}
}

func (scraper *Scraper) getRandomProxy() string {
	rand.Seed(time.Now().UnixNano())
	return scraper.proxies[rand.Intn(len(scraper.proxies))]
}

func readLines(filename string) []string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v", err)
		return nil
	}

	lines := strings.Split(strings.Replace(string(content), "\r", "\n", -1), "\n")

	return lines
}

func writeToFile(filename string, content string) {
	err := ioutil.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		fmt.Printf("Error writing to file: %v", err)
	}
}

func main() {
	scraper := NewScraper()
	scraper.Scrape()
}
