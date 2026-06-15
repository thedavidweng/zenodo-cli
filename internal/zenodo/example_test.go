package zenodo

import "fmt"

func ExampleNewClient() {
	c := NewClient("https://sandbox.zenodo.org", "my-api-token")
	fmt.Println("BaseURL:", c.BaseURL)
	fmt.Println("Retries:", c.Retries)
	// Output:
	// BaseURL: https://sandbox.zenodo.org
	// Retries: 3
}
