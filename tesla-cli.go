package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

// TeslaTokens represents the OAuth token data
type TeslaTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

// loadEnvFile loads credentials from .env file
func loadEnvFile() (string, string, string, string) {
	clientID := ""
	clientSecret := ""
	redirectURI := ""
	vehicleID := ""

	file, err := os.Open(".env")
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "#") || len(strings.TrimSpace(line)) == 0 {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			// Remove quotes if present
			val = strings.Trim(val, "\"'")
			// For VIN, keep only alphanumeric (removes any hidden chars)
			if key == "TESLA_VEHICLE_VIN" {
				cleanVIN := ""
				for _, r := range val {
					if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
						cleanVIN += string(r)
					}
				}
				val = cleanVIN
			}
			switch key {
			case "TESLA_CLIENT_ID":
				clientID = val
			case "TESLA_CLIENT_SECRET":
				clientSecret = val
			case "TESLA_REDIRECT_URI":
				redirectURI = val
			case "TESLA_VEHICLE_VIN":
				vehicleID = val
			}
		}
	}
	return clientID, clientSecret, redirectURI, vehicleID
}

// loadTokens loads tokens from tesla-tokens.json
func loadTokens() (TeslaTokens, error) {
	tokens := TeslaTokens{}
	file, err := os.Open("tesla-tokens.json")
	if err != nil {
		return tokens, err
	}
	defer file.Close()
	err = json.NewDecoder(file).Decode(&tokens)
	return tokens, err
}

// authenticateAndGetTokens performs OAuth authentication
func authenticateAndGetTokens() (TeslaTokens, error) {
	fmt.Println("\n🔐 OAuth Authentication")
	fmt.Println("======================\n")

	clientID, clientSecret, redirectURI, _ := loadEnvFile()

	reader := bufio.NewReader(os.Stdin)

	if clientID == "" {
		fmt.Print("Enter Tesla CLIENT_ID: ")
		clientID, _ = reader.ReadString('\n')
		clientID = strings.TrimSpace(clientID)
	}
	if clientSecret == "" {
		fmt.Print("Enter Tesla CLIENT_SECRET: ")
		clientSecret, _ = reader.ReadString('\n')
		clientSecret = strings.TrimSpace(clientSecret)
	}
	if redirectURI == "" {
		fmt.Print("Enter Tesla REDIRECT_URI (default: http://localhost:8080/callback): ")
		inputURI, _ := reader.ReadString('\n')
		inputURI = strings.TrimSpace(inputURI)
		if inputURI != "" {
			redirectURI = inputURI
		} else {
			redirectURI = "http://localhost:8080/callback"
		}
	}

	scopes := "openid offline_access vehicle_device_data vehicle_cmds vehicle_charging_cmds"

	// Build OAuth URL
	params := url.Values{}
	params.Add("client_id", clientID)
	params.Add("redirect_uri", redirectURI)
	params.Add("response_type", "code")
	params.Add("scope", scopes)
	params.Add("state", fmt.Sprintf("%d", time.Now().Unix()))

	authURL := "https://auth.tesla.com/oauth2/v3/authorize?" + params.Encode()

	fmt.Println("\n📱 Opening browser for Tesla login...")
	fmt.Println()
	fmt.Println("If browser doesn't open, copy and paste this URL:")
	fmt.Println(authURL)
	fmt.Println()

	// Try to open browser - Windows: use rundll32 for URLs
	exec.Command("rundll32", "url.dll,FileProtocolHandler", authURL).Run()

	time.Sleep(2 * time.Second)

	fmt.Print("After logging in, paste the authorization CODE from the redirect URL: ")
	code, _ := reader.ReadString('\n')
	code = strings.TrimSpace(code)

	if code == "" {
		return TeslaTokens{}, fmt.Errorf("no authorization code provided")
	}

	fmt.Println("\n📡 Exchanging code for tokens...")

	// Exchange code for tokens
	resp, err := http.PostForm("https://auth.tesla.com/oauth2/v3/token",
		url.Values{
			"grant_type":    {"authorization_code"},
			"client_id":     {clientID},
			"client_secret": {clientSecret},
			"code":          {code},
			"redirect_uri":  {redirectURI},
		})
	if err != nil {
		return TeslaTokens{}, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	var tokensResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tokensResp); err != nil {
		return TeslaTokens{}, fmt.Errorf("failed to decode token response: %w", err)
	}

	if errMsg, ok := tokensResp["error"]; ok {
		return TeslaTokens{}, fmt.Errorf("token error: %v", errMsg)
	}

	accessToken, _ := tokensResp["access_token"].(string)
	refreshToken, _ := tokensResp["refresh_token"].(string)
	expiresIn, _ := tokensResp["expires_in"].(float64)
	expiresAt := time.Now().Unix() + int64(expiresIn)

	tokens := TeslaTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}

	// Save tokens
	file, err := os.Create("tesla-tokens.json")
	if err == nil {
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(tokens)
		file.Close()
		fmt.Println("✅ Tokens saved to tesla-tokens.json!")
	}

	return tokens, nil
}

