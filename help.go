package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// 嵌入安装脚本
//
//go:embed service.sh
var serviceFile embed.FS

func installServiceScript() error {
	// 获取可执行文件所在目录
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	exeDir := filepath.Dir(exePath)

	// 读取嵌入的脚本内容
	scriptContent, err := fs.ReadFile(serviceFile, "service.sh")
	if err != nil {
		return fmt.Errorf("failed to read embedded script: %w", err)
	}

	// 构建目标文件路径
	targetPath := filepath.Join(exeDir, "service.sh")

	// 写入文件
	err = os.WriteFile(targetPath, scriptContent, 0755)
	if err != nil {
		return fmt.Errorf("failed to write script file: %w", err)
	}

	// 设置文件权限 (虽然WriteFile已经设置了，但为了确保再次设置)
	err = os.Chmod(targetPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to set script permissions: %w", err)
	}

	return nil
}
