package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/term"
	"gopkg.in/ini.v1"
)

type Config struct {
	url, username, password string
}

type PaperlessTag struct {
	Id   uint64
	Name string
}

type PaperlessTags struct {
	Results []PaperlessTag
}

func encodeCredentials(username, password string) string {
	data := []byte(username + ":" + password)
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(dst, data)
	return string(dst)
}

func readConfig() (Config, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return Config{}, err
	}

	iniFilePath := filepath.Join(userConfigDir, "paperless-uploader.ini")

	file, err := ini.Load(iniFilePath)
	if err != nil {
		return Config{}, err
	}

	url := file.Section("server").Key("url").String()
	username := file.Section("server").Key("username").String()
	password := file.Section("server").Key("password").String()
	return Config{url, username, password}, nil
}

func writeConfig(config Config) (Config, error) {
	var file *ini.File
	oldConfig, _ := readConfig()
	oldConfig.url = config.url
	oldConfig.username = config.username
	oldConfig.password = config.password

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		panic(err) // no good way to recover from this
	}

	iniFilePath := filepath.Join(userConfigDir, "paperless-uploader.ini")
	file, err = ini.Load(iniFilePath)
	if err != nil {
		file = ini.Empty()
	}

	serverSection := file.Section("server")

	serverSection.Key("url").SetValue(oldConfig.url)
	serverSection.Key("username").SetValue(oldConfig.username)
	serverSection.Key("password").SetValue(oldConfig.password)

	err = file.SaveTo(iniFilePath)
	if err != nil {
		panic(err)
	}

	fmt.Println("Config successfully written to disk.")

	return oldConfig, nil
}

func allTags(config Config) ([]PaperlessTag, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", config.url+"/api/tags/", nil)
	if err != nil {
		return []PaperlessTag{}, err
	}

	req.Header.Add("Accept", "application/json; version=2")
	req.Header.Add("Authorization", "Basic "+encodeCredentials(config.username, config.password))

	resp, err := client.Do(req)
	if err != nil {
		return []PaperlessTag{}, err
	}
	if resp.StatusCode != 200 {
		return []PaperlessTag{}, fmt.Errorf("api_tags_get_unsuccessful")
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	var tags PaperlessTags
	err = json.Unmarshal(respBody, &tags)
	if err != nil {
		return []PaperlessTag{}, err
	}
	return tags.Results, nil
}

func findTag(tags []PaperlessTag, tag string) uint64 {
	for _, v := range tags {
		if v.Name == tag {
			return v.Id
		}
	}
	return 0
}

func createTag(config Config, tagName string) (uint64, error) {
	client := &http.Client{}

	body := new(bytes.Buffer)
	body.WriteString("name=" + tagName)

	req, err := http.NewRequest("POST", config.url+"/api/tags/", body)
	if err != nil {
		return 0, err
	}

	req.Header.Add("Accept", "application/json; version=2")
	req.Header.Add("Authorization", "Basic "+encodeCredentials(config.username, config.password))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 201 {
		fmt.Println(string(respBody))
		return 0, fmt.Errorf("api_tags_post_unsuccessful")
	}
	var tag PaperlessTag
	err = json.Unmarshal(respBody, &tag)
	if err != nil {
		return 0, err
	}
	return tag.Id, nil
}

func ensureTag(config Config, tagName string) (uint64, error) {
	var tagId uint64
	tags, err := allTags(config)
	if err != nil {
		fmt.Println("Error reading tags")
		return 0, err
	}
	tagId = findTag(tags, tagName)
	if tagId == 0 {
		return createTag(config, tagName)
	}
	return tagId, nil
}

func uploadFile(filePath string, config Config, tagId uint64) error {
	if _, err := os.Stat(filePath); err != nil {
		return err
	}

	filename := filepath.Base(filePath)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	if tagId != 0 {
		fieldWriter, err := writer.CreateFormField("tags")
		if err != nil {
			return err
		}
		fieldWriter.Write([]byte(fmt.Sprint(tagId)))
	}

	fileWriter, err := writer.CreateFormFile("document", filename)
	if err != nil {
		return err
	}

	fileData, _ := os.ReadFile(filePath)

	io.Copy(fileWriter, bytes.NewReader(fileData))
	writer.Close()

	postReq, err := http.NewRequest("POST", config.url+"/api/documents/post_document/", body)
	if err != nil {
		return err
	}
	postReq.Header.Add("Accept", "application/json; version=2")
	postReq.Header.Add("Authorization", "Basic "+encodeCredentials(config.username, config.password))

	postReq.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Do(postReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		fmt.Printf("Status: %s\n", resp.Status)
		fmt.Println(string(respBody))
		return fmt.Errorf("post error")
	}
	fmt.Printf("File %s successfully uploaded\n", filename)
	return nil
}

func testAPI(config Config) error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", config.url+"/api/", nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json; version=2")
	req.Header.Add("Authorization", "Basic "+encodeCredentials(config.username, config.password))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("api_test_unsuccessful")
	}
	fmt.Println("Config successfully tested, writing to disk")
	return nil
}

func loginAndSaveConfig() {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		panic("Not a terminal")
	}
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	screen := struct {
		io.Reader
		io.Writer
	}{os.Stdin, os.Stdout}

	terminal := term.NewTerminal(screen, "")

	terminal.Write([]byte("Server URL: "))
	url, err := terminal.ReadLine()
	if err != nil {
		panic(err)
	}

	terminal.Write([]byte("Username: "))
	username, err := terminal.ReadLine()
	if err != nil {
		panic(err)
	}

	password, err := terminal.ReadPassword("Password: ")
	if err != nil {
		panic(err)
	}

	config := Config{url, username, password}
	if err := testAPI(config); err == nil {
		writeConfig(config)
	} else {
		panic(err)
	}
}

func uploadFiles(filePaths []string, tag *string) {
	config, err := readConfig()

	if err != nil {
		panic(err)
	}

	tagId, err := ensureTag(config, *tag)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found/Made Tag %s with id %d\n", *tag, tagId)

	files := len(filePaths)
	if files == 0 {
		panic("No Files given")
	}
	for i := 0; i < files; i++ {
		filePath := filePaths[i]
		err := uploadFile(filePath, config, tagId)
		if err != nil {
			fmt.Println("ERROR", err.Error())
		}
	}
}

func main() {
	var login = flag.Bool("login", false, "Provide credentials to Paperless")
	var tag = flag.String("tag", "", "A tag to set on the uploaded file")

	flag.Parse()

	if *login {
		loginAndSaveConfig()
	} else {
		uploadFiles(flag.Args(), tag)
	}
}
