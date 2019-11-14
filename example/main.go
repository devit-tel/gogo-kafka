package main

import (
	"errors"
	"fmt"
	"log"

	gk "github.com/devit-tel/gogo-kafka"
	"github.com/devit-tel/gogo-kafka/retrymanager"
)

func testFuncNaruto(key string, data []byte) error {
	fmt.Println("NARUTO -> KEY: ", key, " - DATA: ", string(data))
	if string(data) == "error_1" {
		return errors.New("sample_error")
	}

	// if string(data) == "panic_1" {
	// 	panic("sample_error")
	// }

	return nil
}

func testFuncSasuke(key string, data []byte) error {
	fmt.Println("SASUKE -> KEY: ", key, " - DATA: ", string(data))
	// if string(data) == "error_1" {
	// 	return errors.New("sample_error")
	// }

	// if string(data) == "panic_1" {
	// 	panic("sample_error")
	// }

	return nil
}

func main() {
	// load config
	config := gk.NewConfig([]string{"localhost:9092"}, "kaenin")

	// create retry manager (inmem)
	rt := retrymanager.NewInmemManager(3, 2)

	// create sarama client from config
	saramaClient, err := gk.NewSaramaConsumerWithConfig(config)
	if err != nil {
		panic("Unable create sarama client!")
	}

	// create worker
	worker, err := gk.New(config, saramaClient, rt)
	if err != nil {
		log.Fatal(err)
	}

	worker.SetPanicHandler(func(err interface{}, topic, key string, data []byte) error {
		fmt.Println("Panic from: ", err)
		fmt.Printf("Debug data: %s %s - %s", topic, key, string(data))
		return nil
	})

	// register method
	worker.RegisterHandler("top_naruto", testFuncNaruto)
	worker.RegisterHandler("top_sasuke", testFuncSasuke)

	// start worker
	worker.Start()
}