// sendCommand sends a command to the Tesla proxy
func sendCommand(vin string, token string, command string, proxyURL string, params map[string]interface{}) error {
	endpoint := fmt.Sprintf("%s/api/1/vehicles/%s/command/%s", proxyURL, vin, command)

	// Create HTTP client with insecure TLS
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// Convert params to JSON
	var bodyData []byte
	if len(params) > 0 {
		bodyData, _ = json.Marshal(params)
	} else {
		bodyData = []byte("{}")
	}

	// Create request
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(string(bodyData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check if successful
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	} else {
		return fmt.Errorf("command failed with status %d: %s", resp.StatusCode, string(body))
	}
}

// wakeVehicle sends a request to wake the vehicle
func wakeVehicle(vin string, token string, proxyURL string) error {
	endpoint := fmt.Sprintf("%s/api/1/vehicles/%s", proxyURL, vin)

	// Create HTTP client with insecure TLS
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// Try waking the vehicle up to 3 times
	for attempt := 1; attempt <= 3; attempt++ {
		fmt.Printf("  📡 Wake attempt %d/3...\n", attempt)

		// Send GET request to wake the vehicle
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			continue
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
		}

		if attempt < 3 {
			time.Sleep(1 * time.Second)
		}
	}

	// Vehicle is now waking (or was already awake)
	return nil
}

// sendCommandWithRetry sends a command and retries if vehicle is offline/asleep
func sendCommandWithRetry(vin string, token string, command string, proxyURL string, params map[string]interface{}) error {
	err := sendCommand(vin, token, command, proxyURL, params)

	// If we got an error about vehicle being offline/asleep, try waking it and retry multiple times
	if err != nil && strings.Contains(err.Error(), "vehicle unavailable") {
		fmt.Println("  ⏳ Vehicle is sleeping... waking it up...")
		fmt.Println("  💡 Tip: For deeply sleeping vehicles, wake from Tesla app first for faster response")
		wakeVehicle(vin, token, proxyURL)

		// Try the command up to 3 times with increasing wait times
		for attempt := 1; attempt <= 3; attempt++ {
			waitTime := 5 + (attempt * 3) // 8, 11, 14 seconds
			fmt.Printf("  ⏸️  Waiting %d seconds before retry %d/3...\n", waitTime, attempt)
			for i := waitTime; i > 0; i-- {
				fmt.Printf("  %d.", i)
				time.Sleep(1 * time.Second)
			}
			fmt.Println()

			err = sendCommand(vin, token, command, proxyURL, params)
			if err == nil || !strings.Contains(err.Error(), "vehicle unavailable") {
				break
			}

			if attempt < 3 {
				fmt.Println("  🔄 Still offline, trying again...")
				wakeVehicle(vin, token, proxyURL)
			}
		}
	}

	return err
}

func main() {
	fmt.Println("🚗 Tesla Vehicle Command CLI")
	fmt.Println("============================\n")

	// Load VIN from .env
	_, _, _, vehicleID := loadEnvFile()
	if vehicleID == "" {
		fmt.Println("❌ TESLA_VEHICLE_VIN not found in .env")
		fmt.Println("\nPlease add your vehicle VIN to .env:")
		fmt.Println("  TESLA_VEHICLE_VIN=your_vin_here")
		fmt.Println("\nYou can find your VIN in your Tesla vehicle:")
		fmt.Println("  Settings > Software")
		os.Exit(1)
	}

	fmt.Printf("✅ Vehicle VIN: %s (length: %d)\n\n", vehicleID, len(vehicleID))

	// Get tokens
	tokens, err := loadTokens()
	if err != nil || tokens.AccessToken == "" {
		fmt.Println("📂 No valid tokens found, performing authentication...")
		tokens, err = authenticateAndGetTokens()
		if err != nil {
			fmt.Printf("❌ Authentication failed: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("✅ Connected to Tesla\n\n")

	// Main menu loop
	reader := bufio.NewReader(os.Stdin)
	proxyURL := "https://localhost:4443"

	for {
		fmt.Println("What would you like to do?")
		fmt.Println("1. Wake Vehicle")
		fmt.Println("2. Lock Doors")
		fmt.Println("3. Unlock Doors")
		fmt.Println("4. Sentry Mode ON")
		fmt.Println("5. Sentry Mode OFF")
		fmt.Println("6. Quit")
		fmt.Print("\nEnter choice [1-6]: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		fmt.Println()

		switch choice {
		case "1":
			fmt.Println("⚡ Waking vehicle...")
			if err := wakeVehicle(vehicleID, tokens.AccessToken, proxyURL); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			} else {
				fmt.Println("✅ Vehicle waking up (may take a few seconds)...")
			}

		case "2":
			fmt.Println("🔒 Locking doors...")
			if err := sendCommandWithRetry(vehicleID, tokens.AccessToken, "door_lock", proxyURL, map[string]interface{}{}); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			} else {
				fmt.Println("✅ Doors locked successfully!")
			}

		case "3":
			fmt.Println("🔓 Unlocking doors...")
			if err := sendCommandWithRetry(vehicleID, tokens.AccessToken, "door_unlock", proxyURL, map[string]interface{}{}); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			} else {
				fmt.Println("✅ Doors unlocked successfully!")
			}

		case "4":
			fmt.Println("🛡️  Enabling Sentry Mode...")
			if err := sendCommandWithRetry(vehicleID, tokens.AccessToken, "set_sentry_mode", proxyURL, map[string]interface{}{"on": true}); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			} else {
				fmt.Println("✅ Sentry Mode enabled!")
			}

		case "5":
			fmt.Println("🛡️  Disabling Sentry Mode...")
			if err := sendCommandWithRetry(vehicleID, tokens.AccessToken, "set_sentry_mode", proxyURL, map[string]interface{}{"on": false}); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			} else {
				fmt.Println("✅ Sentry Mode disabled!")
			}

		case "6":
			fmt.Println("Goodbye! 👋")
			return

		default:
			fmt.Println("❌ Invalid choice. Please enter 1-6.")
		}

		fmt.Println()
	}
}
