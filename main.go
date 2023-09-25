package main

import (
	"fmt"
    "io"
    "net/http"
    "net/url"
    "strings"
    "golang.org/x/net/html"
	"os"
	"errors"
)

func main() {
    sourceURL := "https://cs272-0304-f23.github.io/tests/top10/" // web-crawler usage
    downloadedURLs := crawl(sourceURL)

    fmt.Println("Downloaded URLs:") // Print the downloaded URLs
    for _, url := range downloadedURLs {
        fmt.Println(url)
    }
}

func extract(body []byte) ([]string, []string) {
    var words []string
    var hrefs []string //Returns 2 list: words && links

    tokenizer := html.NewTokenizer(strings.NewReader(string(body))) //creating HTML tokenizer to process HTML content

    for {
        tokenType := tokenizer.Next() //get next token from tokenizer

        switch tokenType {
        case html.ErrorToken:
            return words, hrefs // When end of the HTML content is reached, return the extracted data(strings && links)
        case html.TextToken:
            wordsInText := strings.Fields(tokenizer.Token().Data) //In case Token is text, split it into words and add them to the words slice
            words = append(words, wordsInText...)
        case html.StartTagToken, html.SelfClosingTagToken:      // Token is a start tag or self-closing tag
            token := tokenizer.Token()
            if token.Data == "a" {
                for _, attr := range token.Attr {
                    if attr.Key == "href" {
                        hrefs = append(hrefs, attr.Val)     // If token is an anchor tag (a), extract the href attribute
                    }
                }
            }
        }
    }
    return words, hrefs
}

// crawl starts the web crawling process from a given source URL.
func crawl(sourceURL string) []string {
    check := make(map[string]bool)
    urlqueue := []string{sourceURL}
    downloadedURLs := []string{}
    for len(urlqueue) > 0 {
        urlWeGot := urlqueue[0]         // Dequeue the first URL
        urlqueue = urlqueue[1:]

        if check[urlWeGot] {
            continue         // Check if the URL is viewed or not
        }

        body, err := download(urlWeGot)         // Download contents
        if err != nil {
            fmt.Printf("download: url=%s, result=error: %v\n", urlWeGot, err)
            continue
        }
        fmt.Printf("download: url=%s, result=ok\n", urlWeGot)

        words, hrefs := extract([]byte(body))         // Extract words and cleaned URLs
        fmt.Println("Words:", words)         // Print the words and cleaned URLs if needed
        fmt.Println("Cleaned URLs:", hrefs)
        check[urlWeGot] = true // Mark this URL as viewed

        // Add all the cleaned URLs to the queue (just for further crawling process)
        for _, cleanedURL := range clean(urlWeGot, hrefs) {
            if !check[cleanedURL] {
                urlqueue = append(urlqueue, cleanedURL)
            }
        }
        downloadedURLs = append(downloadedURLs, urlWeGot)
    }
    return downloadedURLs
}


func clean(host string, hrefs []string) []string {
	var cleanedURLs []string
	hostURL, err := url.Parse(host)

	if err != nil {
		return cleanedURLs 	// Handle the error if parsing the host URL fails
	}

	for _, href := range hrefs {
		parsedURL, err := url.Parse(href) // Parse the href URL
		if err != nil {
			continue // Skip invalid URLs
		}
		resolvedURL := hostURL.ResolveReference(parsedURL) // Resolve the relative URL with the host URL
		cleanedURLs = append(cleanedURLs, resolvedURL.String()) // Add the resolved URL to the cleaned URLs
	}
	return cleanedURLs
}


func download(url string) ([]byte, error) {
    if rsp, err := http.Get(url); err == nil {
        if b, err := io.ReadAll(rsp.Body); err == nil {
            return b, nil
        }
    }
    return []byte{}, nil
}

func splitBook(bookText string) (chapterMap map[string]string, err error) {
    chapterDelimiter := "CHAPTER"
    chapters := strings.Split(bookText, chapterDelimiter)

    if len(chapters) < 2 {
        return nil, errors.New("chapter delimiter not found")
    }

    chapterMap = make(map[string]string)

    for i := 1; i < len(chapters); i++ {
        title := strings.TrimSpace(chapters[i-1 ])
        content := strings.TrimSpace(chapters[i])
        chapterMap[title] = content
    }

    return chapterMap, nil
}