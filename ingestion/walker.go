package ingestion

import (
	"os"
	"path/filepath"
	"strings"
)

var languageMap = map[string]string{
	".go": "golang",
	".py": "python",
	".cpp":"cpp",
	".js": "javascript",
	".ts": "typescript",
	".jsx": "javascript",
	".tsx": "typescript",
	".html": "html",
	".css": "css",
	".scss": "scss",
	".sass": "sass",
	".less": "less",
	".md": "markdown",
	".json": "json",
	".yaml": "yaml",
	".yml": "yaml",
	".toml": "toml",
	".xml": "xml",
	".php": "php",
	".rb": "ruby",
	".java": "java",
	".rs":"rust",
	".c": "c",
	".sh":"bash",
}

//remove noisy dirs 
var skipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,  
	".venv":        true,
	"__pycache__":  true,
	"dist":         true,
	"build":        true,
	".idea":        true,  
	".vscode":      true,
}

var skipFiles = map[string]bool{
	"go.sum":           true,
	"package-lock.json": true,
	"yarn.lock":        true,
	"poetry.lock":      true,
}

type FileInfo struct{
	AbsPath string
	RelPath string
	Language string
}

func WalkRepo(repoPath string)([]FileInfo , error){
	var files []FileInfo

	err := filepath.Walk(repoPath , func(path string , info os.FileInfo , err error)error{
		if err != nil {
			return err
		}

		if info.IsDir(){
			if skipDirs[info.Name()]{
				return filepath.SkipDir
			}
			return nil
		}

		if skipFiles[info.Name()]{
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		lang , ok := languageMap[ext]
		if !ok{
			return nil
		}

		if info.Size() > 500 *1024{
			return nil
		}

		relPath , _ := filepath.Rel(repoPath ,path)

		files = append(files , FileInfo{
			AbsPath: path,
			RelPath: relPath,
			Language: lang,
		})

		return nil
	})

	return files , err
}