package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	// Step 0: Find the only EPUB file
	files, err := filepath.Glob("*.epub")
	if err != nil {
		panic(err)
	}
	if len(files) != 1 {
		panic("There must be exactly one .epub file in the current directory.")
	}

	inputEPUB := files[0]
	tempOutput := inputEPUB + ".tmp.epub"

	tempDir, err := os.MkdirTemp("", "epubconvert")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	// Step 1: Unzip the EPUB
	err = unzip(inputEPUB, tempDir)
	if err != nil {
		panic(err)
	}

	// Step 2: Convert text content to Traditional Chinese
	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".xhtml") ||
			strings.HasSuffix(path, ".html") ||
			strings.HasSuffix(path, ".xml") ||
			strings.HasSuffix(path, ".ncx") ||
			strings.HasSuffix(path, ".opf") {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			converted, err := convertToTraditional(string(content))
			if err != nil {
				return err
			}

			err = os.WriteFile(path, []byte(converted), 0644)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	// Step 3: Zip it back to EPUB (temporary output)
	err = zipDir(tempDir, tempOutput)
	if err != nil {
		panic(err)
	}

	// Step 4: Replace original file with converted version
	err = os.Remove(inputEPUB)
	if err != nil {
		panic(err)
	}

	err = os.Rename(tempOutput, inputEPUB)
	if err != nil {
		panic(err)
	}

	fmt.Println("Conversion complete:", inputEPUB)
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func zipDir(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == source {
			return nil
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = relPath
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			return err
		}

		return nil
	})
}

// Wrapper for opencc CLI
func convertToTraditional(input string) (string, error) {
	cmd := exec.Command("opencc", "-c", "s2hk.json")
	cmd.Stdin = strings.NewReader(input)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}
