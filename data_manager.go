package main

import (
	"io/ioutil"
)

// readPath читает файл, где записан путь до файлов с данными
func readPath() (string, error) {
	rawData, err := ioutil.ReadFile("path.txt")
	if err != nil {
		return "", err
	}

	return string(rawData), nil
}

// readJSON читает json файл
func readJSON(path string) ([]byte, error) {
	rawData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return rawData, nil
}

// writeJSON записывает данные в json файл
func writeJSON(path string, dataIn []byte) error {
	err := ioutil.WriteFile(path, dataIn, 0)
	if err != nil {
		return err
	}

	return nil
}
