/*

Monitors eztv feed and downloads torrent files that match a list of shows specified in a file.

Takes command line arguments which can be seen by running in the command with the -? flag

*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/ungerik/go-rss"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

const (
	feedUrl         = "http://www.ezrss.it/feed/"
	downloadedFiles = "downloaded.txt"

	// Defaults
	defaultSleepMinutes = 60
	defaultShowsFile    = "shows.txt"
)

func getLines(fileName string) ([]string, error) {
	lines := make([]string, 0)

	f, err := os.Open(fileName)
	if err != nil {
		return lines, err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	var lineBuffer []byte
	for {
		bytes, isPrefix, err := reader.ReadLine()
		if err != nil {
			// No more lines
			break
		}

		lineBuffer = append(lineBuffer, bytes...)

		if !isPrefix {
			line := string(lineBuffer)
			if len(line) > 0 {
				lines = append(lines, line)
				lineBuffer = make([]byte, 0)
			}
		}
	}

	return lines, nil
}

func titleInShowList(title string, shows []string) bool {
	for _, show := range shows {
		// (?i) makes the regex case insensitive
		if matched, err := regexp.MatchString("(?i)"+show, title); matched && err == nil {
			return true
		}
	}
	return false
}

func alreadyDownloaded(title string) bool {
	titles, err := getLines(downloadedFiles)
	if err != nil {
		return false
	}

	for _, downloadedTitle := range titles {
		if title == downloadedTitle {
			return true
		}
	}

	return false
}

func tryDownload(item rss.Item, downloadPath string) {
	if alreadyDownloaded(item.Title) {
		log.Println("Already downloaded")
		return
	}

	fmt.Println("Downloading")
	res, err := http.Get(item.Link)
	if err != nil {
		log.Printf("Error downloading torrent: %v\n", err)
		return
	}

	fileName := fmt.Sprintf("%v.torrent", item.Title)
	defer res.Body.Close()
	torrentData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error getting torrent data: %v\n", err)
		return
	}
	fmt.Println("Downloaded")

	filePath := filepath.Join(downloadPath, fileName)
	err = ioutil.WriteFile(filePath, torrentData, 0666)
	if err != nil {
		log.Printf("Error writing file: %v\n", err)
		return
	}

	file, err := os.OpenFile(downloadedFiles, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		log.Println("Error opening download file for writing: %v\n", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("%v\n", item.Title))
	if err != nil {
		log.Println("Error writing to file: %v\n", err)
	}
}

func main() {
	// Parse flags
	var showsFileName string
	flag.StringVar(&showsFileName, "showsfile", defaultShowsFile, "The name of the file containing the shows")
	var sleepMinutes int
	flag.IntVar(&sleepMinutes, "sleep", defaultSleepMinutes, "The number of minutes to sleep between checks")
	var downloadPath string
	flag.StringVar(&downloadPath, "downloadPath", "", "The path where torrents files are downloaded to")
	flag.Parse()

	c := time.Tick(time.Duration(sleepMinutes) * time.Minute)

	for ; ; <-c {
		log.Println("Getting feed...")

		// Get shows from file
		shows, err := getLines(showsFileName)
		if err != nil {
			log.Println(err)
			continue
		}

		// Download feed
		rssChannel, err := rss.Read(feedUrl)
		if err != nil {
			log.Println(err)
			continue
		}

		for _, item := range rssChannel.Item {
			if titleInShowList(item.Title, shows) {
				log.Printf("Show matches: %v\n", item.Title)
				tryDownload(item, downloadPath)
			}
		}

		log.Println("Waiting...")
	}
}
