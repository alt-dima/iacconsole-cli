package utils

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ExecuteAgentCommand runs a command and streams output to WebSocket
func ExecuteAgentCommand(conn *websocket.Conn, cmd AgentCommand, state *State) {
	// 1. Prepare environment
	state.OrgName = cmd.Org
	state.UnitName = cmd.Unit
	state.Workspace = cmd.Workspace
	if state.Workspace == "" {
		state.Workspace = "master"
	}

	state.ParsedDimensions = make(map[string]string)
	for _, dp := range cmd.Dimensions {
		state.ParsedDimensions[dp.Key] = dp.Value
	}

	// 2. Setup paths (reuse existing logic from exec.go)
	// Note: We assume the agent has access to local units/inventory if FileSync wasn't used
	// If FileSync was used, those files should already be in a temp dir.

	// For now, let's assume we use the temp dir logic
	state.PrepareTemp()

	// Generate variables
	state.GenerateVarsByDims()
	state.GenerateVarsByDimOptional("defaults")
	state.GenerateVarsByEnvVars()

	backendConfig := state.SetupBackendConfig()
	state.GenerateVarsByDimAndData("config", "backend", backendConfig)

	// 3. Prepare execution
	cmdToExec := state.GetStringFromViperByOrgOrDefault("cmd_to_exec")
	if cmdToExec == "" {
		cmdToExec = "tofu"
	}

	args := []string{cmd.Action}
	if cmd.Action == "init" {
		for param, value := range backendConfig {
			args = append(args, "-backend-config="+param+"="+value.(string))
		}
	}
	args = append(args, cmd.ExtraArgs...)

	// 4. Spawn process
	log.Printf("Agent executing: %s %s", cmdToExec, strings.Join(args, " "))
	child := exec.Command(cmdToExec, args...)
	child.Dir = state.CmdWorkTempDir
	child.Env = os.Environ()

	stdout, _ := child.StdoutPipe()
	stderr, _ := child.StderrPipe()

	err := child.Start()
	if err != nil {
		sendComplete(conn, cmd.ID, 1, err.Error())
		return
	}

	// 5. Stream output
	var mu sync.Mutex
	done := make(chan bool)
	go streamPipe(conn, &mu, cmd.ID, "stdout", stdout, done)
	go streamPipe(conn, &mu, cmd.ID, "stderr", stderr, done)

	err = child.Wait()
	<-done
	<-done

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	// 6. Cleanup
	if exitCode == 0 && (cmd.Action == "apply" || cmd.Action == "destroy") {
		os.RemoveAll(state.CmdWorkTempDir)
	}

	mu.Lock()
	sendComplete(conn, cmd.ID, exitCode, "")
	mu.Unlock()
}

func streamPipe(conn *websocket.Conn, mu *sync.Mutex, cmdID string, stream string, pipe io.ReadCloser, done chan bool) {
	defer pipe.Close()
	defer func() { done <- true }()

	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()
		msg := AgentOutput{
			AgentMessage: AgentMessage{Type: "output"},
			CommandID:    cmdID,
			Stream:       stream,
			Data:         line + "\n",
			Timestamp:    time.Now().Unix(),
		}
		mu.Lock()
		conn.WriteJSON(msg)
		mu.Unlock()
	}
}

func sendComplete(conn *websocket.Conn, cmdID string, exitCode int, errMsg string) {
	complete := AgentComplete{
		AgentMessage: AgentMessage{Type: "complete"},
		CommandID:    cmdID,
		ExitCode:     exitCode,
		Error:        errMsg,
	}
	conn.WriteJSON(complete)
}
