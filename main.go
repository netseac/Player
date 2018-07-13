package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/dhowden/tag"
	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"
)

type Metadata struct {
	Title  string
	Artist string
}

type File struct {
	Metadata
	Path string
}

func main() {
	var path string
	flag.StringVar(&path, "path", "", "set path for playing")
	flag.Parse()

	if path == "" {
		fmt.Println("Please set path.")
		os.Exit(1)
	}

	p, err := getPlaylist(path)
	if err != nil {
		panic(err)
	}

	for _, f := range p {
		fmt.Printf("Current playing: %v - %v \n", f.Artist, f.Title)
		play(f.Path)
	}
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

	if _, err := io.Copy(p, d); err != nil {
		return err
	}
	return nil
}

func getPlaylist(root string) ([]File, error) {
	var f []File

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if path == root {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".mp3" {
			return nil
		}
		m, err := getMetadata(path)
		if err != nil {
			return nil
		}
		f = append(f, File{
			Metadata: m,
			Path:     path,
		})
		return nil
	})
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
		return Metadata{}, err
	}
	metadata := Metadata{
		Title:  m.Title(),
		Artist: m.Artist(),
	}
	return metadata, nil
}
