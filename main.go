package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/JannoTjarks/tankerkoenig/api"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
)

func main() {
	var cfg Config
	if len(os.Args[1:]) != 1 {
		err := fmt.Errorf("You need to specify a path to the configuration yaml-file")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	readConfig(&cfg, os.Args[1])
	var stationResults []StationResult = getOpenStations(cfg)
	var client mqtt.Client = connectMqttClient(cfg)
	publishFuelWithMqtt(client, stationResults)
	time.Sleep(6 * time.Second)

	client.Disconnect(250)
	time.Sleep(1 * time.Second)
}

func readConfig(cfg *Config, path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		log.Fatal(err)
	}
}

func publishFuelWithMqtt(client mqtt.Client, stationResults []StationResult) {
	for _, station := range stationResults {
		tokenE5 := client.Publish("fuel/"+station.Name+"/e5",
			0, false, fmt.Sprintf("%f", station.E5))
		tokenE5.Wait()
		tokenE10 := client.Publish("fuel/"+station.Name+"/e10",
			0, false, fmt.Sprintf("%f", station.E10))
		tokenE10.Wait()
		tokenDiesel := client.Publish("fuel/"+station.Name+"/diesel",
			0, false, fmt.Sprintf("%f", station.Diesel))
		tokenDiesel.Wait()
	}
}

func getOpenStations(cfg Config) []StationResult {
	var stationResults []StationResult

	for _, station := range cfg.Stations {
		var response string = api.RequestPrice(cfg.APIKey, station.Id)
		var priceInfo TankerkoenigResponse
		json.Unmarshal([]byte(response), &priceInfo)
		if priceInfo.Ok != false {
			// TODO handle error
		}

		var status string = gjson.Get(
			response, "prices."+station.Id+".status").String()

		if status != "open" {
			continue
		}

		var e10 float64 = gjson.Get(
			response, "prices."+station.Id+".e10").Float()
		var e5 float64 = gjson.Get(
			response, "prices."+station.Id+".e5").Float()
		var diesel float64 = gjson.Get(
			response, "prices."+station.Id+".diesel").Float()

		stationResult := StationResult{
			Name:   station.Name,
			Status: status,
			E5:     e5,
			E10:    e10,
			Diesel: diesel,
		}

		stationResults = append(stationResults, stationResult)
	}

	return stationResults
}

type Config struct {
	Broker   string `yaml:"mqttBroker"`
	Port     string `yaml:"mqttPort"`
	APIKey   string `yaml:"apiKey"`
	Stations []Station
}

type Station struct {
	Id       string `yaml:"id"`
	Name     string `yaml:"name"`
	Location string `yaml:"location"`
}

type TankerkoenigResponse struct {
	Ok      bool   `json:"ok"`
	License string `json:"license"`
	Data    string `json:"data"`
}

type StationResult struct {
	Name   string
	Status string
	E5     float64
	E10    float64
	Diesel float64
}
