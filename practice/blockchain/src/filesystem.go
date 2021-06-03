package database

import (
	"os"	"path/filepath"
)


func getDatabaseDirPath(dataDir string) string {
	return filepath.Join(dataDir, "database")
}

func getGenesisJsonFilePath(dataDir string) string {
	return filepath.Join(getDatabaseDirPath(dataDir), "genesis.json")
}

func getBlocksDbFilePath(dataDir string) string {
	return filepath.Join(getDatabaseDirPath(dataDir), "blocks.db")
}

func initDataDirIfNotExists(dataDir string) error {
	if fileExists(dataDir) {
		return nil
	}

	// need a database directory
	dbDir := getDatabaseDirPath(dataDir)
	if err := os.MkdirAll(dbDir); err != nil {
		return err
	}
	// need to write an empty genesis file to that directory
	gen = getGenesisJsonFilePath(dbDir)
	if err := writeGenesisToDisk(gen); err != nil {
		return err
	}
	// need to write an empty blocks db file to that directory
	blocks := getBlocksDbFilePath(dataDir)
	if err := writeEmptyBlocksDbToDisk(blocks); err != nil {
		return err
	}

	return nil
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}

func dirExists(dir string) (bool, error) {
	_, err := os.Stat(dir)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}
