package main

import (
	"fmt"
	"io"
    "testing"
    "net/http/httptest"
    "io/ioutil"
    "strings"
    "os"
    "hash/fnv"
)
func TestCrawl(t *testing.T) {
    urlToTest := "https://cs272-0304-f23.github.io/tests/top10/"
    getURL := crawl(urlToTest)
    urlsWanted := []string{
		"https://cs272-0304-f23.github.io/tests/top10/1513/1513-h.htm",
		"https://cs272-0304-f23.github.io/tests/top10/84/84-h.htm",
		"https://cs272-0304-f23.github.io/tests/top10/6130/6130-h.htm",
		"https://cs272-0304-f23.github.io/tests/top10/11/11-h.htm",
		"https://cs272-0304-f23.github.io/tests/top10/174/174-h.htm",
		"https://cs272-0304-f23.github.io/tests/top10/345/345-h.htm",
		"https://cs272-0304-f23.github.io/tests/top10/5200/5200-h.htm",
		"https://cs272-0304-f23.github.io/tests/top10/98/98-h.htm",
		"https://cs272-0304-f23.github.io/tests/top10/43/43-h.htm",
		"https://cs272-0304-f23.github.io/tests/top10/1232/1232-h.htm",
    }

    if !reflect.DeepEqual(getURL, urlsWanted) {
        t.Errorf("Something went wrong with URL cleaning. Got: %v, Want: %v", getURL, urlsWanted)
    }
}

func TestClean(t *testing.T) {
	// Add your tests for the clean function here if needed.
    testCases := []struct {
        host     string
        hrefs    []string
        expected []string
    }{
        // Test case 1: Valid host and relative URLs
        {
            host:     "https://cs272-0304-f23.github.io/tests/top10/",
            hrefs:    []string{"about", "contact", "https://example.com/news"},
            expected: []string{"https://example.com/about", "https://example.com/contact", "https://example.com/news"},
        },
        // Test case 2: Invalid host (parse error)
        {
            host:     "invalidhost",
            hrefs:    []string{"about", "contact"},
            expected: []string{},
        },
        // Test case 3: Invalid relative URLs (parse error)
        {
            host:     "https://example.com",
            hrefs:    []string{"about", "http://invalid-url", "contact"},
            expected: []string{"https://example.com/about", "https://example.com/contact"},
        },
    }

    for _, tc := range testCases {
        t.Run(tc.host, func(t *testing.T) {
            cleanedURLs := clean(tc.host, tc.hrefs)

            if len(cleanedURLs) != len(tc.expected) {
                t.Errorf("Expected %d cleaned URLs, but got %d", len(tc.expected), len(cleanedURLs))
            }

            for i, cleanedURL := range cleanedURLs {
                if cleanedURL != tc.expected[i] {
                    t.Errorf("Expected cleaned URL '%s', but got '%s'", tc.expected[i], cleanedURL)
                }
            }
        })
    }
}

func TestDownload(t *testing.T) {
		// Use the live web page URL for testing
		urlToTest := "https://cs272-0304-f23.github.io/tests/lab03/"
		body, err := download(urlToTest)
		if err != nil {
			t.Errorf("download: expected no error, got %v", err)
		}
	
		expectedBodyContains := []byte(`<html>
	
	<body>
	
	<ul>
	
	<li>
	
	<a href="/tests/lab03/simple.html">simple.html</a>
	
	</li>
	
	<li>
	
	<a href="/tests/lab03/href.html">href.html</a>
	
	</li>
	
	<li>
	
	<a href="/tests/lab03/style.html">style.html</a>
	
	</ul>
	
	</body>
	
	</html>
	`)
	
		if !bytes.Equal(body, expectedBodyContains) {
			t.Errorf("download: expected body to contain '%s', got '%s'", expectedBodyContains, body)
		}
	
	
}

func TestExtract(t *testing.T) {
	// Test cases
	testCases := []struct {
		name         string
		html         string
		expectedWords []string
		expectedHrefs []string
	}{
		{
			name: "Simple HTML",
			html: `
				<html>
					<body>
						<p>This is some text.</p>
						<a href="https://example.com">Link</a>
					</body>
				</html>
			`,
			expectedWords: []string{"This", "is", "some", "text.", "Link"},
			expectedHrefs: []string{"https://example.com"},
		},
		{
			name: "No Text or Links",
			html: `
				<html>
					<body>
						<div></div>
					</body>
				</html>
			`,
			expectedWords: []string{},
			expectedHrefs: []string{},
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			words, hrefs := extract([]byte(tc.html))

			// Verify extracted words
			if !stringSlicesEqual(words, tc.expectedWords) {
				t.Errorf("Extracted words mismatch. Expected: %v, Got: %v", tc.expectedWords, words)
			}

			// Verify extracted hrefs
			if !stringSlicesEqual(hrefs, tc.expectedHrefs) {
				t.Errorf("Extracted hrefs mismatch. Expected: %v, Got: %v", tc.expectedHrefs, hrefs)
			}
		})
	}
}

