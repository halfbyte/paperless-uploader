package main

import (
	"bytes"
	"encoding/base64"
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

func uploadFile(filePath, url, username, password string) error {
	if _, err := os.Stat(filePath); err != nil {
		return err
	}

	filename := filepath.Base(filePath)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	fileWriter, err := writer.CreateFormFile("document", filename)
	if err != nil {
		return err
	}

	fileData, _ := os.ReadFile(filePath)

	io.Copy(fileWriter, bytes.NewReader(fileData))
	writer.Close()

	postReq, err := http.NewRequest("POST", url+"/api/documents/post_document/", body)
	if err != nil {
		return err
	}
	postReq.Header.Add("Accept", "application/json; version=2")
	postReq.Header.Add("Authorization", "Basic "+encodeCredentials(username, password))

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

func uploadFiles(filePaths []string) {
	config, err := readConfig()

	if err != nil {
		panic(err)
	}

	files := len(filePaths)
	if files == 0 {
		panic("No Files given")
	}
	for i := 0; i < files; i++ {
		filePath := filePaths[i]
		err := uploadFile(filePath, config.url, config.username, config.password)
		if err != nil {
			fmt.Println("ERROR", err.Error())
		}
	}
}

func main() {
	var login = flag.Bool("login", false, "Provide credentials to Paperless")

	flag.Parse()

	if *login {
		loginAndSaveConfig()
	} else {
		uploadFiles(flag.Args())
	}
}
