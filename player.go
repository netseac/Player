package player

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dhowden/tag"
	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"
)

type playlist struct {
	name  string
	Files []File
}

type Metadata struct {
	Title  string
	Artist string
}

type File struct {
	Metadata
	Path string
}

func Run() {
	var path string
	flag.StringVar(&path, "path", "", "set path for playing")
	flag.Parse()

	if path == "" {
		fmt.Println("Please set path.")
		os.Exit(1)
	}

	fmt.Printf("Path %v", path)

	_, err := setData(path)
	if err != nil {
		panic(err)
	}
	data, err := getData()
	if err != nil {
		panic(err)
	}

	for _, d := range data {
		fmt.Printf("Current playing: %v - %v \n", d.Artist, d.Title)
		play(d.Path)
	}
}

func setData(root string) ([]File, error) {
	var f []File
	fmt.Printf("Current root %v", root)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		fmt.Printf("Current path %v", path)
		if path == root {
			fmt.Printf("Path %v is root path", path)
			return nil
		}
		if info.IsDir() {
			fmt.Printf("Path %v is dir", path)
			return nil
		}
		if filepath.Ext(path) != ".mp3" {
			return nil
		}
		metadata, err := getMetadata(path)
		if err != nil {
			fmt.Printf("Couldn't get metadata for %v", path)
			return nil
		}
		f = append(f, File{
			Metadata: metadata,
			Path:     path,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	fmt.Println(f)
	data, err := json.Marshal(f)
	ioutil.WriteFile("data", data, 0666)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func getData() ([]File, error) {
	var f []File
	data, err := ioutil.ReadFile("data")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	err = json.Unmarshal(data, &f)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func getMetadata(path string) (Metadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return Metadata{}, err
	}
	m, err := tag.ReadFrom(f)
	if err != nil {
		// fmt.Printf("Can't read metadata from file %s\n", path)
		return Metadata{}, err
	}
	metadata := Metadata{
		Title:  m.Title(),
		Artist: m.Artist(),
	}
	return metadata, nil
}

func play(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	d, err := mp3.NewDecoder(f)
	if err != nil {
		return err
	}
	defer d.Close()

	p, err := oto.NewPlayer(d.SampleRate(), 2, 2, 8192)
	if err != nil {
		return err
	}
	defer p.Close()

	// fmt.Printf("Length: %d[bytes]\n", d.Length())

	if _, err := io.Copy(p, d); err != nil {
		return err
	}
	return nil
}
