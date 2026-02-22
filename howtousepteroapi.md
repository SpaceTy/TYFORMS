# Using Pterodactyl API in Go

Quick guide to sending commands to your Minecraft server via the Pterodactyl API using Go.

## Prerequisites

- Client API key from Pterodactyl panel (Account → API Credentials)
- Server ID (found in your panel URL: `/server/{SERVER_ID}`)

## Installation

No external dependencies needed—uses Go's standard library:

```bash
go mod init pterodactyl-cmd
```

## Basic Usage

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

// PterodactylClient handles API communication
type PterodactylClient struct {
    PanelURL string
    APIKey   string
    Client   *http.Client
}

// NewClient creates a new API client
func NewClient(panelURL, apiKey string) *PterodactylClient {
    return &PterodactylClient{
        PanelURL: panelURL,
        APIKey:   apiKey,
        Client:   &http.Client{},
    }
}

// SendCommand sends a command to a server
func (c *PterodactylClient) SendCommand(serverID, command string) error {
    url := fmt.Sprintf("%s/api/client/servers/%s/command", c.PanelURL, serverID)
    
    body, _ := json.Marshal(map[string]string{
        "command": command,
    })
    
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
    if err != nil {
        return fmt.Errorf("creating request: %w", err)
    }
    
    req.Header.Set("Authorization", "Bearer "+c.APIKey)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "Application/vnd.pterodactyl.v1+json")
    
    resp, err := c.Client.Do(req)
    if err != nil {
        return fmt.Errorf("sending request: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusNoContent {
        return fmt.Errorf("API error: status %d", resp.StatusCode)
    }
    
    return nil
}
```

## Complete Example

```go
package main

import (
    "log"
    "os"
)

func main() {
    // Load from environment variables
    panelURL := os.Getenv("PTERODACTYL_URL") // e.g., "https://your-panel.com"
    apiKey := os.Getenv("PTERODACTYL_API_KEY") // e.g., "ptlc_xxxxxxxxxxxxxxxx"
    serverID := os.Getenv("MINECRAFT_SERVER_ID") // e.g., "d3aac109"

    if panelURL == "" || apiKey == "" || serverID == "" {
        log.Fatal("Missing required environment variables")
    }

    client := NewClient(panelURL, apiKey)

    // Send a command
    err := client.SendCommand(serverID, "say Hello from Go!")
    if err != nil {
        log.Fatalf("Failed to send command: %v", err)
    }

    log.Println("Command sent successfully!")
}
```

## Running

```bash
export PTERODACTYL_URL="https://your-panel.com"
export PTERODACTYL_API_KEY="ptlc_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
export MINECRAFT_SERVER_ID="d3aac109"

go run main.go
```

## Security Notes

- **Never hardcode API keys**—use environment variables or a secrets manager
- Restrict API key permissions to specific servers in the panel
- Consider IP whitelisting for the API key
- Rate limit commands in your application to prevent abuse

## Error Handling

The API returns:
- `204 No Content` on success
- `403 Forbidden` if unauthorized
- `400 Bad Request` if server is offline or command invalid
- Rate limit headers: `X-RateLimit-Remaining`, `X-RateLimit-Reset`

## WebSocket Console (Advanced)

For real-time console, use Gorilla WebSocket:

```go
// Get websocket token first
type WSResponse struct {
    Data struct {
        Token string `json:"token"`
        Socket string `json:"socket"`
    } `json:"data"`
}

// Then connect with gorilla/websocket
```
