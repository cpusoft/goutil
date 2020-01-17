package errorutil

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorChanToError(t *testing.T) {
	errCh := make(chan error, 10)
	err1 := errors.New("11")
	err2 := errors.New("22")
	err3 := errors.New("33")
	err4 := errors.New("44")
	errCh <- err1
	errCh <- err2
	errCh <- err3
	errCh <- err4
	errs := CloseErrorChanToError(errCh)

	fmt.Println(errs)

}

func TestErrorChanToError2(t *testing.T) {
	errCh := make(chan error, 10)
	errs := CloseErrorChanToError(errCh)

	fmt.Println(errs)

}
