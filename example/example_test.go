package main

import (
	"fmt"
	"log"
	"os/exec"
)

func Example() {
	cmd := exec.Command("go", "run", "./app")

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("go run ./app failed: %v\n%s", err, out)
	}

	fmt.Print(string(out))

	// Output:
	// Hello, World!
	// 2 + 3 = 5
	// 10 + 20 = 30
	// notification sent
}
