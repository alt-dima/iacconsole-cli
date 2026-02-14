package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/alt-dima/iacconsole-cli/utils"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

var (
	autoExecute bool
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Run iacconsole-cli in agent mode",
	Long:  `Connects to the IaCConsole server via WebSocket to receive and execute infrastructure commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		apiUrl := os.Getenv("IACCONSOLE_API_URL")
		wsURL, authHeader, accountID, err := utils.ParseAPIURL(apiUrl)
		if err != nil {
			log.Fatalf("Configuration error: %v", err)
		}

		// Generate unique agent ID (hostname + random suffix)
		agentID := generateAgentID()

		log.Printf("Starting agent for account: %s", accountID)
		log.Printf("Agent ID: %s", agentID)
		log.Printf("Connecting to: %s", wsURL)

		runAgent(wsURL, authHeader, agentID)
	},
}

func runAgent(wsURL, authHeader, agentID string) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	for {
		header := http.Header{}
		header.Add("Authorization", authHeader)

		c, _, err := websocket.DefaultDialer.Dial(wsURL, header)
		if err != nil {
			log.Printf("Dial error: %v. Retrying in 5s...", err)
			select {
			case <-time.After(5 * time.Second):
				continue
			case <-interrupt:
				return
			}
		}

		log.Printf("Connected to IaCConsole server")

		// Register agent with unique agent ID
		reg := utils.AgentRegister{
			AgentMessage: utils.AgentMessage{Type: "register"},
			AgentID:      agentID,
			Version:      rootCmd.Version,
			OS:           runtime.GOOS,
			Arch:         runtime.GOARCH,
		}
		if err := c.WriteJSON(reg); err != nil {
			log.Printf("Register error: %v", err)
			c.Close()
			continue
		}

		done := make(chan struct{})

		// Configure read deadline and pong handler for WebSocket control frames from API
		if err := c.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			log.Printf("Failed to set read deadline: %v", err)
			c.Close()
			continue
		}
		c.SetPongHandler(func(string) error {
			log.Printf("Pong received from server, resetting read deadline")
			return c.SetReadDeadline(time.Now().Add(60 * time.Second))
		})
		c.SetPingHandler(func(appData string) error {
			log.Printf("Ping received from server, sending pong and resetting read deadline")
			if err := c.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(10*time.Second)); err != nil {
				return err
			}
			return c.SetReadDeadline(time.Now().Add(60 * time.Second))
		})

		// Read loop
		go func() {
			defer close(done)
			for {
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Printf("Read error: %v", err)
					return
				}

				var baseMsg utils.AgentMessage
				if err := json.Unmarshal(message, &baseMsg); err != nil {
					log.Println(string(message))
					log.Printf("Unmarshal error: %v", err)
					continue
				}

				switch baseMsg.Type {
				case "error":
					var errMsg utils.AgentError
					if err := json.Unmarshal(message, &errMsg); err != nil {
						log.Printf("Error message unmarshal error: %v", err)
						continue
					}
					log.Printf("Server error: %s", errMsg.Error)
					// Exit if it's an agent already connected error
					if containsString(errMsg.Error, "already connected") {
						log.Fatalf("Cannot start agent: %s", errMsg.Error)
					}
					return
				case "command":
					var cmd utils.AgentCommand
					if err := json.Unmarshal(message, &cmd); err != nil {
						log.Printf("Command unmarshal error: %v", err)
						continue
					}
					// log.Printf("Received command: %+v", cmd)
					go executeCommand(c, cmd, autoExecute)
				case "ping":
					pong := utils.AgentPong{
						AgentMessage: utils.AgentMessage{Type: "pong"},
						Timestamp:    time.Now().Unix(),
					}
					pongBytes, _ := json.Marshal(pong)
					if err := c.WriteMessage(websocket.TextMessage, pongBytes); err != nil {
						log.Printf("Failed to send pong: %v", err)
					}
				}
			}
		}()

		select {
		case <-done:
			c.Close()
			log.Printf("Connection lost. Retrying in 5s...")
			time.Sleep(5 * time.Second)
		case <-interrupt:
			log.Println("Interrupt received, closing connection...")
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Write close error:", err)
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			c.Close()
			return
		}
	}
}

func executeCommand(c *websocket.Conn, cmd utils.AgentCommand, autoExecute bool) {
	// Format command for display
	cmdStr := formatCommandString(cmd)

	// Check for approval if auto-execute is disabled
	if !autoExecute {
		log.Printf("\n=== Command Approval Required ===")
		log.Printf("Command to execute: %s", cmdStr)
		fmt.Print("Approve execution? (yes/no): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading input: %v", err)

			// Send error message back to browser
			completeMsg := utils.AgentComplete{
				AgentMessage: utils.AgentMessage{Type: "complete"},
				CommandID:    cmd.ID,
				ExitCode:     1,
				Error:        fmt.Sprintf("Failed to read user input: %v", err),
			}
			if writeErr := c.WriteJSON(completeMsg); writeErr != nil {
				log.Printf("Failed to send error message: %v", writeErr)
			}
			return
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" && response != "y" {
			log.Printf("Command execution rejected by user")

			// Send rejection message back to browser
			completeMsg := utils.AgentComplete{
				AgentMessage: utils.AgentMessage{Type: "complete"},
				CommandID:    cmd.ID,
				ExitCode:     1,
				Error:        "Command execution rejected by user",
			}
			if err := c.WriteJSON(completeMsg); err != nil {
				log.Printf("Failed to send rejection message: %v", err)
			}
			return
		}

		log.Printf("Command approved, executing...")
	} else {
		log.Printf("Auto-executing command: %s", cmdStr)
	}

	state := &utils.State{}
	state.IacconsoleApiUrl = os.Getenv("IACCONSOLE_API_URL")
	state.StateS3Path = "./state"

	utils.ExecuteAgentCommand(c, cmd, state)
}

// formatCommandString creates a human-readable string representation of the command
func formatCommandString(cmd utils.AgentCommand) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("iacconsole-cli %s", cmd.Action))

	if cmd.Org != "" {
		parts = append(parts, fmt.Sprintf("--org=%s", cmd.Org))
	}

	if cmd.Unit != "" {
		parts = append(parts, fmt.Sprintf("--unit=%s", cmd.Unit))
	}

	for _, dim := range cmd.Dimensions {
		parts = append(parts, fmt.Sprintf("--dimension=%s:%s", dim.Key, dim.Value))
	}

	if cmd.Workspace != "" {
		parts = append(parts, fmt.Sprintf("--workspace=%s", cmd.Workspace))
	}

	if len(cmd.ExtraArgs) > 0 {
		parts = append(parts, strings.Join(cmd.ExtraArgs, " "))
	}

	return strings.Join(parts, " ")
}

// generateAgentID creates a unique agent identifier using hostname and random suffix
func generateAgentID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	// Generate a short random suffix (5 chars)
	rand.Seed(time.Now().UnixNano())
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	suffix := make([]byte, 5)
	for i := range suffix {
		suffix[i] = charset[rand.Intn(len(charset))]
	}
	return fmt.Sprintf("%s-%s", hostname, string(suffix))
}

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && contains(s, substr))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(agentCmd)
	agentCmd.Flags().BoolVar(&autoExecute, "auto-execute", false, "Automatically approve and execute commands without prompting")
}
