package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Henry-Sarabia/igdb/v2"
	"gopkg.in/yaml.v2"
)

var key string
var token string
var name string

type Release struct {
	Region       string `yaml:"Region"`
	Release_date string `yaml:"Release_date"`
}

type Description struct {
	Name       string    `yaml:"Name"`
	Genre      []string  `yaml:"Genre"`
	Platform   string    `yaml:"Platform"`
	Rating     float64   `yaml:"Rating"`
	IGDB_URL   string    `yaml:"IGDB_URL"`
	Story_Line string    `yaml:"Story_Line"`
	Summary    string    `yaml:"Summary"`
	Release    []Release `yaml:"Release"`
}

func init() {
	flag.StringVar(&key, "k", "", "Client-ID")
	flag.StringVar(&token, "t", "", "AppAccessToken")
	flag.StringVar(&name, "n", "", "GameName")
	flag.Parse()
}

func main() {
	if key == "" {
		fmt.Println("No key provided. Please run: app -k YOUR_CLIENT_ID -t YOUR_APP_ACCESS_TOKEN -name GAME_NAME")
		return
	}
	if token == "" {
		fmt.Println("No token provided. Please run: app -k YOUR_CLIENT_ID -t YOUR_APP_ACCESS_TOKEN -name GAME_NAME")
		return
	}
	if name == "" {
		fmt.Println("No token provided. Please run: app -k YOUR_CLIENT_ID -t YOUR_APP_ACCESS_TOKEN -name GAME_NAME")
		return
	}

	c := igdb.NewClient(key, token, nil)

	games, err := c.Games.Search(
		name,
		igdb.SetFields("*"),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Found: ", len(games), " Games")
	order := 0
	for _, g := range games {
		for _, platform := range g.Platforms {
			pm, err := c.Platforms.Get(platform, igdb.SetFields("name"))
			if err != nil {
				log.Fatal(err)
			}
			if (pm.ID == 48) || (pm.ID == 165) {
				description := new(Description)
				fmt.Println("++++++++++++++++++++++++++++++++++")
				order += 1
				var Name = name + "_" + strconv.Itoa(order)
				if err := os.Mkdir(Name, os.ModePerm); err != nil {
					log.Fatal(err)
				}
				description.Name = g.Name
				fmt.Println("Name: ", g.Name)
				for _, genre := range g.Genres {
					rg, err := c.Genres.Get(genre, igdb.SetFields("name"))
					if err != nil {
						log.Fatal(err)
					}
					description.Genre = append(description.Genre, rg.Name)
					fmt.Printf("Genre: %v\n", rg.Name)
				}
				description.Rating = g.AggregatedRating
				fmt.Println("Rating: ", g.AggregatedRating)
				description.IGDB_URL = g.URL
				fmt.Println("IGDB URL: ", g.URL)

				description.Story_Line = g.Storyline
				fmt.Println("Story Line: ", g.Storyline)
				description.Summary = g.Summary
				fmt.Println("Summary: ", g.Summary)
				description.Platform = pm.Name
				fmt.Println("Platform: ", pm.Name)

				//165, 48 - PS4, PS4 VR
				for _, rdate := range g.ReleaseDates {
					date, err := c.ReleaseDates.Get(rdate, igdb.SetFields("human"))
					if err != nil {
						log.Fatal(err)
					}
					pfm, err := c.ReleaseDates.Get(rdate, igdb.SetFields("platform"))
					if err != nil {
						log.Fatal(err)
					}
					reg, err := c.ReleaseDates.Get(rdate, igdb.SetFields("region"))
					if err != nil {
						log.Fatal(err)
					}
					if (pfm.Platform == 48) || (pfm.Platform == 165) {
						rel := new(Release)
						rel.Region = reg.Region.String()
						rel.Release_date = date.Human
						description.Release = append(description.Release, *rel)
						fmt.Println(reg.Region)
						fmt.Println("Release date: ", date.Human)
					}
				}

				yamlData, err := yaml.Marshal(&description)

				if err != nil {
					fmt.Printf("Error while Marshaling. %v", err)
				}

				err = ioutil.WriteFile(Name+"/Description.yaml", yamlData, 0644)
				if err != nil {
					panic("Unable to write data into the file")
				}

				cover, err := c.Covers.Get(g.Cover, igdb.SetFields("image_id"))
				if err != nil {
					log.Fatal(err)
				}
				img, err := cover.SizedURL(igdb.Size1080p, 1) // resize to largest image available
				if err != nil {
					log.Fatal(err)
				}
				download(Name, img, g.Name)

				screenShots, err := c.Screenshots.List(g.Screenshots)
				if err != nil {
					fmt.Println(err)
				}
				if len(screenShots) > 0 {
					for num, scr := range screenShots {
						s, err := c.Screenshots.Get(scr.ID, igdb.SetFields("image_id"))
						if err != nil {
							log.Fatal(err)
						}
						image, err := s.SizedURL(igdb.Size1080p, 1)
						if err != nil {
							log.Fatal(err)
						}
						download_screenshots(Name, image, g.Name, num)
					}
				}
				fmt.Println("__________________________")
			}
		}

	}
	return
}

func download(path string, url string, filename string) (err error) {

	fmt.Println("Downloading ", url, " to ", filename)

	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var fname string
	if strings.Contains(url, ".jpg") {
		fname = filename + ".jpg"
	} else {
		strings.Contains(url, ".png")
		fname = filename + ".png"
	}
	f, err := os.Create(path + "/" + fname)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return
}

func download_screenshots(path string, url string, filename string, number int) (err error) {

	fmt.Println("Downloading ", url, " to ", filename)

	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var fname string
	if strings.Contains(url, ".jpg") {
		fname = filename + "_scr_" + strconv.Itoa(number) + ".jpg"
	} else {
		strings.Contains(url, ".png")
		fname = filename + "_scr_" + strconv.Itoa(number) + ".png"
	}
	f, err := os.Create(path + "/" + fname)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return
}
