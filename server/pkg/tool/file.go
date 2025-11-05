package tool

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// HasDir 判断文件夹是否存在
func HasDir(path string) (bool, error) {
	_, _err := os.Stat(path)
	if _err == nil {
		return true, nil
	}
	if os.IsNotExist(_err) {
		return false, nil
	}
	return false, _err
}

// CreateDir 创建文件夹
func CreateDir(path string) error {
	_exist, _err := HasDir(path)
	if _err != nil {
		return _err
	}
	if _exist {
		return fmt.Errorf("文件夹已存在！")
	}
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return _err
	}
	return nil
}

// FileExists 文件是否存在
func FileExists(name string) bool {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func GetCurDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return ""
	}
	return strings.Replace(dir, "\\", "/", -1) + "/"
}

// WriteFile 写内容进入文件
func WriteFile(file *os.File, content string) error {
	if _, err := file.Write([]byte(content)); err != nil {
		return err
	}
	return nil
}

// CreateAndWriteFile 写内容进入文件
func CreateAndWriteFile(absTargetFilePath, content string) error {
	f, err := os.Create(absTargetFilePath)
	if err != nil {
		return err
	}
	return WriteFile(f, content)
}

// AppendWriteFile 写内容进入文件
func AppendWriteFile(absTargetFilePath, content string) error {
	// 打开文件
	f, err := os.OpenFile(absTargetFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
	}()
	return WriteFile(f, content)
}
