// internal/services/command_service.go
package services

import (
	"context"
	"fmt"
	"mind-weaver/config"
	"mind-weaver/pkg/logger"
	"mind-weaver/pkg/util"
	"strings"
)

// CommandService handles execution of shell commands and code
type CommandService struct {
	commandService *util.CommandService
}

// NewCommandService creates a new command execution service
func NewCommandService() *CommandService {

	return &CommandService{
		commandService: util.NewCommandService(),
	}
}

// ExecuteCommand runs a shell command and streams the output
func (s *CommandService) ExecuteCommand(ctx context.Context, command string, outputChan chan<- util.CommandOutput) (*util.CommandExecutionResult, error) {
	return s.commandService.ExecuteCommand(ctx, command, outputChan)
}

// 执行html+JavaScript代码文件的检查逻辑
func (s *CommandService) JsInspector(htmlPath string) (*util.CommandExecutionResult, []util.CommandOutput, error) {
	// Collect command output
	var outputs []util.CommandOutput
	outputChan := make(chan util.CommandOutput, 200)

	// Start a goroutine to collect outputs
	done := make(chan struct{})
	go func() {
		for output := range outputChan {
			outputs = append(outputs, output)
		}
		close(done)
	}()

	commandStr := fmt.Sprintf("%s ./scripts/run_jsinspector.py %s", config.GetConfig().Bin.Python, htmlPath)
	logger.Infof("JsInspector command: %s", commandStr)
	// Execute command
	result, err := s.ExecuteCommand(context.Background(), commandStr, outputChan)
	if err != nil {
		logger.Errorf("JsInspector error: %v", err)
		return nil, nil, err
	}

	// Wait for output collection to complete
	<-done

	return result, outputs, nil
}

// 执行html+JavaScript代码文件的检查逻辑
func (s *CommandService) CodeDiff(code1, code2 string) (string, error) {
	// Collect command output
	var outputs []util.CommandOutput
	outputChan := make(chan util.CommandOutput, 200)

	// Start a goroutine to collect outputs
	done := make(chan struct{})
	go func() {
		for output := range outputChan {
			outputs = append(outputs, output)
		}
		close(done)
	}()

	commandStr := fmt.Sprintf("%s ./scripts/code_diff_merge.py --task compare --input1 %s --input2 %s --base64",
		config.GetConfig().Bin.Python, util.Base64ToolEncode(code1), util.Base64ToolEncode(code2))

	logger.Infof("CodeDiff command: %s", commandStr)
	// Execute command
	result, err := s.ExecuteCommand(context.Background(), commandStr, outputChan)
	if err != nil {
		logger.Errorf("CodeDiff error: %v", err)
		return "", err
	}

	// Wait for output collection to complete
	<-done

	if !result.Success {
		logger.Errorf("CodeDiff error: %s", result.ErrorMessage)
		return "", fmt.Errorf("CodeDiff error: %s", result.ErrorMessage)
	}

	outputData := []string{}
	for _, output := range outputs {
		if output.IsError {
			logger.Errorf("CodeDiff error: %s", output.Line)
			return "", fmt.Errorf("CodeDiff error: %s", output.Line)
		}
		outputData = append(outputData, output.Line)
	}

	return strings.Join(outputData, "\n"), nil
}
