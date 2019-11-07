# gogo-kafka
Kafka worker by Sarama (enhance support recovery and retry process)

<p align="center">
  <a href="https://github.com/devit-tel/gogo-kafka"><img alt="GitHub Actions status" src="https://github.com/devit-tel/gogo-kafka/workflows/go-unit-test/badge.svg"></a>
</p>

### Run example
Run docker kafka by lenses (http://localhost:3030 user: admin, password: admin)
```shell script
    docker run -e ADV_HOST=127.0.0.1 -e EULA="https://dl.lenses.io/d/?id=8914c158-2090-4132-b00a-ebc3175800c0" --rm -p  3030:3030 -p 9092:9092 lensesio/box
```


### Example

```go
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

        // set custom recovery when panic
	worker.SetPanicHandler(func(err interface{}) {
		fmt.Println("Error in panic: ", err)
	})

	// register method
	worker.RegisterHandler("konohax", testFunc)

	// start worker
	worker.Start()
```