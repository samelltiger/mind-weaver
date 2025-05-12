package main

import (
	"log"
	"os/exec"
)

func main() {
	cmd := exec.Command("swag",
		"init",
		"-g", "cmd/main.go",
		"-o", "internal/api/docs",
		"--exclude", "internal/api/init.go", // 可以排除特定文件
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("swag init failed: %v\n%s", err, output)
	}

	log.Println("Swagger docs generated successfully")
}

// swag init -g cmd/main.go -o internal/api/docs
