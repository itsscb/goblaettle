package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

const baseURI = "https://www.rottenacker.de/"

func main() {

	// Set the path to the Chromium binary on your Raspberry Pi.
	execPath, err := exec.LookPath("chromium-browser")
	if err != nil {
		log.Fatalf("Chromium binary not found: %v", err)
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.ExecPath(execPath),
	)

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	// Navigate to the URL and wait for the page to load.
	var responseText string
	err = chromedp.Run(ctx,
		chromedp.Navigate(baseURI+"cgi-seiten/amtsblatt.htm"), // Replace with the URL you want to crawl.
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Evaluate(`document.documentElement.innerHTML`, &responseText),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Parse the HTML content using goquery.
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(responseText))
	if err != nil {
		log.Fatal(err)
	}

	// Find and extract the data you need from the parsed HTML using goquery selectors.
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if exists {
			nameRAW := strings.Split(link, "/")
			name := nameRAW[len(nameRAW)-1]
			link = baseURI + strings.Replace(link, "../", "", -1)
			downloadFile(link, "mitteilungsblatt_"+name)
		}
	})
}

func downloadFile(url string, filepath string) error {
	if fileExists(filepath) {
		fmt.Printf("File %s already exists. Skipping download.\n", filepath)
		return nil
	}

	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file, status code: %d", response.StatusCode)
	}

	// Determine the file extension based on the Content-Type header
	// contentType := response.Header.Get("Content-Type")
	// ext := ".txt" // Default extension if Content-Type is not present or not recognized
	// if strings.Contains(contentType, "pdf") {
	// 	ext = ".pdf"
	// } else if strings.Contains(contentType, "jpeg") {
	// 	ext = ".jpeg"
	// } // Add more cases for other file types as needed

	// Create the file with the appropriate extension
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	fmt.Println("File downloaded successfully!")
	return nil
}
func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}
