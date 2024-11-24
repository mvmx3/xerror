package xerror

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func D(ch chan error) {
	err := errors.New("test error")
	werr := WrapSend(err)
	ch <- werr
}

func C(ch chan error) error {
	err := <-ch
	return WrapReceive(err)
}

func TestChannelError(t *testing.T) {
	fmt.Println("=== Channel Error Test ===")
	ch := make(chan error)
	go D(ch)
	err := C(ch)
	fmt.Printf("Error trace:\n%v\n\n", err)
}

func A() error {
	err := B()
	if err != nil {
		return WrapError(err)
	}
	return nil
}

func B() error {
	err := errors.New("test error")
	if err != nil {
		return WrapError(err)
	}
	return nil
}

func TestWrapError(t *testing.T) {
	fmt.Println("=== Debug Mode ===")
	debugMode = true
	err := A()
	fmt.Printf("Error trace:\n%v\n\n", err)
	time.Sleep(time.Second * 10)

}
