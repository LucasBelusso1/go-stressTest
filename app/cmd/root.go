/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
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

var (
	requestUrl       string
	totalRequestsQty int
	concurrency      int
)

var rootCmd = &cobra.Command{
	Use:   "go-stressTest",
	Short: "Stress test tool using concurrency.",
	Long:  `This tool uses the concurrency concept of GO programming lenguage to stress some URL.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		parsedUrl, err := url.Parse(requestUrl)
		if err != nil {
			panic("invalid URL")
		}

		requestUrl = parsedUrl.String()

		if totalRequestsQty <= 0 {
			panic("invalid requests value")
		}

		concurrency, err := cmd.Flags().GetInt("concurrency")
		if err != nil {
			panic(err)
		}

		if concurrency <= 0 {
			panic("invalid concurrency value")
		}

		if totalRequestsQty < concurrency {
			panic("requests must be bigger than concurrency")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		responseData := make(chan Resp, totalRequestsQty)
		done := make(chan bool)

		for i := 0; i < concurrency; i++ {
			go makeRequests(responseData, done)
		}

		var report Report
		report.httpCodes = make(map[int]int)

		go func() {
			for i := 0; i < totalRequestsQty; i++ {
				reportRegister := <-responseData
				_, ok := report.httpCodes[reportRegister.httpCode]
				if !ok {
					report.httpCodes[reportRegister.httpCode] = 1
				} else {
					report.httpCodes[reportRegister.httpCode]++
				}
				report.totalExecutionTime += reportRegister.executionTime
			}
			done <- true
		}()

		<-done

		fmt.Println("Execution time in seconds: ", report.totalExecutionTime.Seconds())
		fmt.Println("Total of requests:", totalRequestsQty)
		fmt.Println("Total of http code 200:", report.httpCodes[200])
		fmt.Println("All http codes")

		for httpCode, qty := range report.httpCodes {
			if httpCode == -1 {
				fmt.Printf("Quantity of Errors (-1): %v\n", qty)
			} else {
				fmt.Printf("Quantity of %v: %v\n", httpCode, qty)
			}
		}
		return nil
	},
}

func makeRequests(responseData chan<- Resp, done <-chan bool) {
	for {
		select {
		case <-done:
			return
		default:
			startTime := time.Now()
			res, err := http.DefaultClient.Get(requestUrl)
			if err != nil {
				responseData <- Resp{
					httpCode:      -1,
					executionTime: time.Since(startTime),
				}
				continue
			}

			responseData <- Resp{
				httpCode:      res.StatusCode,
				executionTime: time.Since(startTime),
			}

			res.Body.Close()
		}
	}
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&requestUrl, "url", "", "URL of the service to be tested.")
	rootCmd.Flags().IntVar(&totalRequestsQty, "requests", 1, "Total number of requests.")
	rootCmd.Flags().IntVar(&concurrency, "concurrency", 1, "Number of simultaneous calls.")
}