func stringSlicesEqual(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}
	return true
}

func TestSplit(t *testing.T) {
	bookURL := "https://cs272-0304-f23.github.io/tests/top10/" //	// URL of the book content 

	resp, err := http.Get(bookURL) 	// Fetch the book content from the URL
	if err != nil {
		t.Fatalf("Failed to fetch book content: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to fetch book content. Status code: %d", resp.StatusCode)
	}

	bookBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read book content: %v", err)
	}

	tempDir, err := ioutil.TempDir("", "book") 	// Create a temporary file to store the book content
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tempFilePath := filepath.Join(tempDir, "book.txt")

	file, err := os.Create(tempFilePath)
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer file.Close()

	_, err = file.Write(bookBytes) // Write the book content to the temporary file
	if err != nil {
		t.Fatalf("Failed to write book content to temporary file: %v", err)
	}

	// Get the file path of the temporary file
	filePath := tempFile.Name()

	// Expected chapter titles and web links to hash values
	expectedChapters := map[string]string{
		"https://cs272-0304-f23.github.io/tests/lab04/hashes.txt",
	}

	// Split the book
	chapterMap, err := Split(string(bookBytes))
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	// Test each chapter's digest
	for title, content := range chapterMap {
		expectedHashValue, err := fetchHashValue(expectedChapters[title]) 		// Fetch the expected hash value from the web link
		if err != nil {
			t.Fatalf("Failed to fetch expected hash value: %v", err)
		}

		digest := hashFile(filePath) 		// Calculate the digest using the hash function on the temporary file

		if digest != expectedHashValue {
			t.Errorf("Digest mismatch for chapter %s. Expected: %d, Got: %d", title, expectedHashValue, digest) // Verify the digest matches the expected value
		}
	}
}


func generateExpectedDigest(title, content string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(title))
	_, _ = h.Write([]byte(content))
	return h.Sum32()
}

func hashFile(path string) uint32 {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalln(err)
	}
	bts, err := io.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}
	h := fnv.New32a()
	h.Write(bts)
	return h.Sum32()
}

func fetchHashValue(hashURL string) (uint32, error) {
	resp, err := http.Get(hashURL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Failed to fetch hash value. Status code: %d", resp.StatusCode)
	}

	hashBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	hashValue, err := strconv.ParseUint(strings.TrimSpace(string(hashBytes)), 10, 32) // Convert the hash value from string to uint32 (assuming it's a valid uint32 string)
	if err != nil {
		return 0, err
	}

	return uint32(hashValue), nil
}

func TestMainFunction(t *testing.T) {
	// Create a test server to mock HTTP responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock the response for the sourceURL
		if r.URL.String() == "https://cs272-0304-f23.github.io/tests/lab03/" {
			fmt.Fprintln(w, `<html><a href="https://example.com/page1">Page 1</a></html>`)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	oldStdout := os.Stdout 	// Redirect standard output to capture the printed output
	r, w, _ := os.Pipe()
	os.Stdout = w

	sourceURL := server.URL 	// Call the main function with the test server URL
	main()

	w.Close() 	// Close the pipe, restore standard output, and read captured output
	os.Stdout = oldStdout
	capturedOutput, _ := ioutil.ReadAll(r)

	// Verify the captured output
	expectedOutput := "Downloaded URLs:\n" +
		"https://cs272-0304-f23.github.io/tests/lab03/\n" +
		"https://example.com/page1\n"

	if !strings.Contains(string(capturedOutput), expectedOutput) {
		t.Errorf("Unexpected output. Expected: %s, Got: %s", expectedOutput, string(capturedOutput))
	}
}

func TestAllFunctions(t *testing.T) {
	// Run all the individual function tests here.
	TestGenerateDigest(t)
	TestSplitBook(t)
	TestClean(t)
	TestDownload(t)
	TestExtract(t)
	TestMainFunction(t)
}

func main() {
	// Run all the tests
	TestAllFunctions()
	fmt.Println("All tests passed.")
}