package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func shortenURL(endpoint, longURL string, client *http.Client) (string, error) {
	data := url.Values{}
	data.Set("url", longURL)

	request, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}

func main() {
	endpoint := "http://localhost:8080/"
	fmt.Println("Введите длинный URL:")
	reader := bufio.NewReader(os.Stdin)
	long, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	long = strings.TrimSpace(long)

	client := &http.Client{}
	result, err := shortenURL(endpoint, long, client)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Результат:", result)
}
