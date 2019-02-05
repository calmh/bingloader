package main

import (
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	outDir := "."
	resolution := "1920x1200"
	setBG := false
	flag.StringVar(&outDir, "dir", outDir, "Output directory")
	flag.StringVar(&resolution, "res", resolution, "Resolution")
	flag.BoolVar(&setBG, "bg", setBG, "Set as desktop background")
	flag.Parse()

	base, err := urlBase()
	if err != nil {
		fmt.Println("Listing images:", err)
		os.Exit(1)
	}

	url := fmt.Sprintf("http://www.bing.com%s_%s.jpg", base, resolution)
	resp, err := http.Get(url)
	if err == nil && resp.StatusCode > 299 {
		err = errors.New(resp.Status)
	}
	if err != nil {
		fmt.Println("Loading image:", err)
		os.Exit(1)
	}

	outFile := filepath.Join(outDir, path.Base(url))
	if _, err := os.Stat(outFile); err != nil {
		fd, err := os.Create(outFile)
		if err != nil {
			fmt.Println("Outfile:", err)
			os.Exit(1)
		}

		if _, err := io.Copy(fd, resp.Body); err != nil {
			fmt.Println("Copy:", err)
			os.Exit(1)
		}

		if err := fd.Close(); err != nil {
			fmt.Println("Close:", err)
			os.Exit(1)
		}

		if setBG {
			if err := setBackground(outFile); err != nil {
				fmt.Println("Setting background:", err)
				os.Exit(1)
			}
		}
	}

	if !setBG {
		fmt.Println(outFile)
	}
}

type bingXML struct {
	Image []struct {
		URLBase string `xml:"urlBase"`
	} `xml:"image"`
}

func urlBase() (string, error) {
	resp, err := http.Get("http://www.bing.com/HPImageArchive.aspx?format=xml&idx=0&n=1&mkt=en_UK")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	dec := xml.NewDecoder(resp.Body)
	var result bingXML
	if err := dec.Decode(&result); err != nil {
		return "", err
	}

	if len(result.Image) < 1 {
		return "", errors.New("no image")
	}

	return result.Image[0].URLBase, nil
}

func setBackground(image string) error {
	script := fmt.Sprintf(`tell application "Finder"
set desktop picture to POSIX file "%s"
end tell
`, image)
	cmd := exec.Command("/usr/bin/osascript")
	cmd.Stdin = strings.NewReader(script)
	return cmd.Run()
}
