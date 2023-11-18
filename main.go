package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type GitHubTreeResponse struct {
	Tree []struct {
		Path string `json:"path"`
		Type string `json:"type"`
	} `json:"tree"`
}

func init() {
	// загружаем значения из файла .env
	if err := godotenv.Load(); err != nil {
		log.Print("Файл .env не найден")
	}
}

func main() {
	http.HandleFunc("/", fileListHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func fileListHandler(w http.ResponseWriter, r *http.Request) {
	owner := os.Getenv("OWNER")
	repo := os.Getenv("REPO")
	branch := os.Getenv("BRANCH")

	token := os.Getenv("TOKEN")

	fileURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/%s?recursive=1", owner, repo, branch)
	req, err := http.NewRequest("GET", fileURL, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "token "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Запрос к API GitHub не удался: "+resp.Status, resp.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var tree GitHubTreeResponse
	err = json.Unmarshal(body, &tree)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var markdownFiles []string
	for _, file := range tree.Tree {
		if file.Type == "blob" && len(file.Path) >= 3 && file.Path[len(file.Path)-3:] == ".md" {
			markdownFiles = append(markdownFiles, file.Path)
		}
	}

	for _, file := range markdownFiles {
		fmt.Fprintln(w, file)
	}

	log.Println("Запрос успешно обработан")
}
