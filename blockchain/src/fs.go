package database

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// group all file system related business logic for creating a new datadir to stoer the database files
func getDatabaseDirPath(dataDir string) string {
	return filepath.Join(dataDir, "database")
}

func getGenesisJsonFilePath(dataDir string) string {
	return filepath.Join(getDatabaseDirPath(dataDir), "genesis.json")
}

func getBlocksDbFilePath(dataDir string) string {
	return filepath.Join(getDatabaseDirPath(dataDir), "blocks.db")
}

func getStateJsonFilePath(dataDir string) string {
	return filepath.Join(getDatabaseDirPath(dataDir), "state.json")
}

func initDataDirIfNotExists(dataDir string) error {
	if fileExist(getGenesisJsonFilePath(dataDir)) {
		return nil
	}

	dbDir := getDatabaseDirPath(dataDir)
	if err := os.MkdirAll(dbDir, os.ModePerm); err != nil {
		return err
	}

	gen := getGenesisJsonFilePath(dataDir)
	if err := writeGenesisToDisk(gen); err != nil {
		return err
	}

	blocks := getBlocksDbFilePath(dataDir)
	if err := writeEmptyBlocksDbToDisk(blocks); err != nil {
		return err
	}

	return nil
}

func fileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}

func dirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}

func writeEmptyBlocksDbToDisk(path string) error {
	return ioutil.WriteFile(path, []byte(""), os.ModePerm)
}
