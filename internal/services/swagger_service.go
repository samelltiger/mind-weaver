package services

import (
	"fmt"
	"path/filepath"

	"mind-weaver/internal/utils"
)

// SwaggerService provides swagger document generation functionality
type SwaggerService struct {
}

// NewSwaggerService creates a new SwaggerService
func NewSwaggerService() *SwaggerService {
	return &SwaggerService{}
}

// ListInterfaces lists all interfaces in a swagger document
func (s *SwaggerService) ListInterfaces(swaggerSource string, isBase64 bool, outputFile string, outputBase64 bool) ([]utils.ApiInterface, error) {
	// Load swagger data
	swaggerData, err := utils.LoadSwaggerData(swaggerSource, isBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to load swagger data: %v", err)
	}

	// Get all interfaces
	parser := utils.NewSwaggerParser(swaggerData)
	interfaces := parser.GetAllApiInterfaces()

	// Output interfaces
	if outputFile != "" {
		// Create the directory if it doesn't exist
		dir := filepath.Dir(outputFile)
		err := utils.CreateDirectoryIfNotExist(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to create directory: %v", err)
		}

		// Write to file
		err = utils.WriteJSONToFile(interfaces, outputFile)
		if err != nil {
			return nil, fmt.Errorf("failed to write to file: %v", err)
		}

		fmt.Printf("Interfaces list saved to %s\n", outputFile)
		return nil, nil
	}

	// Return as string
	// output, err := utils.InterfacesToJSON(interfaces, outputBase64)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to convert to JSON: %v", err)
	// }

	return interfaces, nil
}

// GenerateDoc generates API documentation
func (s *SwaggerService) GenerateDoc(swaggerSource string, swaggerBase64 bool,
	interfaces []utils.ApiInterface, interfacesBase64 bool,
	outputFile string, outputBase64 bool) (string, error) {

	// Load swagger data
	swaggerData, err := utils.LoadSwaggerData(swaggerSource, swaggerBase64)
	if err != nil {
		return "", fmt.Errorf("failed to load swagger data: %v", err)
	}

	// Load interfaces
	// interfaces, err := utils.LoadApiInterfaces(interfacesSource, interfacesBase64)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to load interfaces: %v", err)
	// }

	// Generate API documentation
	if outputFile != "" {
		// Create the directory if it doesn't exist
		dir := filepath.Dir(outputFile)
		err := utils.CreateDirectoryIfNotExist(dir)
		if err != nil {
			return "", fmt.Errorf("failed to create directory: %v", err)
		}
	}

	result, err := utils.GenerateAPIDocumentation(swaggerData, interfaces, outputFile, outputBase64)
	if err != nil {
		return "", fmt.Errorf("failed to generate documentation: %v", err)
	}

	return result, nil
}
