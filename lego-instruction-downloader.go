package main

import (
	"flag"
	"github.com/beevik/etree"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var id = flag.String("id", "0", "The set to find instructions for")

type conf struct {
	ApiKey       string `yaml:"api_key"`
	ApiUrl       string `yaml:"api_url"`
	Papersize    string
	Password     string
	Username     string
	DownloadPath string `yaml:"download_path"`
	token        string
}

type userHash struct {
	String string
}

func login(c conf) string {
	log.Println("Logging into brickset")
	formData := url.Values{
		"apiKey":   {c.ApiKey},
		"username": {c.Username},
		"password": {c.Password},
	}
	resp, err := http.PostForm(c.ApiUrl+"/login", formData)
	if err != nil {
		log.Fatal(err)
	}
	doc := etree.NewDocument()
	if _, err := doc.ReadFrom(resp.Body); err != nil {
		log.Fatal(err)
	}
	s := doc.SelectElement("string")
	return s.Text()
}

func get_sets(c conf, id string) map[string]string {
	log.Printf("Getting sets which match %v", id)
	formData := url.Values{
		"apiKey":     {c.ApiKey},
		"query":      {id},
		"userHash":   {c.token},
		"theme":      {""},
		"subtheme":   {""},
		"setNumber":  {""},
		"year":       {""},
		"owned":      {""},
		"wanted":     {""},
		"orderBy":    {""},
		"pageSize":   {"20"},
		"pageNumber": {"1"},
		"userName":   {""},
	}
	resp, err := http.PostForm(c.ApiUrl+"/getSets", formData)
	if err != nil {
		log.Fatal(err)
	}
	doc := etree.NewDocument()
	if _, err := doc.ReadFrom(resp.Body); err != nil {
		log.Fatal(err)
	}
	setids := doc.FindElements("//sets/setID")
	numbers := doc.FindElements("//sets/number")

	result := make(map[string]string)
	for index, setid := range setids {
		number := numbers[index].Text()
		numberParts := strings.Split(number, "-")
		if numberParts[0] == id {
			result[number] = setid.Text()
		} else {
			log.Printf("setId %v with number %v doesn't match searched %v", setid.Text(), number, id)
		}
	}

	return result
}

func save_instructions(c conf, number string, setId string) {
	log.Printf("Saving Instructions for %v", setId)

	formData := url.Values{
		"apiKey": {c.ApiKey},
		"setID":  {setId},
	}
	resp, err := http.PostForm(c.ApiUrl+"/getInstructions", formData)
	if err != nil {
		log.Fatal(err)
	}
	doc := etree.NewDocument()
	if _, err := doc.ReadFrom(resp.Body); err != nil {
		log.Fatal(err)
	}
	urls := doc.FindElements("//instructions/URL")
	descriptions := doc.FindElements("//instructions/description")
	for index, url := range urls {
		desc := descriptions[index].Text()
		if strings.Contains(desc, c.Papersize) {
			desc := strings.Replace(desc, "/", "_", -1)
			desc = strings.Replace(desc, " ", "_", -1)
			log.Printf("Found instructions '''%v'''", desc)
			dirPath := filepath.Join(c.DownloadPath, number)
			fileName := desc + ".pdf"
			download(url.Text(), dirPath, fileName)
		} else {
			log.Printf("Description '''%v''' does not match papersize", desc)
		}
	}

}

func download(url string, dirPath string, fileName string) {
	log.Printf("Download %v to %v %v", url, dirPath, fileName)
	// if output dir doesn't exist, create it
	os.MkdirAll(dirPath, os.ModePerm)

	outputFile := filepath.Join(dirPath, fileName)

	// open URL
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to download %v", url)
		return
	}
	// write bytes to output file
	file, err := os.Create(outputFile)
	if err != nil {
		log.Printf("Failed to create %v", outputFile)
		return
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		log.Printf("Failed to copy %v to %v", url, outputFile)
	}
}

func main() {
	yamlFile, err := ioutil.ReadFile("/home/lbedford/.lego-instructions.yml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	var c conf
	if err := yaml.UnmarshalStrict(yamlFile, &c); err != nil {
		log.Fatal(err)
	}
	flag.Parse()
	if *id == "0" {
		log.Fatal("Need an ID to continue. --help")
	}
	token := login(c)
	c.token = token
	setIds := get_sets(c, *id)
	for number, setId := range setIds {
		save_instructions(c, number, setId)
	}
}
