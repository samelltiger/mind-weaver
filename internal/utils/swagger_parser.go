package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// SwaggerData represents the main Swagger document structure
type SwaggerData struct {
	BasePath    string                `json:"basePath"`
	Host        string                `json:"host"`
	Schemes     []string              `json:"schemes"`
	Paths       map[string]PathItem   `json:"paths"`
	Definitions map[string]Definition `json:"definitions"`
}

// PathItem represents a path and its operations
type PathItem map[string]MethodInfo

// MethodInfo represents HTTP method info
type MethodInfo struct {
	Summary     string              `json:"summary"`
	Description string              `json:"description"`
	Parameters  []Parameter         `json:"parameters"`
	Responses   map[string]Response `json:"responses"`
	Tags        []string            `json:"tags"`
}

// Parameter represents an API parameter
type Parameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"`
	Required    bool        `json:"required"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Schema      *Schema     `json:"schema"`
	Example     interface{} `json:"example"`
}

// Schema represents a schema definition
type Schema struct {
	Ref   string   `json:"$ref"`
	Type  string   `json:"type"`
	Items *Schema  `json:"items"`
	AllOf []Schema `json:"allOf"`
}

// Response represents an API response
type Response struct {
	Description string  `json:"description"`
	Schema      *Schema `json:"schema"`
}

// Definition represents a model definition
type Definition struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
}

// Property represents a model property
type Property struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Example     interface{} `json:"example"`
	Ref         string      `json:"$ref"`
	Items       *Schema     `json:"items"`
}

// ApiInterface represents an API interface
type ApiInterface struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

// SwaggerParser provides methods to parse and process Swagger data
type SwaggerParser struct {
	Data        SwaggerData
	BasePath    string
	Host        string
	Schemes     []string
	Definitions map[string]Definition
}

// NewSwaggerParser creates a new SwaggerParser instance
func NewSwaggerParser(data SwaggerData) *SwaggerParser {
	return &SwaggerParser{
		Data:        data,
		BasePath:    data.BasePath,
		Host:        data.Host,
		Schemes:     data.Schemes,
		Definitions: data.Definitions,
	}
}

// GetAllApiInterfaces returns all interfaces in the Swagger document
func (p *SwaggerParser) GetAllApiInterfaces() []ApiInterface {
	interfaces := []ApiInterface{}

	for path, methods := range p.Data.Paths {
		for method := range methods {
			interfaces = append(interfaces, ApiInterface{
				Path:   path,
				Method: method,
			})
		}
	}

	return interfaces
}

// GetApiInterfaceInfo returns information about a specific interface
func (p *SwaggerParser) GetApiInterfaceInfo(path, method string) map[string]interface{} {
	pathItem, exists := p.Data.Paths[path]
	if !exists {
		return nil
	}

	methodInfo, exists := pathItem[strings.ToLower(method)]
	if !exists {
		return nil
	}

	return map[string]interface{}{
		"path":        path,
		"method":      strings.ToUpper(method),
		"info":        methodInfo,
		"parameters":  methodInfo.Parameters,
		"responses":   methodInfo.Responses,
		"summary":     methodInfo.Summary,
		"description": methodInfo.Description,
		"tags":        methodInfo.Tags,
	}
}

// GenerateCurlCommand generates a cURL command for a given interface
func (p *SwaggerParser) GenerateCurlCommand(interfaceInfo map[string]interface{}) string {
	method := interfaceInfo["method"].(string)
	path := interfaceInfo["path"].(string)

	scheme := "http"
	if len(p.Schemes) > 0 {
		scheme = p.Schemes[0]
	}

	fullURL := fmt.Sprintf("%s://%s%s%s", scheme, p.Host, p.BasePath, path)

	headers := []string{}
	bodyParams := []string{}
	queryParams := []string{}

	parameters, _ := interfaceInfo["parameters"].([]Parameter)
	for _, param := range parameters {
		if param.In == "header" {
			headers = append(headers, fmt.Sprintf("-H '%s: %s'", param.Name, p.getParamExample(param)))
		} else if param.In == "query" {
			queryParams = append(queryParams, fmt.Sprintf("%s=%s", param.Name, p.getParamExample(param)))
		} else if param.In == "body" {
			bodyParams = append(bodyParams, p.generateBodyExample(param))
		}
	}

	if len(queryParams) > 0 {
		fullURL += "?" + strings.Join(queryParams, "&")
	}

	curlCmd := fmt.Sprintf("curl -X %s '%s'", method, fullURL)

	if len(headers) > 0 {
		curlCmd += " " + strings.Join(headers, " ")
	}

	if len(bodyParams) > 0 && (method == "POST" || method == "PUT" || method == "PATCH") {
		curlCmd += fmt.Sprintf(" -d '%s'", bodyParams[0])
	}

	return curlCmd
}

// getParamExample returns an example value for a parameter
func (p *SwaggerParser) getParamExample(param Parameter) string {
	if param.Example != nil {
		return fmt.Sprintf("%v", param.Example)
	}

	if param.Schema != nil && param.Schema.Ref != "" {
		ref := getLastPart(param.Schema.Ref)
		return p.getDefinitionExample(ref, false)
	}

	if param.Type != "" {
		switch param.Type {
		case "string":
			return "string"
		case "integer":
			return "0"
		case "boolean":
			return "true"
		case "array":
			return "[]"
		case "object":
			return "{}"
		}
	}

	return ""
}

// generateBodyExample generates an example request body
func (p *SwaggerParser) generateBodyExample(param Parameter) string {
	if param.Schema != nil {
		if param.Schema.Ref != "" {
			ref := getLastPart(param.Schema.Ref)
			example, _ := json.MarshalIndent(p.getDefinitionExampleAsMap(ref, true), "", "  ")
			return string(example)
		} else if param.Schema.Type == "array" && param.Schema.Items != nil && param.Schema.Items.Ref != "" {
			ref := getLastPart(param.Schema.Items.Ref)
			example := []interface{}{p.getDefinitionExampleAsMap(ref, true)}
			exampleJSON, _ := json.MarshalIndent(example, "", "  ")
			return string(exampleJSON)
		}
	}

	return "{}"
}

// getDefinitionExample returns a string representation of a definition example
func (p *SwaggerParser) getDefinitionExample(definitionName string, fullExample bool) string {
	result, _ := json.Marshal(p.getDefinitionExampleAsMap(definitionName, fullExample))
	return string(result)
}

// getDefinitionExampleAsMap returns a map representation of a definition example
func (p *SwaggerParser) getDefinitionExampleAsMap(definitionName string, fullExample bool) map[string]interface{} {
	definition, exists := p.Definitions[definitionName]
	if !exists {
		return map[string]interface{}{}
	}

	example := make(map[string]interface{})

	for propName, propInfo := range definition.Properties {
		if propInfo.Example != nil {
			example[propName] = propInfo.Example
		} else if propInfo.Ref != "" {
			ref := getLastPart(propInfo.Ref)
			example[propName] = p.getDefinitionExampleAsMap(ref, false)
		} else if propInfo.Type != "" {
			if propInfo.Type == "array" && propInfo.Items != nil && propInfo.Items.Ref != "" {
				ref := getLastPart(propInfo.Items.Ref)
				example[propName] = []interface{}{p.getDefinitionExampleAsMap(ref, false)}
			} else {
				switch propInfo.Type {
				case "string":
					example[propName] = "string"
				case "integer":
					example[propName] = 0
				case "boolean":
					example[propName] = true
				case "object":
					example[propName] = map[string]interface{}{}
				default:
					example[propName] = ""
				}
			}
		} else if fullExample {
			example[propName] = ""
		}
	}

	return example
}

// generateApiInterfaceDoc generates documentation for an interface
func (p *SwaggerParser) GenerateApiInterfaceDoc(interfaceInfo map[string]interface{}) string {
	var doc strings.Builder

	summary, _ := interfaceInfo["summary"].(string)
	description, _ := interfaceInfo["description"].(string)
	method, _ := interfaceInfo["method"].(string)
	path, _ := interfaceInfo["path"].(string)

	doc.WriteString(fmt.Sprintf("## %s\n\n", summary))
	doc.WriteString(fmt.Sprintf("**接口描述**: %s\n\n", description))
	doc.WriteString(fmt.Sprintf("**请求方法**: `%s`\n\n", method))
	doc.WriteString(fmt.Sprintf("**接口路径**: `%s`\n\n", path))

	// Request parameters
	doc.WriteString("### 请求参数\n\n")

	parameters, _ := interfaceInfo["parameters"].([]Parameter)
	if len(parameters) == 0 {
		doc.WriteString("无\n\n")
	} else {
		doc.WriteString("| 参数名 | 位置 | 类型 | 必填 | 描述 | 示例 |\n")
		doc.WriteString("|--------|------|------|------|------|------|\n")

		for _, param := range parameters {
			paramType := param.Type

			if param.Schema != nil {
				if param.Schema.Ref != "" {
					paramType = getLastPart(param.Schema.Ref)
				} else if param.Schema.Type != "" {
					paramType = param.Schema.Type
					if paramType == "array" && param.Schema.Items != nil {
						if param.Schema.Items.Ref != "" {
							paramType = fmt.Sprintf("array[%s]", getLastPart(param.Schema.Items.Ref))
						} else {
							paramType = fmt.Sprintf("array[%s]", param.Schema.Items.Type)
						}
					}
				}
			}

			doc.WriteString(fmt.Sprintf("| %s | %s | %s | %t | %s | %s |\n",
				param.Name, param.In, paramType, param.Required, param.Description, p.getParamExample(param)))
		}
		doc.WriteString("\n")
	}

	// Response parameters
	doc.WriteString("### 响应参数\n\n")

	responses, _ := interfaceInfo["responses"].(map[string]Response)
	var successResponse Response

	if response, exists := responses["200"]; exists {
		successResponse = response
	} else if response, exists := responses["201"]; exists {
		successResponse = response
	}

	if successResponse.Schema == nil {
		doc.WriteString("无\n\n")
	} else {
		if successResponse.Schema.Ref != "" {
			ref := getLastPart(successResponse.Schema.Ref)
			doc.WriteString(p.generateResponseParamsDoc(ref, 0))
		} else if len(successResponse.Schema.AllOf) > 0 {
			doc.WriteString(p.processAllOfSchema(successResponse.Schema.AllOf))
		} else if successResponse.Schema.Type == "array" && successResponse.Schema.Items != nil {
			if successResponse.Schema.Items.Ref != "" {
				ref := getLastPart(successResponse.Schema.Items.Ref)
				doc.WriteString(fmt.Sprintf("返回类型: array[%s]\n\n", ref))
				doc.WriteString(p.generateResponseParamsDoc(ref, 0))
			}
		} else {
			doc.WriteString("无\n\n")
		}
	}

	// Request example curl命令示例
	// doc.WriteString("### 请求示例\n\n")
	// doc.WriteString(fmt.Sprintf("```bash\n%s\n```\n\n", p.GenerateCurlCommand(interfaceInfo)))

	// If there's a request body, add JSON example
	var bodyParams []Parameter
	for _, param := range parameters {
		if param.In == "body" {
			bodyParams = append(bodyParams, param)
		}
	}

	if len(bodyParams) > 0 {
		doc.WriteString("请求体示例:\n\n")
		doc.WriteString(fmt.Sprintf("```json\n%s\n```\n\n", p.generateBodyExample(bodyParams[0])))
	}

	// Response example
	doc.WriteString("### 响应示例\n\n")

	if successResponse.Schema != nil {
		if successResponse.Schema.Ref != "" {
			ref := getLastPart(successResponse.Schema.Ref)
			example, _ := json.MarshalIndent(p.getDefinitionExampleAsMap(ref, true), "", "  ")
			doc.WriteString(fmt.Sprintf("```json\n%s\n```\n\n", string(example)))
		} else if len(successResponse.Schema.AllOf) > 0 {
			example, _ := json.MarshalIndent(p.getAllOfExample(successResponse.Schema.AllOf), "", "  ")
			doc.WriteString(fmt.Sprintf("```json\n%s\n```\n\n", string(example)))
		} else if successResponse.Schema.Type == "array" && successResponse.Schema.Items != nil {
			if successResponse.Schema.Items.Ref != "" {
				ref := getLastPart(successResponse.Schema.Items.Ref)
				example := []interface{}{p.getDefinitionExampleAsMap(ref, true)}
				exampleJSON, _ := json.MarshalIndent(example, "", "  ")
				doc.WriteString(fmt.Sprintf("```json\n%s\n```\n\n", string(exampleJSON)))
			} else {
				doc.WriteString("```json\n[]\n```\n\n")
			}
		} else {
			doc.WriteString("无\n\n")
		}
	} else {
		doc.WriteString("无\n\n")
	}

	return doc.String()
}

// processAllOfSchema processes an allOf schema structure
func (p *SwaggerParser) processAllOfSchema(allOf []Schema) string {
	var doc strings.Builder

	for _, item := range allOf {
		if item.Ref != "" {
			ref := getLastPart(item.Ref)
			doc.WriteString(p.generateResponseParamsDoc(ref, 0))
		}
		// Add implementation for properties if needed
	}

	return doc.String()
}

// generateResponseParamsDocFromProperties generates response parameter docs from properties
func (p *SwaggerParser) generateResponseParamsDocFromProperties(properties map[string]Property) string {
	if len(properties) == 0 {
		return ""
	}

	var doc strings.Builder
	doc.WriteString("| 参数名 | 类型 | 描述 | 示例 |\n")
	doc.WriteString("|--------|------|------|------|\n")

	for propName, propInfo := range properties {
		propType := propInfo.Type
		if propInfo.Ref != "" {
			ref := getLastPart(propInfo.Ref)
			propType = ref
		} else if propType == "array" && propInfo.Items != nil {
			if propInfo.Items.Ref != "" {
				ref := getLastPart(propInfo.Items.Ref)
				propType = fmt.Sprintf("array[%s]", ref)
			} else {
				propType = fmt.Sprintf("array[%s]", propInfo.Items.Type)
			}
		}

		example := ""
		if propInfo.Example != nil {
			example = fmt.Sprintf("%v", propInfo.Example)
		}

		doc.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			propName, propType, propInfo.Description, example))

		// Handle nested objects
		if propInfo.Ref != "" {
			ref := getLastPart(propInfo.Ref)
			doc.WriteString(p.generateResponseParamsDoc(ref, 1))
		} else if strings.HasPrefix(propType, "array[") && propInfo.Items != nil && propInfo.Items.Ref != "" {
			ref := getLastPart(propInfo.Items.Ref)
			doc.WriteString(p.generateResponseParamsDoc(ref, 1))
		}
	}

	return doc.String()
}

// getAllOfExample gets an example for an allOf structure
func (p *SwaggerParser) getAllOfExample(allOf []Schema) map[string]interface{} {
	example := make(map[string]interface{})

	for _, item := range allOf {
		if item.Ref != "" {
			ref := getLastPart(item.Ref)
			refExample := p.getDefinitionExampleAsMap(ref, true)
			for k, v := range refExample {
				example[k] = v
			}
		}
		// Add implementation for properties if needed
	}

	return example
}

// generateResponseParamsDoc generates documentation for response parameters
func (p *SwaggerParser) generateResponseParamsDoc(definitionName string, level int) string {
	definition, exists := p.Definitions[definitionName]
	if !exists {
		return ""
	}

	properties := definition.Properties
	if len(properties) == 0 {
		return ""
	}

	var doc strings.Builder
	indent := strings.Repeat("  ", level)

	if level == 0 {
		doc.WriteString("| 参数名 | 类型 | 描述 | 示例 |\n")
		doc.WriteString("|--------|------|------|------|\n")
	}

	for propName, propInfo := range properties {
		propType := propInfo.Type
		if propInfo.Ref != "" {
			ref := getLastPart(propInfo.Ref)
			propType = ref
		} else if propType == "array" && propInfo.Items != nil {
			if propInfo.Items.Ref != "" {
				ref := getLastPart(propInfo.Items.Ref)
				propType = fmt.Sprintf("array[%s]", ref)
			} else {
				propType = fmt.Sprintf("array[%s]", propInfo.Items.Type)
			}
		}

		example := ""
		if propInfo.Example != nil {
			example = fmt.Sprintf("%v", propInfo.Example)
		}

		doc.WriteString(fmt.Sprintf("| %s%s | %s | %s | %s |\n",
			indent, propName, propType, propInfo.Description, example))

		// Handle nested objects
		if propInfo.Ref != "" {
			ref := getLastPart(propInfo.Ref)
			doc.WriteString(p.generateResponseParamsDoc(ref, level+1))
		} else if strings.HasPrefix(propType, "array[") && propInfo.Items != nil && propInfo.Items.Ref != "" {
			ref := getLastPart(propInfo.Items.Ref)
			doc.WriteString(p.generateResponseParamsDoc(ref, level+1))
		}
	}

	return doc.String()
}

// Helper function to get the last part of a reference string
func getLastPart(ref string) string {
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}

// LoadSwaggerData loads Swagger data from a file or a JSON string
func LoadSwaggerData(source string, isBase64 bool) (SwaggerData, error) {
	var data SwaggerData
	var jsonData []byte
	var err error

	if _, err = os.Stat(source); err == nil {
		// It's a file
		jsonData, err = ioutil.ReadFile(source)
		if err != nil {
			return data, fmt.Errorf("cannot read file: %v", err)
		}
	} else {
		// It's a string
		if isBase64 {
			decoded, err := base64.StdEncoding.DecodeString(source)
			if err != nil {
				return data, fmt.Errorf("cannot decode base64: %v", err)
			}
			jsonData = decoded
		} else {
			jsonData = []byte(source)
		}
	}

	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return data, fmt.Errorf("cannot parse Swagger data: %v", err)
	}

	return data, nil
}

// LoadApiInterfaces loads interfaces from a file or a JSON string
func LoadApiInterfaces(source string, isBase64 bool) ([]ApiInterface, error) {
	var interfaces []ApiInterface

	if source == "" {
		return interfaces, nil
	}

	var jsonData []byte
	var err error

	if _, err = os.Stat(source); err == nil {
		// It's a file
		jsonData, err = ioutil.ReadFile(source)
		if err != nil {
			return interfaces, fmt.Errorf("cannot read file: %v", err)
		}
	} else {
		// It's a string
		if isBase64 {
			decoded, err := base64.StdEncoding.DecodeString(source)
			if err != nil {
				return interfaces, fmt.Errorf("cannot decode base64: %v", err)
			}
			jsonData = decoded
		} else {
			jsonData = []byte(source)
		}
	}

	err = json.Unmarshal(jsonData, &interfaces)
	if err != nil {
		return interfaces, fmt.Errorf("cannot parse interfaces: %v", err)
	}

	return interfaces, nil
}

// GenerateAPIDocumentation generates API documentation from Swagger data and interfaces
func GenerateAPIDocumentation(swaggerData SwaggerData, interfaces []ApiInterface, outputFile string, encodeBase64 bool) (string, error) {
	parser := NewSwaggerParser(swaggerData)

	var docContent strings.Builder
	docContent.WriteString("# API 接口文档\n\n")

	for i, iface := range interfaces {
		interfaceInfo := parser.GetApiInterfaceInfo(iface.Path, iface.Method)
		if interfaceInfo == nil {
			fmt.Fprintf(os.Stderr, "警告: 接口 %s %s 未找到，跳过\n", iface.Method, iface.Path)
			continue
		}

		docContent.WriteString(parser.GenerateApiInterfaceDoc(interfaceInfo))

		if i < len(interfaces)-1 {
			docContent.WriteString("\n---\n\n")
		}
	}

	result := docContent.String()

	if outputFile != "" {
		err := ioutil.WriteFile(outputFile, []byte(result), 0644)
		if err != nil {
			return "", fmt.Errorf("cannot write to file: %v", err)
		}
		return "", nil
	}

	if encodeBase64 {
		return base64.StdEncoding.EncodeToString([]byte(result)), nil
	}

	return result, nil
}
