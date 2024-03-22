package device

import (
	"os"
	"path"
	"strings"
)

type Directory struct {
	Name string
	Path string
}

func GetDirectories(rootPath, filterPrefix string) ([]Directory, error) {

	var subDirs []Directory
	files, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			if strings.HasPrefix(file.Name(), filterPrefix) {
				subDirs = append(subDirs, Directory{
					Name: file.Name(),
					Path: path.Join(rootPath, file.Name()),
				})
			}
		}
	}

	return subDirs, nil
}

func GetDirectoriesToMap(rootPath, filterPrefix string) (map[string]Directory, error) {
	dirs, err := GetDirectories(rootPath, filterPrefix)
	if err != nil {
		return nil, err
	}

	dirsMap := make(map[string]Directory)
	for _, dir := range dirs {
		dirsMap[dir.Name] = dir
	}

	return dirsMap, nil
}
