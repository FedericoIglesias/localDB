package local_db

import (
	"encoding/json"
	"fmt"

	"os"
	"path/filepath"
	"sync"

	"github.com/jcelliott/lumber"
)

const version = "0.0.4"

type (
	Logger interface {
		Fatal(string, ...interface{})
		Error(string, ...interface{})
		Warn(string, ...interface{})
		Info(string, ...interface{})
		Debug(string, ...interface{})
		Trace(string, ...interface{})
	}

	Driver struct {
		mutex   sync.Mutex
		mutexes map[string]*sync.Mutex
		dir     string
		log     Logger
	}
)

type Options struct {
	Logger
}

func New(dir string, options *Options) (*Driver, error) {

	fmt.Printf("DB Versión %s \n ", version)

	dir = filepath.Clean(dir)

	opts := Options{}

	if options != nil {
		opts = *options
	}

	if opts.Logger == nil {
		opts.Logger = lumber.NewConsoleLogger(lumber.INFO)
	}

	drive := Driver{
		dir:     dir,
		mutexes: make(map[string]*sync.Mutex),
		log:     opts.Logger,
	}

	if _, err := os.Stat(dir); err == nil {
		opts.Logger.Debug("Using '%s' (database already exists)\n", dir)
		return &drive, nil
	}

	opts.Logger.Debug("Creating the database at '%s'...\n", dir)
	return &drive, os.MkdirAll(dir, 0775)

}

func (d *Driver) Write(collection, resource string, v interface{}) error {
	if collection == "" {
		return fmt.Errorf("missing collection - no place to save record")
	}

	if resource == "" {
		return fmt.Errorf("missing resource - unable to save record (no name)")
	}

	mutex := d.getOrCreatMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.dir, collection)
	fnlPath := filepath.Join(dir, resource+".json")
	tmpPath := fnlPath + ".tmp"

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}

	b = append(b, byte('\n'))

	if err := os.WriteFile(tmpPath, b, 0644); err != nil {
		return err
	}

	return os.Rename(tmpPath, fnlPath)
}

func (d *Driver) Read(collection, resources string, v interface{}) error {

	if collection == "" {
		return fmt.Errorf("missing collection - no place to save record")
	}

	if resources == "" {
		return fmt.Errorf("missing resource - unable to save record (no name)")
	}

	record := filepath.Join(d.dir, collection, resources)

	if _, err := stat(record); err != nil {
		return err
	}

	b, err := os.ReadFile(record + ".json")

	if err != nil {
		return err
	}

	return json.Unmarshal(b, &v)

}

func (d *Driver) ReadAll(collection string) ([]string, error) {

	if collection == "" {
		return nil, fmt.Errorf("missing collection - unable to read")
	}

	dir := filepath.Join(d.dir, collection)

	if _, err := stat(dir); err != nil {
		return nil, err
	}

	files, _ := os.ReadDir(dir)

	var records []string

	for _, file := range files {
		b, err := os.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}
		records = append(records, string(b))
	}

	return records, nil

}

func (d *Driver) Delete(collection, resource string) error {

	path := filepath.Join(collection, resource)

	mutex := d.getOrCreatMutex(collection)

	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.dir, path)

	switch fi, err := stat(dir); {
	case fi == nil, err != nil:
		return fmt.Errorf("unable to find file or directory named %v\n ", path)
	case fi.Mode().IsDir():
		return os.RemoveAll(dir)
	case fi.Mode().IsRegular():
		return os.RemoveAll((dir + ".json"))
	}

	return nil
}

func (d *Driver) getOrCreatMutex(collection string) *sync.Mutex {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	m, ok := d.mutexes[collection]

	if !ok {
		m = &sync.Mutex{}
		d.mutexes[collection] = m
	}
	return m

}

func stat(path string) (fi os.FileInfo, err error) {
	if fi, err = os.Stat(path); os.IsNotExist(err) {
		fi, err = os.Stat(path + ".json")
	}
	return
}