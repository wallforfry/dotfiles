package telemetry

import (
	"errors"
	"fmt"
)

type ReadInterruptedError struct {
	Source string
	Cause  error
}

func (err *ReadInterruptedError) Error() string {
	return fmt.Sprintf("%s : lecture interrompue", err.Source)
}

func (err *ReadInterruptedError) Unwrap() error { return err.Cause }

func readInterrupted(source string, cause error) error {
	var existing *ReadInterruptedError
	if errors.As(cause, &existing) {
		return cause
	}
	return &ReadInterruptedError{Source: source, Cause: cause}
}

type FormatDriftError struct {
	Unknown int
	Invalid int
}

func (err *FormatDriftError) Error() string {
	return fmt.Sprintf("formats non mesurés : %d inconnus, %d invalides", err.Unknown, err.Invalid)
}

func ValidateFormats(summary Summary) error {
	unknown := sumValues(summary.UnknownRecords)
	invalid := sumValues(summary.InvalidRecords)
	if unknown == 0 && invalid == 0 {
		return nil
	}
	return &FormatDriftError{Unknown: unknown, Invalid: invalid}
}
