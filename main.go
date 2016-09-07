package main

import (
	"bytes"
	"flag"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
)

var (
	httpListen  = flag.String("http", "127.0.0.1:5000", "host:port to listen on")
	openBrowser = flag.Bool("openbrowser", true, "open browser automatically")
)

func main() {
	r := gin.Default()
	m := melody.New()

	r.GET("/", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "views/index.html")
	})

	r.GET("/mini.css", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "views/mini.css")
	})

	r.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		pigeon := exec.Command("pigeon")
		pigeon.Stdin = strings.NewReader(string(msg))

		var out, stderr bytes.Buffer
		pigeon.Stdout = &out
		pigeon.Stderr = &stderr
		err := pigeon.Run()
		if err != nil {
			log.Printf("STDERR: %v", stderr.String())
		}
		m.Broadcast([]byte(out.String()))
	})
	httpAddr := getHTTPAddr()
	go func() {
		url := "http://" + httpAddr
		if waitServer(url) && *openBrowser && startBrowser(url) {
			log.Printf("A browser window should open. If not, please visit %s", url)
		} else {
			log.Printf("Please open your web browser and visit %s", url)
		}
	}()

	r.Run(httpAddr)

}

func getHTTPAddr() string {
	host, port, err := net.SplitHostPort(*httpListen)
	if err != nil {
		log.Fatal(err)
	}
	if host == "" {
		host = "localhost"
	}
	if host != "127.0.0.1" && host != "localhost" {
		log.Print(localhostWarning)
	}
	httpAddr := host + ":" + port
	return httpAddr
}

const localhostWarning = `
WARNING!  WARNING!  WARNING!
I appear to be listening on an address that is not localhost.
Anyone with access to this address and port will have access
to this machine as the user running gotour.
If you don't understand this message, hit Control-C to terminate this process.
WARNING!  WARNING!  WARNING!
`

// waitServer waits some time for the http Server to start
// serving url. The return value reports whether it starts.
func waitServer(url string) bool {
	tries := 20
	for tries > 0 {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			return true
		}
		time.Sleep(100 * time.Millisecond)
		tries--
	}
	return false
}

// startBrowser tries to open the URL in a browser, and returns
// whether it succeed.
func startBrowser(url string) bool {
	// try to start the browser
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open"}
	case "windows":
		args = []string{"cmd", "/c", "start"}
	default:
		args = []string{"xdg-open"}
	}
	cmd := exec.Command(args[0], append(args[1:], url)...)
	return cmd.Start() == nil
}
