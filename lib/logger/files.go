package logger

import (
	"fmt"
	"os"
)

// checkNotExist 检查文件或目录是否存在。
func checkNotExist(src string) bool {
	_, err := os.Stat(src)
	return os.IsNotExist(err)
}

// checkPermission 检查是否有权限访问文件或目录。
func checkPermission(src string) bool {
	_, err := os.Stat(src)
	return os.IsPermission(err)
}

// isNotExistMkDir 如果目录不存在，则创建目录。
func isNotExistMkDir(src string) error {
	if notExit := checkNotExist(src); notExit == true {
		if err := mkDir(src); err != nil {
			return err
		}
	}
	return nil
}

// mkDir 创建目录
func mkDir(src string) error {
	err := os.MkdirAll(src, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// mustOpen 打开文件，如果文件不存在，则创建文件。
func mustOpen(fileName, dir string) (*os.File, error) {
	// 检查目录是否有权限
	perm := checkPermission(dir)
	if perm == true {
		return nil, fmt.Errorf("permission denied dir: %s", dir)
	}
	// 如果目录不存在，则创建目录
	err := isNotExistMkDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error during make dir %s, err: %s", dir, err)
	}
	f, err := os.OpenFile(dir+string(os.PathSeparator)+fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("fail to open file, err: %s", err)
	}
	return f, nil
}
