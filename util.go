package main

import (
	"encoding/json"
	"net/http"

	"github.com/brutella/hc/log"
)

func mapInt(x, in_min, in_max, out_min, out_max int) int {
	return (x-in_min)*(out_max-out_min)/(in_max-in_min) + out_min
}

func getJSON(url string, result interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		log.Info.Printf("cannot fetch URL %q: %v", url, err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Info.Printf("unexpected http GET status: %s", resp.Status)
		return err
	}

	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		log.Info.Printf("cannot decode JSON: %v", err)
		return err
	}
	return nil
}
