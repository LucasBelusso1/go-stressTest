/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

type Resp struct {
	httpCode      int
	executionTime time.Duration
}

type Report struct {
	httpCodes          map[int]int
	totalExecutionTime time.Duration
}

var rootCmd = &cobra.Command{
	Use:   "go-stressTest",
	Short: "Stress test tool using concurrency.",
	Long:  `This tool uses the concurrency concept of GO programming lenguage to stress some URL.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		urlAddress, err := cmd.Flags().GetString("url")
		if err != nil {
			return err
		}

		requestUrl, err := url.Parse(urlAddress)
		if err != nil {
			return errors.New("invalid URL")
		}

		requests, err := cmd.Flags().GetInt("requests")
		if err != nil {
			return err
		}

		if requests <= 0 {
			return errors.New("invalid requests value")
		}

		concurrency, err := cmd.Flags().GetInt("concurrency")
		if err != nil {
			return err
		}

		if concurrency <= 0 {
			return errors.New("invalid concurrency value")
		}

		requestQty := int(requests / concurrency)
		wg := &sync.WaitGroup{}
		wg.Add(requests)

		responseData := make(chan Resp, concurrency)

		for i := 0; i < concurrency; i++ {
			go makeRequests(wg, requestUrl, responseData, requestQty)
		}

		var report Report
		report.httpCodes = make(map[int]int)

		go func() {
			for {
				select {
				case reportRegister := <-responseData:
					_, ok := report.httpCodes[reportRegister.httpCode]
					if !ok {
						report.httpCodes[reportRegister.httpCode] = 1
					} else {
						report.httpCodes[reportRegister.httpCode]++
					}
					report.totalExecutionTime += reportRegister.executionTime
				}
			}
		}()

		wg.Wait()

		fmt.Println("Execution time in seconds: ", report.totalExecutionTime.Seconds())
		fmt.Println("Total of requests:", requests)
		fmt.Println("Total of http code 200:", report.httpCodes[200])
		fmt.Println("All http codes")

		for httpCode, qty := range report.httpCodes {
			fmt.Printf("Quantity of %v: %v\n", httpCode, qty)
		}
		return nil
	},
}

func makeRequests(wg *sync.WaitGroup, url *url.URL, responseData chan<- Resp, requestQty int) {
	req, err := http.NewRequest("GET", url.String(), nil)

	if err != nil {
		panic(err)
	}

	for i := 0; i <= requestQty; i++ {
		startTime := time.Now()

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(err)
			wg.Done()
			continue
		}

		res.Body.Close()

		responseData <- Resp{
			httpCode:      res.StatusCode,
			executionTime: time.Since(startTime),
		}

		wg.Done()
	}
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().String("url", "", "URL of the service to be tested.")
	rootCmd.Flags().Int("requests", 1, "Total number of requests.")
	rootCmd.Flags().Int("concurrency", 1, "Number of simultaneous calls.")
}
