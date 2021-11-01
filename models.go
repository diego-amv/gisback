package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"
)

type PageData struct {
	ContentTitle string
	LoginURL     string
	Path         string
	MapsApiKey   string
	LoggedUser   bool
	Items        []GISMessage
	Page         Page
}

type PageItem struct {
	Name   int
	URL    string
	Active bool
}

type Page struct {
	Current  int
	Previous int
	Next     int
	Pages    []PageItem
	Offset   int
}

func (p *Page) build(current int, total int, path string) {
	if current == 0 {
		p.Current = 1
	} else {
		p.Current = current
	}
	pages := total / 5
	pagesRest := total % 5
	if pagesRest > 0 {
		pages++

	}
	if p.Current > 1 {
		p.Previous = p.Current - 1
	}
	if p.Current != pages {
		p.Next = p.Current + 1
	}
	if p.Current > 1 {
		p.Offset = (p.Current - 1) * 5
	}
	p.Pages = make([]PageItem, pages)
	for i := 1; pages >= i; i++ {
		p.Pages[i-1] = PageItem{
			Name:   i,
			URL:    fmt.Sprintf("/%s?page=%d", path, i),
			Active: i == p.Current,
		}
	}
}

type GISMessage struct {
	Id        int64     `json:"id" datastore:"-"`
	IMEI      int64     `json:"i"`
	Conn      string    `json:"c"`
	Time      time.Time `json:"t"`
	Altitude  float64   `json:"al"`
	Latitude  float64   `json:"la"`
	Longitude float64   `json:"lo"`
}

func decodeGIS(r io.ReadCloser) (*GISMessage, error) {
	defer func(r io.ReadCloser) {
		err := r.Close()
		if err != nil {
			log.Printf("Error closing body")
		}
	}(r)
	var gis GISMessage
	err := json.NewDecoder(r).Decode(&gis)
	if err == nil {
		return &gis, err
	}
	if time.Time.IsZero(gis.Time) {
		gis.Time = time.Now()
	}
	return &gis, nil
}
