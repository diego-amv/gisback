package main

import (
	"cloud.google.com/go/datastore"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/appengine/v2"
	"google.golang.org/appengine/v2/user"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

const ProjectId = "id-de-tu-proyecto"
const MapsAPIKey = "tu-api-key-de-google-maps"

func main() {
	// Se establecen controladores de requests
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/track", newGISHandler)
	http.HandleFunc("/search", searchHandler)

	// Validamos el entorno de ejecución para pruebas
	if appengine.IsAppEngine() && appengine.IsStandard() {
		log.Println("Inicia servicio App Engine")
		appengine.Main()
	} else {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
			log.Printf("Defaulting to port %s", port)
		}
		log.Printf("Listening on port %s", port)
		log.Printf("Open http://localhost:%s in the browser", port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	var gisItems []GISMessage
	var page Page
	var loggedUser = true
	if appengine.IsAppEngine() && appengine.IsStandard() {
		log.Println("Se detecta en AppEngine")
		c := appengine.NewContext(r)
		u := user.Current(c)
		if u == nil || !u.Admin {
			loggedUser = false
		}
		page = getPage(r.URL.Query(), nil, c, "", 0)
		gisItems, _ = getAllGIS(c, page.Offset)
	} else {
		log.Println("Se detecta en Local")
		client, err := datastore.NewClient(r.Context(), ProjectId)
		if err != nil {
			log.Printf("Error en generación de cliente datastore")
		}
		defer func(client *datastore.Client) {
			err := client.Close()
			if err != nil {
				log.Printf("Error en cierre de cliente datastore")
			}
		}(client)
		page = getPage(r.URL.Query(), client, r.Context(), "", 0)
		gisItems, _ = getAllGISLocal(client, r.Context(), page.Offset)
	}

	loginUrl, _ := user.LoginURL(r.Context(), "/")
	tmpl := template.Must(template.ParseFiles([]string{filepath.Join("templates/base.html")}...))

	data := PageData{
		ContentTitle: "Últimos 5 registros",
		LoginURL:     loginUrl,
		MapsApiKey:   MapsAPIKey,
		LoggedUser:   loggedUser,
		Items:        gisItems,
		Page:         page,
	}

	log.Printf("Data: %+v", data)
	err := tmpl.Execute(w, data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func newGISHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		gis, err := decodeGIS(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		var res *GISMessage
		if appengine.IsAppEngine() && appengine.IsStandard() {
			c := appengine.NewContext(r)
			res, err = gis.save(c)
		} else {
			client, err := datastore.NewClient(r.Context(), ProjectId)
			if err != nil {
				log.Printf("Error en generación de cliente datastore")
			}
			defer func(client *datastore.Client) {
				err := client.Close()
				if err != nil {
					log.Printf("Error en cierre de cliente datastore")
				}
			}(client)
			res, err = gis.saveLocal(client, r.Context())
		}
		if err == nil {
			err = json.NewEncoder(w).Encode(res)
		}
		if err != nil {
			log.Printf("gis error: %#v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		queryParams := r.URL.Query()
		imei := queryParams.Get("q")
		regex, _ := regexp.Compile(`\d{15}`)
		if regex.MatchString(imei) {
			imeiValue, _ := strconv.Atoi(imei)
			if r.URL.Path != "/search" {
				http.NotFound(w, r)
				return
			}
			var gisItems []GISMessage
			var page Page
			var loggedUser = true
			if appengine.IsAppEngine() && appengine.IsStandard() {
				log.Println("Se detecta en AppEngine")
				c := appengine.NewContext(r)
				u := user.Current(c)
				if u == nil || !u.Admin {
					loggedUser = false
				}
				page = getPage(r.URL.Query(), nil, c, "search", imeiValue)
				gisItems, _ = searchGIS(c, imei, page.Offset)
			} else {
				log.Println("Se detecta en Local")
				client, err := datastore.NewClient(r.Context(), ProjectId)
				if err != nil {
					log.Printf("Error en generación de cliente datastore")
				}
				defer func(client *datastore.Client) {
					err := client.Close()
					if err != nil {
						log.Printf("Error en cierre de cliente datastore")
					}
				}(client)
				page = getPage(r.URL.Query(), client, r.Context(), "search", imeiValue)
				gisItems, _ = searchGISLocal(client, r.Context(), imei, page.Offset)
			}

			loginURL, _ := user.LoginURL(r.Context(), "/search")
			tmpl := template.Must(template.ParseFiles([]string{filepath.Join("templates/base.html")}...))

			data := PageData{
				ContentTitle: fmt.Sprintf("Registros IMEI: %s", imei),
				Path:         "search",
				MapsApiKey:   MapsAPIKey,
				LoggedUser:   loggedUser,
				LoginURL:     loginURL,
				Items:        gisItems,
				Page:         page,
			}

			log.Printf("Data: %+v", data)
			err := tmpl.Execute(w, data)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}

func getPage(params url.Values, c *datastore.Client, ctx context.Context, path string, imei int) Page {
	var page = Page{}
	current, _ := strconv.Atoi(params.Get("page"))
	var total int
	if appengine.IsAppEngine() && appengine.IsStandard() {
		total = count(ctx, imei)
	} else {
		total = countLocal(c, ctx, imei)
	}
	page.build(current, total, path)
	return page
}
