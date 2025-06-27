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

const (
	appVersion = "v1.1.0"
)

var (
	version       = flag.Bool("version", false, "print tool version")
	encrypt       = flag.Bool("encrypt", false, "encrypt the given source")
	publicKeyPath = flag.String("public-key", "./public.pem", "public key")
)

func main() {
	flag.Parse()

	if *version {
		fmt.Println("Version:", appVersion)
		return
	}

	if len(flag.Args()) == 0 {
		fmt.Println("missing source zip file or directory")
		return
	}

	if len(flag.Args()) > 1 {
		fmt.Println("only 1 zip file or directory repacked")
		return
	}

	sourceDir := flag.Args()[0]
	sourceDir = path.Clean(sourceDir)

	dstFilename := fmt.Sprintf("%s_repacked", strings.TrimSuffix(filepath.Base(sourceDir), ".zip"))
	if *encrypt {
		dstFilename += ".enc"
	} else {
		dstFilename += ".bin"
	}

	fw, err := os.Create(dstFilename)
	if err != nil {
		fmt.Printf("error on creating a file:%s, error:%s\n", dstFilename, err.Error())
		return
	}
	defer fw.Close()

	stream := estream.NewWriter(fw)
	addStream := stream.AddUnencryptedStream
	if *encrypt {
		// load public certificate
		publicKeyFile := path.Clean(*publicKeyPath)
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
		addStream = stream.AddEncryptedStream
	}

	if strings.HasSuffix(sourceDir, ".zip") { // process zip archive
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

			f, err := addStream(filepath.Base(file.Name), nil)
			if err != nil {
				fmt.Printf("error on adding a file to stream:%s, error:%s\n", file.Name, err.Error())
				return
			}

			_, err = io.Copy(f, fileReader)
			if err != nil {
				fmt.Printf("error on adding a data to stream:%s, error:%s\n", file.Name, err.Error())
				return
			}
			if err = f.Close(); err != nil {
				fmt.Printf("error on closing a stream:%s, error:%s\n", file.Name, err.Error())
				return
			}
		}
	} else { // process files from a directory
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
			f, err := addStream(file.Name(), nil)
			if err != nil {
				fmt.Printf("error on adding a file to stream:%s, error:%s\n", _filename, err.Error())
				return
			}
			_, err = f.Write(data)
			if err != nil {
				fmt.Printf("error on adding a data to stream:%s, error:%s\n", _filename, err.Error())
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
