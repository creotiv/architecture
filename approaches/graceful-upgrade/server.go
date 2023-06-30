package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/sys/unix"
)

func main() {
	network := "tcp"
	address := ":8881"

    lc := net.ListenConfig{
        Control: func(network, address string, c syscall.RawConn) error {
            var opErr error
            if err := c.Control(func(fd uintptr) {
                opErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
            }); err != nil {
                return err
            }

            return opErr
        },
    }

    l, err := lc.Listen(context.Background(), network, address)
    if err != nil {
        log.Fatalf("Listen error: %v", err)
    }

	fmt.Println("Serving in port", address)

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		time.Sleep(200*time.Millisecond)
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.Listener = l

	// e.Start("")

	go func() {
		e.Start("")
	}()

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, syscall.SIGTERM)

	<-interruptChan

	fmt.Println("Shutting down server...")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := e.Shutdown(ctx); err != nil {
        fmt.Printf("Could not gracefully shutdown the server: %v\n", err)
    }
    fmt.Println("Server stopped")
}
