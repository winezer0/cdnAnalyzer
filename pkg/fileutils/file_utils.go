package fileutils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// IsEmptyFile 检查文件是否为空或不存在
func IsEmptyFile(filename string) bool {
	// Get file info
	fileInfo, err := os.Stat(filename)
	if os.IsNotExist(err) || fileInfo.Size() == 0 {
		return true
	}
	return false
}

// IsFileExists 判断是否是普通文件存在
func IsFileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// IsNotExists 判断是否是普通文件不存在
func IsNotExists(filename string) bool {
	return !IsFileExists(filename)
}

// IsDirExists 判断是否是目录存在
func IsDirExists(dirname string) bool {
	info, err := os.Stat(dirname)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// FindFile 按照优先级查找文件是否存在，返回所有找到的完整路径
func FindFile(filename string) []string {
	var searchPaths []string

	// 1. 获取当前工作目录
	wd, err := os.Getwd()
	if err == nil {
		searchPaths = append(searchPaths, wd)
	}

	// 2. 程序入口目录（可执行文件所在目录）
	exePath, err := os.Executable()
	if err == nil {
		searchPaths = append(searchPaths, filepath.Dir(exePath))
	}

	// 3. 脚本所在目录（调用此函数的源码目录）
	_, callerFile, _, _ := runtime.Caller(0)
	callerDir := filepath.Dir(callerFile)
	searchPaths = append(searchPaths, callerDir)

	// 4. 用户主目录
	homeDir, err := os.UserHomeDir()
	if err == nil {
		searchPaths = append(searchPaths, homeDir)
	}

	// 去重并规范化路径
	seen := make(map[string]bool)
	finalPaths := make([]string, 0)
	for _, path := range searchPaths {
		normalPath := filepath.Clean(path)
		if !seen[normalPath] {
			seen[normalPath] = true
			finalPaths = append(finalPaths, normalPath)
		}
	}

	// 如果是绝对路径，直接检查是否存在
	if filepath.IsAbs(filename) {
		if _, err := os.Stat(filename); err == nil {
			return []string{filename}
		}
		return nil
	}

	// 查找文件
	var foundPaths []string
	for _, dir := range finalPaths {
		fullPath := filepath.Join(dir, filename)
		if _, err := os.Stat(fullPath); err == nil {
			foundPaths = append(foundPaths, fullPath)
		}
	}

	return foundPaths
}

// GetAbsolutePath 返回给定文件或目录的绝对路径
func GetAbsolutePath(relativePath string) (string, error) {
	absPath, err := filepath.Abs(relativePath)
	if err != nil {
		return "", err
	}
	return absPath, nil
}

// GetFileDirectory 返回给定文件路径的目录部分
func GetFileDirectory(filePath string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("文件路径不能为空")
	}

	dir := filepath.Dir(filePath)
	return dir, nil
}

// GetUserSubDir 返回当前用户的某个子目录的完整路径，支持多层目录拼接
func GetUserSubDir(subDirs ...string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "" // 静默失败，返回空字符串
	}
	return filepath.Join(append([]string{homeDir}, subDirs...)...)
}

// JoinPath 拼接多个路径元素，返回平台兼容的完整路径
func JoinPath(elem ...string) string {
	return filepath.Join(elem...)
}

func IsAbsolutePath(path string) bool {
	return filepath.IsAbs(path)
}

func IsRelativePath(path string) bool {
	return !filepath.IsAbs(path)
}
