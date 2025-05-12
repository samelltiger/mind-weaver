package util

import (
	"bufio"
	"context"
	"errors"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

// CommandOutput represents the output from a command execution
type CommandOutput struct {
	Line      string `json:"line"`
	IsError   bool   `json:"isError"`
	Timestamp int64  `json:"timestamp"`
}

// CommandExecutionResult contains the final result of a command execution
type CommandExecutionResult struct {
	Success      bool   `json:"success"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	ExitCode     int    `json:"exitCode"`
}

// CommandService handles execution of shell commands and code
type CommandService struct {
	// Map of dangerous commands/patterns that should be blocked
	dangerousCommands map[string]*regexp.Regexp
	// Map of active command executions
	activeCommands sync.Map
}

// NewCommandService creates a new command execution service
func NewCommandService() *CommandService {
	// Initialize with dangerous commands that should be blocked
	dangerousPatterns := map[string]*regexp.Regexp{
		"rm -rf /":      regexp.MustCompile(`rm\s+-rf\s+\/`),
		"format disk":   regexp.MustCompile(`mkfs|fdisk\s+\/dev`),
		"delete users":  regexp.MustCompile(`userdel\s+(-r\s+)?root`),
		"chmod 777 /":   regexp.MustCompile(`chmod\s+777\s+\/`),
		"wget malware":  regexp.MustCompile(`wget.+\|\s*bash`),
		"curl malware":  regexp.MustCompile(`curl.+\|\s*bash`),
		"fork bomb":     regexp.MustCompile(`:\(\)\{\s*:\|\:&\s*\};\s*:`),
		"write to dev":  regexp.MustCompile(`>\s*\/dev\/[sh]d[a-z]`),
		"shutdown":      regexp.MustCompile(`shutdown|reboot|halt|poweroff`),
		"network flood": regexp.MustCompile(`ping\s+-f`),
	}

	return &CommandService{
		dangerousCommands: dangerousPatterns,
	}
}

// ExecuteCommand runs a shell command and streams the output
func (s *CommandService) ExecuteCommand(ctx context.Context, command string, outputChan chan<- CommandOutput) (*CommandExecutionResult, error) {
	// Validate the command
	safe, reason := s.ValidateCommand(command)
	if !safe {
		return &CommandExecutionResult{
			Success:      false,
			ErrorMessage: reason,
			ExitCode:     -1,
		}, errors.New(reason)
	}

	// Create command
	// cmd := exec.CommandContext(ctx, "bash", "-c", command)
	cmd := exec.Command("bash", "-c", command)

	// Get stdout and stderr pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// Store command in active commands map with a unique ID
	cmdID := time.Now().UnixNano()
	s.activeCommands.Store(cmdID, cmd)

	// Clean up when done
	defer func() {
		s.activeCommands.Delete(cmdID)
		close(outputChan) // Ensure channel is closed when function exits
	}()

	// Process stdout and stderr
	var wg sync.WaitGroup
	wg.Add(2)

	// Process stdout
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			case outputChan <- CommandOutput{
				Line:      scanner.Text(),
				IsError:   false,
				Timestamp: time.Now().UnixMilli(),
			}:
			}
		}
	}()

	// Process stderr
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			case outputChan <- CommandOutput{
				Line:      scanner.Text(),
				IsError:   true,
				Timestamp: time.Now().UnixMilli(),
			}:
			}
		}
	}()

	// Wait for both goroutines to finish
	wg.Wait()

	// Wait for command to finish
	err = cmd.Wait()

	// Prepare result
	result := &CommandExecutionResult{
		Success:  err == nil,
		ExitCode: cmd.ProcessState.ExitCode(),
	}

	if err != nil {
		result.ErrorMessage = err.Error()
	}

	return result, nil
}

// ValidateCommand checks if a command is safe to execute
func (s *CommandService) ValidateCommand(cmd string) (bool, string) {
	// Check against dangerous patterns
	for name, pattern := range s.dangerousCommands {
		if pattern.MatchString(cmd) {
			return false, "Command contains dangerous pattern: " + name
		}
	}

	// Additional validation logic can be added here
	// For example, checking for specific commands that should be allowed

	return true, ""
}

// StopCommand terminates a running command by its ID
func (s *CommandService) StopCommand(cmdID int64) error {
	cmdValue, exists := s.activeCommands.Load(cmdID)
	if !exists {
		return errors.New("command not found")
	}

	cmd, ok := cmdValue.(*exec.Cmd)
	if !ok {
		return errors.New("invalid command type")
	}

	// Kill the process
	if cmd.Process != nil {
		return cmd.Process.Kill()
	}

	return nil
}

// ExecuteCode runs code in a specific language
func (s *CommandService) ExecuteCode(ctx context.Context, code string, language string, outputChan chan<- CommandOutput) (*CommandExecutionResult, error) {
	var cmd string

	// Configure execution based on language
	switch language {
	case "javascript", "js":
		// Create a temporary file to execute
		cmd = `echo '` + strings.ReplaceAll(code, "'", "'\\''") + `' > /tmp/codepilot_exec.js && node /tmp/codepilot_exec.js && rm /tmp/codepilot_exec.js`

	case "python", "py":
		cmd = `echo '` + strings.ReplaceAll(code, "'", "'\\''") + `' > /tmp/codepilot_exec.py && python3 /tmp/codepilot_exec.py && rm /tmp/codepilot_exec.py`

	case "go":
		cmd = `echo '` + strings.ReplaceAll(code, "'", "'\\''") + `' > /tmp/codepilot_exec.go && go run /tmp/codepilot_exec.go && rm /tmp/codepilot_exec.go`

	case "bash", "sh":
		cmd = `echo '` + strings.ReplaceAll(code, "'", "'\\''") + `' > /tmp/codepilot_exec.sh && bash /tmp/codepilot_exec.sh && rm /tmp/codepilot_exec.sh`

	default:
		return nil, errors.New("unsupported language: " + language)
	}

	// Execute the command
	return s.ExecuteCommand(ctx, cmd, outputChan)
}
