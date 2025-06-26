package main

import (
	"archive/zip"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/minio/madmin-go/v4/estream"
)

var (
	publicKeyPath = flag.String("public-key", "./public.pem", "public key")
)

func main() {
	flag.Parse()
	publicKeyFile := path.Clean(*publicKeyPath)

	if len(flag.Args()) == 0 {
		fmt.Println("missing source file or directory")
		return
	}

	if len(flag.Args()) > 1 {
		fmt.Println("only 1 file or directory can be encrypted")
		return
	}

	sourceDir := flag.Args()[0]
	sourceDir = path.Clean(sourceDir)

	encFilename := fmt.Sprintf("%s_alt.enc", strings.TrimSuffix(filepath.Base(sourceDir), ".zip"))

	fw, err := os.Create(encFilename)
	if err != nil {
		fmt.Printf("error on creating a file:%s, error:%s\n", encFilename, err.Error())
		return
	}
	defer fw.Close()

	// load public certificate
	stream := estream.NewWriter(fw)
	pubKey, err := toPublicKey(publicKeyFile)
	if err != nil {
		fmt.Printf("error on reading public key file:%s, error:%s\n", publicKeyFile, err.Error())
		return
	}
	err = stream.AddKeyEncrypted(pubKey)
	if err != nil {
		fmt.Printf("error on adding encryption key, error:%s\n", err.Error())
		return
	}

	if strings.HasSuffix(sourceDir, ".zip") {
		zipReader, err := zip.OpenReader(sourceDir)
		if err != nil {
			fmt.Println("error opening zip file:", err)
			return
		}
		defer zipReader.Close()

		for _, file := range zipReader.File {
			fileReader, err := file.Open()
			if err != nil {
				fmt.Printf("error opening file in zip:%s, error:%v\n", file.Name, err)
				return
			}
			// close the fileReader for each file
			defer fileReader.Close()

			f, err := stream.AddEncryptedStream(filepath.Base(file.Name), nil)
			if err != nil {
				fmt.Printf("error on adding a file to encryption stream:%s, error:%s\n", file.Name, err.Error())
				return
			}

			_, err = io.Copy(f, fileReader)
			if err != nil {
				fmt.Printf("error on adding a data to encryption stream:%s, error:%s\n", file.Name, err.Error())
				return
			}
			if err = f.Close(); err != nil {
				fmt.Printf("error on closing a stream:%s, error:%s\n", file.Name, err.Error())
				return
			}
		}
	} else {
		files, err := os.ReadDir(sourceDir)
		if err != nil {
			fmt.Printf("error on reading given directory:%s, error:%s\n", sourceDir, err.Error())
			return
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			_filename := path.Join(sourceDir, file.Name())
			data, err := os.ReadFile(_filename)
			if err != nil {
				fmt.Printf("error on reading a file:%s, error:%s\n", _filename, err.Error())
				return
			}
			f, err := stream.AddEncryptedStream(file.Name(), nil)
			if err != nil {
				fmt.Printf("error on adding a file to encryption stream:%s, error:%s\n", _filename, err.Error())
				return
			}
			_, err = f.Write(data)
			if err != nil {
				fmt.Printf("error on adding a data to encryption stream:%s, error:%s\n", _filename, err.Error())
				return
			}
			if err = f.Close(); err != nil {
				fmt.Printf("error on closing a stream:%s, error:%s\n", _filename, err.Error())
				return
			}
		}
	}
}

func toPublicKey(publicKeyFile string) (*rsa.PublicKey, error) {
	pub, err := os.ReadFile(publicKeyFile)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(pub)
	if block != nil {
		pub = block.Bytes
	}
	key, err := x509.ParsePKCS1PublicKey(pub)
	if err != nil {
		return nil, err
	}
	return key, nil
}
