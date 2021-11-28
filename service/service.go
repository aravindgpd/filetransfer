package service

import (
	"io/fs"
	"io/ioutil"
	"log"
	"os"
)

func GetFolderLocation() string {
	folderLocation := "/tmp"
	if os.Getenv("STORAGE_LOCATION") != "" {
		folderLocation = os.Getenv("STORAGE_LOCATION")
	}
	return folderLocation
}

func ReadDirectory(filename string) ([]fs.FileInfo, error) {
	file, err := ioutil.ReadDir(filename)
	return file, err
}

func CheckFolderExists(folderLocation string) error {
	_, err := os.Stat(folderLocation)
	if os.IsNotExist(err) {
		err = os.Mkdir(folderLocation, 0644)
		if err != nil {
			log.Printf("cannot create folder %v", err)
			return err
		}
	}
	return err
}

func CheckFileExists(folderLocation string, fileName string) error {
	_, err := os.Stat(folderLocation + "/" + fileName)
	if os.IsNotExist(err) {
		_, err = os.Create(folderLocation + "/" + fileName)
		if err != nil {
			log.Printf("cannot create file %v", err)
			return err
		}
	}
	return err
}
