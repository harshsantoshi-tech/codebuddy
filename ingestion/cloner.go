package ingestion

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)


func CloneRepo(repoURL string) (string , error){

	repoName := ExtractRepoName(repoURL)

	//clone into /temp/codebuddy/repoName
	clonePath := filepath.Join(os.TempDir(),"codebuddy", repoName)

	if _ , err := os.Stat(clonePath); err == nil {
		fmt.Println("Repo already cloned at : ",clonePath)
		return clonePath, nil
	}

	if err := os.MkdirAll(filepath.Dir(clonePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	fmt.Println("Cloning repo : ",repoURL)

	//shallow copy (--depth=1) ,although no git history in this way
	cmd := exec.Command("git", "clone", "--depth=1", repoURL, clonePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to clone repo: %v", err)
	}

	fmt.Println("Cloned to : ",clonePath)
	return clonePath, nil
}

func ExtractRepoName(repoURL string) string {
	repoURL = strings.TrimRight(repoURL, "/")

	parts := strings.Split(repoURL, "/")
	name := parts[len(parts)-1]
	
	return strings.TrimSuffix(name, ".git")
}