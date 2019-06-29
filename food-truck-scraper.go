package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/gocolly/colly"
)

type FoodTruck struct {
	Name        string
	TypesOfFood []string
	URL         string
}

type TypeOfFoodDetail struct {
	TypeOfFood string
	URL        string
}

var foodTrucks []*FoodTruck
var foodTruckLinksChannel chan *FoodTruck
var typesOfFoodChannel chan *TypeOfFoodDetail

//var wait *sync.WaitGroup

func main() {
	baseURL := "https://roaminghunger.com"
	foodTrucks = make([]*FoodTruck, 0, 100)

	c1 := colly.NewCollector()
	c2 := colly.NewCollector()

	/*
	 * Food truck card
	 */
	c1.OnHTML("div.col-sm-4 div.square_truck div.content h4 a", func(e *colly.HTMLElement) {
		resturaunt := e.Text
		linkURL := fmt.Sprintf("%s%s", baseURL, e.Attr("href"))

		newTruck := &FoodTruck{
			Name:        resturaunt,
			TypesOfFood: make([]string, 0, 10),
			URL:         linkURL,
		}

		foodTruckLinksChannel <- newTruck
	})

	/*
	 * Paging link
	 */
	c1.OnHTML("div.paging span a.number", func(e *colly.HTMLElement) {
		linkURL := e.Attr("href")

		if !strings.Contains(linkURL, "/1/") {
			e.Request.Visit(fmt.Sprintf("%s%s", baseURL, linkURL))
		}
	})

	c1.OnRequest(func(r *colly.Request) {
		fmt.Printf("Visiting %s\n", r.URL)
	})

	c2.OnHTML("div.tags div.row-flex div.col-flex a", func(e *colly.HTMLElement) {
		typesOfFoodChannel <- &TypeOfFoodDetail{TypeOfFood: e.Text, URL: e.Request.URL.String()}
	})

	c2.OnRequest(func(r *colly.Request) {
		//fmt.Printf("Getting details at %s\n", r.URL)
	})

	/**************************************************************************
	 * This section gets a list of all food trucks
	 *************************************************************************/
	foodTruckLinksChannel = make(chan *FoodTruck, 20)

	go func() {
		for f := range foodTruckLinksChannel {
			foodTrucks = append(foodTrucks, f)
		}
	}()

	c1.Visit(fmt.Sprintf("%s/food-trucks/tx/houston/1", baseURL))

	/**************************************************************************
	 * This section gets the details for each food truck
	 *************************************************************************/
	typesOfFoodChannel = make(chan *TypeOfFoodDetail, 10)

	go func() {
		for typeOfFood := range typesOfFoodChannel {
			f := findFoodTruck(typeOfFood.URL)

			if f != nil {
				f.TypesOfFood = append(f.TypesOfFood, typeOfFood.TypeOfFood)
			}

			//wait.Done()
		}
	}()

	fmt.Printf("\nNumber of food trucks: %d\n", len(foodTrucks))
	//wait = &sync.WaitGroup{}

	//wait.Add(len(foodTrucks))

	for index, f := range foodTrucks {
		c2.Visit(f.URL)
		fmt.Printf("%d - %s\n", index, f.URL)
	}

	//c2.Wait()
	//wait.Wait()

	/**************************************************************************
	 * Assemble the results
	 *************************************************************************/
	fmt.Printf("\n\n")

	b := &bytes.Buffer{}
	w := csv.NewWriter(b)

	for _, f := range foodTrucks {
		record := []string{
			f.Name,
			strings.Join(f.TypesOfFood, ", "),
			f.URL,
		}

		w.Write(record)
	}

	w.Flush()

	ioutil.WriteFile("./food-trucks.csv", b.Bytes(), 0644)
	close(foodTruckLinksChannel)
	close(typesOfFoodChannel)
}

func findFoodTruck(url string) *FoodTruck {
	for _, f := range foodTrucks {
		if f.URL == url {
			return f
		}
	}

	return nil
}
