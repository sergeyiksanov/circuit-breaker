# Circuit Breaker

Circuit Breaker is a library designed to block repeated requests with high failure rates.

## Installation

Use the go installer to install circuit-breaker.

```bash
go get github.com/sergeyiksanov/circuit-breaker
```

## Usage

```go
import ""

// Simulating an unstable service
func unreliableService() (any, error) {
	if time.Now().Unix()%2 == 0 {
		return nil, errors.New("service failed")
	}

	return nil, nil
}

func main() {
    controller := cb.NewController(
		2,             // Failure threshold
		2*time.Second, // Recovery time
		2,             // Half-open max request
		2*time.Second, // Half-open max time
	)

	for i := 0; i < 5; i++ {
		result, err := controller.Call(unreliableService) // Request to service 
		if err != nil {
			log.Warnf("Service request failed with error: %s", err)
		} else {
			log.Infof("Service request succeeded with result: %s", result)
		}

		time.Sleep(1 * time.Second)
		log.Info("-------------------------------------")
	}
}
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

[MIT](https://github.com/sergeyiksanov/circuit-breaker/blob/main/LICENSE)
