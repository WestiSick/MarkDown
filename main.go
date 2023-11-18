package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func init() {
	// загружаем значения из файла .env
	if err := godotenv.Load(); err != nil {
		log.Print("Файл .env не найден")
	}
}

func main() {
	http.HandleFunc("/", fileListHandler)
	http.HandleFunc("/file", fileHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func fileListHandler(w http.ResponseWriter, r *http.Request) {
	owner := os.Getenv("OWNER")
	repo := os.Getenv("REPO")
	branch := os.Getenv("BRANCH")

	fileURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/%s?recursive=1", owner, repo, branch)

	resp, err := http.Get(fileURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "<html><body>")
	if tree, ok := result["tree"].([]interface{}); ok {
		for _, file := range tree {
			if fileInfo, ok := file.(map[string]interface{}); ok {
				path, pathOk := fileInfo["path"].(string)
				fileType, typeOk := fileInfo["type"].(string)
				if pathOk && typeOk && fileType == "blob" && len(path) > 3 && path[len(path)-3:] == ".md" {
					fmt.Fprintf(w, "<a href=\"/file?owner=%s&repo=%s&branch=%s&path=%s\">%s</a><br>", owner, repo, branch, path, path)
				}
			}
		}
	}
	fmt.Fprintln(w, "</body></html>")
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
	owner := r.URL.Query().Get("owner")
	repo := r.URL.Query().Get("repo")
	branch := r.URL.Query().Get("branch")
	path := r.URL.Query().Get("path")

	fileContent, err := getFileContent(owner, repo, branch, path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	htmlContent := blackfriday.MarkdownCommon([]byte(fileContent))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(htmlContent)
}

func getFileContent(owner, repo, branch, path string) (string, error) {
	fileURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, repo, branch, path)

	resp, err := http.Get(fileURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("запрос к %s не удался: %s", fileURL, resp.Status)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
