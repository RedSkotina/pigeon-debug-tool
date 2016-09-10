package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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

	r.GET("/app.css", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "views/app.css")
	})
	r.GET("codemirror/lib/codemirror.js", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "views/codemirror/lib/codemirror.js")
	})
	r.GET("codemirror/lib/codemirror.css", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "views/codemirror/lib/codemirror.js")
	})
	r.GET("codemirror/mode/pegjs/pegjs.js", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "views/codemirror/lib/codemirror.js")
	})
    r.GET("bootstrap/3.3.7/css/bootstrap.min.css", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "views/bootstrap/3.3.7/css/bootstrap.min.css")
	})
    r.GET("bootstrap/3.3.7/css/bootstrap-theme.min.css", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "views/bootstrap/3.3.7/css/bootstrap-theme.min.css")
	})
    r.GET("bootstrap/3.3.7/js/bootstrap.min.js", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "views/bootstrap/3.3.7/js/bootstrap.min.js")
	})
    r.GET("bootstrap/3.3.7/css/bootstrap.min.css.map", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "views/bootstrap/3.3.7/css/bootstrap.min.css.map")
	})
    r.GET("bootstrap/3.3.7/css/bootstrap-theme.min.css.map", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "views/bootstrap/3.3.7/css/bootstrap-theme.min.css.map")
	})
    r.GET("tether.min.js", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "views/tether.min.js")
	})
    r.GET("jquery-3.1.0.min.js", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "views/jquery-3.1.0.min.js")
	})
    
	r.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleMessage(func(s *melody.Session, bmsg []byte) {
		var msg Msg
		if err := json.Unmarshal(bmsg, &msg); err != nil {
			log.Printf("json.Unmarshall: %v", err)
		}
		parser := generateParser(msg.Grammar)
		out := compileAndRunGoSource(parser, msg.TestString)
		jtrace := buildJsonTrace(out)
		m.Broadcast(jtrace)
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

// Msg for communicating with frontend
type Msg struct {
	Grammar    string `json:"grammar"`
	TestString string `json:"test_string"`
}

func generateParser(msg string) bytes.Buffer {
	pigeon := exec.Command("pigeon")
	pigeon.Stdin = strings.NewReader(msg)

	var pigout, pigerr bytes.Buffer
	pigeon.Stdout = &pigout
	pigeon.Stderr = &pigerr
	err := pigeon.Run()
	if err != nil {
		log.Printf("PIGEON STDERR: %v", pigerr.String())
	}
	var t bytes.Buffer
	t.WriteString(mainFunc + pigout.String())
	goimports := exec.Command("goimports")
	goimports.Stdin = strings.NewReader(t.String())
	var impout, imperr bytes.Buffer
	goimports.Stdout = &impout
	goimports.Stderr = &imperr
	err = goimports.Run()
	if err != nil {
		log.Printf("GOIMPORTS STDERR: %v", imperr.String())
	}
	return impout
}

const mainFunc = `
func main() {
	_, err := ParseReader("stdin", os.Stdin, Debug(true))
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(got)
}
`

func compileAndRunGoSource(source bytes.Buffer, test string) bytes.Buffer {
	tmpfilename := TempFileName("pigeon", ".go")
	err := ioutil.WriteFile(tmpfilename, source.Bytes(), 0644)
	defer os.Remove(tmpfilename)

	log.Printf("go run %v", tmpfilename)
	gorun := exec.Command("go", "run", tmpfilename)
	gorun.Stdin = strings.NewReader(test)

	var out, stderr bytes.Buffer
	gorun.Stdout = &out
	gorun.Stderr = &stderr
	err = gorun.Run()
	if err != nil {
		//log.Printf("GORUN STDERR: %v", stderr.String())
		out.Write(stderr.Bytes())
	}
	return out
}

//TempFileName generates a temporary filename for use in testing or whatever
func TempFileName(prefix, suffix string) string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return filepath.Join(os.TempDir(), prefix+hex.EncodeToString(randBytes)+suffix)
}

func buildJsonTrace(trace bytes.Buffer) []byte {
	qtrace := strings.Replace(trace.String(), "\ufffd", "?", -1)
	trace.Reset()
	trace.WriteString(qtrace)
	log.Printf("%v\n", trace.String())
	got, err := ParseReader("", &trace)
	if err != nil {
		log.Fatal(err)
	}
	strace := got.(Ttrace)
	log.Printf("%v\n", strace)
	jtrace, err := json.Marshal(strace)
	if err != nil {
		log.Printf("Cant marshal json\n")
	}
	return jtrace
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
