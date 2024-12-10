package githelper

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
)

// Extract a token using regex (CSRF-related or others)
func extractToken(body, pattern string) (string, error) {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(body)
	if len(matches) < 2 {
		return "", fmt.Errorf("failed to extract token with pattern: %s", pattern)
	}
	return matches[1], nil
}

// Perform an HTTP GET request and return the response body as a string (CSRF-related)
func getRequestCSRF(client *http.Client, url string) (string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed GET request to %s: %v", url, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}
	return string(body), nil
}

// Perform an HTTP POST request with form data and return the response body as a string (CSRF-related)
func postRequestCSRF(client *http.Client, url string, form url.Values, headers map[string]string) (string, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create POST request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed POST request to %s: %v", url, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}
	return string(body), nil
}

func CreatePersonalAccessTokenCSRF(gitlabUser, gitlabPassword, tokenName string) (string, error) {
	// Replace with your values
	gitlabHost := "https://gitlab.signal18.io"

	// Create a cookie jar to handle cookies
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	// 1. Get the login page to retrieve CSRF token
	loginPageURL := fmt.Sprintf("%s/users/sign_in", gitlabHost)
	body, err := getRequestCSRF(client, loginPageURL)
	if err != nil {
		return "", err
	}

	csrfToken, err := extractToken(body, `name="authenticity_token" value="([^"]+)"`)
	if err != nil {
		return "", err
	}

	// 2. Send login credentials
	form := url.Values{
		"user[login]":        {gitlabUser},
		"user[password]":     {gitlabPassword},
		"authenticity_token": {csrfToken},
	}
	_, err = postRequestCSRF(client, loginPageURL, form, nil)
	if err != nil {
		return "", err
	}

	// 3. Access personal access tokens page to retrieve new CSRF token
	tokensPageURL := fmt.Sprintf("%s/-/user_settings/personal_access_tokens", gitlabHost)
	body, err = getRequestCSRF(client, tokensPageURL)
	if err != nil {
		return "", err
	}

	csrfToken, err = extractToken(body, `name="csrf-token" content="([^"]+)"`)
	if err != nil {
		return "", err
	}
	fmt.Println("New CSRF Token:", csrfToken)

	// 4. Generate a personal access token
	form = url.Values{
		"authenticity_token":              {csrfToken},
		"personal_access_token[name]":     {tokenName},
		"personal_access_token[scopes][]": {"api", "write_repository"},
	}
	headers := map[string]string{
		"X-CSRF-Token": csrfToken,
	}
	body, err = postRequestCSRF(client, tokensPageURL, form, headers)
	if err != nil {
		return "", err
	}

	fmt.Println("New PAT Token:", body)

	// 5. Scrape the personal access token from the response
	return extractToken(body, `"new_token":"([^"]+)"`)
}
