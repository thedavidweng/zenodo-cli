package model

import "fmt"

func ExampleExitCode() {
	fmt.Println(ExitCode(ErrAuthRequired))
	fmt.Println(ExitCode(ErrReadOnlyViolation))
	fmt.Println(ExitCode(ErrInterrupted))
	fmt.Println(ExitCode(ErrValidationFailed))
	// Output:
	// 2
	// 3
	// 130
	// 1
}
