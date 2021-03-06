package database

import (
	"os"
	"runtime"
	"sync"

	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/options"
	"github.com/pkg/errors"
)

var (
	defaultBadger     *badger.DB
	defaultBadgerInit sync.Once
)

// Returns whether the given file or directory exists.
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func checkDir(dir string) error {
	exists, err := exists(dir)
	if err != nil {
		return err
	}

	if !exists {
		return os.Mkdir(dir, 0700)
	}
	return nil
}

func CreateDB(directory string, optionalOptions ...badger.Options) (*badger.DB, error) {
	if err := checkDir(directory); err != nil {
		return nil, errors.Wrap(err, "Could not check directory")
	}

	var opts badger.Options

	if len(optionalOptions) > 0 {
		opts = optionalOptions[0]
	} else {
		opts = badger.DefaultOptions(directory)
		opts.Logger = nil

		opts.LevelOneSize = 256 << 20
		opts.LevelSizeMultiplier = 10
		opts.TableLoadingMode = options.MemoryMap
		opts.ValueLogLoadingMode = options.MemoryMap

		opts.MaxLevels = 7
		opts.MaxTableSize = 64 << 20
		opts.NumCompactors = 2 // Compactions can be expensive. Only run 2.
		opts.NumLevelZeroTables = 5
		opts.NumLevelZeroTablesStall = 10
		opts.NumMemtables = 5
		opts.SyncWrites = true
		opts.NumVersionsToKeep = 1
		opts.CompactL0OnClose = true

		opts.ValueLogFileSize = 1<<30 - 1

		opts.ValueLogMaxEntries = 1000000
		opts.ValueThreshold = 32
		opts.Truncate = false
		opts.LogRotatesToFlush = 2

		if runtime.GOOS == "windows" {
			opts = opts.WithTruncate(true)
		}
	}

	db, err := badger.Open(opts)
	if err != nil {
		return nil, errors.Wrap(err, "Could not open new DB")
	}

	return db, nil
}

func GetBadgerInstance(optionalDirectory ...string) *badger.DB {
	defaultBadgerInit.Do(func() {

		directory := "mainnetdb"
		if len(optionalDirectory) > 0 {
			directory = optionalDirectory[0]
		}

		db, err := CreateDB(directory)
		if err != nil {
			// errors should cause a panic to avoid singleton deadlocks
			panic(err)
		}
		defaultBadger = db
	})
	return defaultBadger
}
