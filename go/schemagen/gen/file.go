package gen

import (
	"io/fs"
	"os"
	"strings"
)

type SourceFile struct {
	Name      string
	Extension string
	Path      []string
}

func NewSourceFile(path string, ext string) *SourceFile {
	p := strings.Split(path, "/")

	f := &SourceFile{
		Extension: ext,
	}
	f.Name, _ = strings.CutSuffix(p[len(p)-1], ext)
	for i := 0; i < len(p)-1; i++ {
		f.Path = append(f.Path, p[i])
	}

	return f
}

func readFiles(dir string, ext string) ([]*SourceFile, error) {
	fileSystem := os.DirFS(dir)

	files := make([]*SourceFile, 0)

	err := fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}
		if !strings.HasSuffix(path, ext) {
			return nil
		}

		f := NewSourceFile(path, ext)
		files = append(files, f)

		return nil
	})

	return files, err
}

// SourcePath формирует строку пути к source file
func (f *SourceFile) SourcePath() string {
	return strings.Join(f.Path, "/") + "/" + f.Name + f.Extension
}

// DestPath формирует строку пути к файлу с результатом генерации
func (f *SourceFile) DestPath(ext string) string {
	return strings.Join(f.Path, "/") + "/" + f.Name + ext
}
