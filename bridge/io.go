package bridge

import (
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func (a *App) Writefile(path string, content string) FlagResult {
	log.Printf("Writefile: %s", path)

	path, err := GetPath(path)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	err = os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	return FlagResult{true, "Success"}
}

func (a *App) Readfile(path string) FlagResult {
	log.Printf("Readfile: %s", path)

	path, err := GetPath(path)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	return FlagResult{true, string(b)}
}

func (a *App) Movefile(source string, target string) FlagResult {
	log.Printf("Movefile: %s -> %s", source, target)

	source, err := GetPath(source)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	target, err = GetPath(target)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	err = os.Rename(source, target)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	return FlagResult{true, "Success"}
}

func UNUSED(x ...interface{}) {}

func (a *App) Copyfile(source string, target string) FlagResult {
	log.Printf("Copyfile: %s -> %s", source, target)

	source, err := GetPath(source)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	target, err = GetPath(target)
	if err != nil {
		return FlagResult{false, err.Error()}
	}
	fmt.Println("Copyfile:", source, target)
	
	src, err := os.Open(source)
	if err != nil {
		return FlagResult{false, err.Error()}
	}
	defer src.Close()

	dest, err := os.Create(target)
	if err != nil {
		return FlagResult{false, err.Error()}
	}
	defer dest.Close()
	nBytes, err := io.Copy(dest, src)
	UNUSED(nBytes)
	if err != nil {
		return FlagResult{false, err.Error()}
	}
	return FlagResult{true, "Success"}
}

func (a *App) Removefile(path string) FlagResult {
	log.Printf("RemoveFile: %s", path)

	path, err := GetPath(path)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	err = os.RemoveAll(path)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	return FlagResult{true, "Success"}
}

func (a *App) Makedir(path string) FlagResult {
	log.Printf("Makedir: %s", path)

	path, err := GetPath(path)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return FlagResult{false, err.Error()}
	}
	return FlagResult{true, "Success"}
}

func (a *App) UnzipZIPFile(path string, output string) FlagResult {
	log.Printf("UnzipZIPFile: %s -> %s", path, output)

	path, err := GetPath(path)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	output, err = GetPath(output)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	archive, err := zip.OpenReader(path)
	if err != nil {
		return FlagResult{false, err.Error()}
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(output, f.Name)

		if !strings.HasPrefix(filePath, filepath.Clean(output)+string(os.PathSeparator)) {
			log.Println("UnzipZIPFile: invalid file path")
			return FlagResult{false, err.Error()}
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return FlagResult{false, err.Error()}
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return FlagResult{false, err.Error()}
		}
		defer dstFile.Close()

		fileInArchive, err := f.Open()
		if err != nil {
			return FlagResult{false, err.Error()}
		}
		defer fileInArchive.Close()

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return FlagResult{false, err.Error()}
		}
	}
	return FlagResult{true, "Success"}
}

func (a *App) UnzipGZFile(path string, output string) FlagResult {
	log.Printf("UnzipGZFile: %s -> %s", path, output)

	path, err := GetPath(path)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	output, err = GetPath(output)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	gzipFile, err := os.Open(path)
	if err != nil {
		return FlagResult{false, err.Error()}
	}
	defer gzipFile.Close()

	outputFile, err := os.Create(output)
	if err != nil {
		return FlagResult{false, err.Error()}
	}
	defer outputFile.Close()

	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		return FlagResult{false, err.Error()}
	}
	defer gzipReader.Close()

	_, err = io.Copy(outputFile, gzipReader)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	return FlagResult{true, "Success"}
}

func (a *App) FileExists(path string) FlagResult {
	log.Printf("FileExists: %s", path)

	path, err := GetPath(path)
	if err != nil {
		return FlagResult{false, err.Error()}
	}
	_, err = os.Stat(path)
	if err == nil {
		return FlagResult{true, "true"}
	} else if os.IsNotExist(err) {
		return FlagResult{true, "false"}
	}
	return FlagResult{false, err.Error()}
}
