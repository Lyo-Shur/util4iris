package form

import (
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

// 文件保存配置
type SaveConfig struct {
	// 服务器访问路径
	ServerPath string
	// 磁盘保存路径
	DiskPath string
	// 计算文件保存路径 参数 请求中的文件名
	// 返回值 次级路径 文件名
	GetSavePath func(FileName string) (string, string, error)
}

// 复制-文件保存配置-方法
func (s *SaveConfig) Clone() SaveConfig {
	sc := SaveConfig{}
	sc.ServerPath = s.ServerPath
	sc.DiskPath = s.DiskPath
	sc.GetSavePath = s.GetSavePath
	return sc
}

// 文件持有者
type FileHolder struct {
	m map[string][]*multipart.FileHeader
}

// 获取保存的文件
func (f *FileHolder) GetFile(name string) File {
	fs := f.GetFiles(name)
	if fs == nil {
		return File{}
	}
	return fs[0]
}

// 获取保存的文件们
func (f *FileHolder) GetFiles(name string) []File {
	fhs := f.m[name]
	if fhs == nil {
		return nil
	}
	l := len(fhs)
	fs := make([]File, l)
	for i := 0; i < l; i++ {
		f := File{}
		f.FH = fhs[i]
		fs[i] = f
	}
	return fs
}

// 上传的表单文件
type File struct {
	FH *multipart.FileHeader
}

// 是否存在
func (file *File) Exist() bool {
	return file.FH != nil
}

// 保存文件
func (file *File) Save(sc SaveConfig) (string, string, error) {
	return saveFile(file.FH, sc)
}

// 保存文件，返回（访问路径 保存路径 错误）
func saveFile(fh *multipart.FileHeader, sc SaveConfig) (string, string, error) {
	// 根据文件名获取次级保存路径、文件名、错误信息
	secondPath, fileName, err := sc.GetSavePath(fh.Filename)
	if err != nil {
		return "", "", err
	}
	// 保存路径(不含文件名) 合并主保存路径和次级保存路径
	filePath := filepath.Join(sc.DiskPath, secondPath)
	// 保存路径 尝试保存文件到硬盘
	savePath, err := writeFile(fh, filePath, fileName)
	if err != nil {
		return "", "", err
	}
	// 访问路径
	visitPath := ""
	// 计算访问路径
	if sc.DiskPath != "" {
		visitPath = strings.Replace(savePath, "\\", "/", -1)
		visitPath = strings.Replace(visitPath, sc.DiskPath, "", -1)
	}
	return sc.ServerPath + visitPath, savePath, nil
}

// 将文件写在硬盘上
func writeFile(fh *multipart.FileHeader, filePath string, fileName string) (string, error) {
	// 源文件
	src, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer func() {
		err := src.Close()
		if err != nil {
			log.Println(err)
		}
	}()
	// 保存路径检查
	err = os.MkdirAll(filePath, os.FileMode(0777))
	if err != nil {
		return "", err
	}
	// 保存文件
	savePath := filepath.Join(filePath, fileName)
	out, err := os.OpenFile(savePath, os.O_WRONLY|os.O_CREATE, os.FileMode(0666))
	if err != nil {
		return "", err
	}
	defer func() {
		err := out.Close()
		if err != nil {
			log.Println(err)
		}
	}()
	_, err = io.Copy(out, src)
	return savePath, err
}
