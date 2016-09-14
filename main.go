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

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
)

var (
	httpListen  = flag.String("http", "127.0.0.1:5000", "host:port to listen on")
	openBrowser = flag.Bool("openbrowser", true, "open browser automatically")
)

// Msg for communicating with frontend
type Msg struct {
	Grammar    string `json:"grammar"`
	TestString string `json:"test_string"`
}

func main() {
	r := gin.Default()
	m := melody.New()
	m.Config.MaxMessageSize = 1048576

	r.Use(static.Serve("/", static.LocalFile("./public", true)))

	api := r.Group("/api")
	{
		api.GET("/ws", func(c *gin.Context) {
			m.HandleRequest(c.Writer, c.Request)
		})
	}

	m.HandleMessage(func(s *melody.Session, bmsg []byte) {
		defer track(time.Now(), "HandleMessage")

		analyzePEG := func(msg Msg) Ttrace {
			parser,err := generateParser(msg.Grammar)
			if (err != nil) {
				return Ttrace{ Errors: parser.String() }
			}
			out,err := runParser(parser, msg.TestString)
			if (err != nil) {
				return Ttrace{ Errors: out.String() }
			}
			trace,err := analyzeTrace(out)
			if (err != nil) {
				return Ttrace{ Errors: err.Error() }
			}
			trace.Errors = ""
			return trace
		}

		var msg Msg
		if err := json.Unmarshal(bmsg, &msg); err != nil {
			log.Printf("json.Unmarshall: %v", err)
		}
		trace := analyzePEG(msg)
		resp,err := buildResponse(trace)
		if (err != nil) {
			resp = []byte(err.Error())
		}
		m.Broadcast(resp)
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

func generateParser(msg string) (bytes.Buffer, error) {
	var pigout, pigerr bytes.Buffer
	pigeon := exec.Command("pigeon")
	pigeon.Stdin = strings.NewReader(msg)
	pigeon.Stdout = &pigout
	pigeon.Stderr = &pigerr
	err := pigeon.Run()
	if err != nil {
		log.Printf("generateParser pigeon: %v", pigerr.String())
		return *bytes.NewBufferString("generateParser pigeon: "+pigerr.String()), err
	}
	var impin, impout, imperr bytes.Buffer
	goimports := exec.Command("goimports")
	impin.WriteString(mainFunc + pigout.String())
	goimports.Stdin = strings.NewReader(impin.String())
	goimports.Stdout = &impout
	goimports.Stderr = &imperr
	err = goimports.Run()
	if err != nil {
		log.Printf("generateParser goimports: %v", imperr.String())
		return *bytes.NewBufferString("generateParser goimports: "+imperr.String()), err
	}
	return impout, nil
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

func runParser(source bytes.Buffer, test string) (bytes.Buffer, error) {
	tmpfilename := TempFileName("pigeon", ".go")
	err := ioutil.WriteFile(tmpfilename, source.Bytes(), 0644)
	if err != nil {
		log.Printf("runParser(WriteFile): %v", err.Error())
		//runout.Write(runerr.Bytes())
		return *bytes.NewBufferString("runParser(WriteFile): "+err.Error()), err
	}
	defer os.Remove(tmpfilename)

	//log.Printf("go run %v", tmpfilename)
	var runout, runerr bytes.Buffer
	gorun := exec.Command("go", "run", tmpfilename)
	gorun.Stdin = strings.NewReader(test)
	gorun.Stdout = &runout
	gorun.Stderr = &runerr
	_ = gorun.Run()
	/*
	if err != nil {
		log.Printf("GORUN ERROR: %v", runerr.String())
		//runout.Write(runerr.Bytes())
		return *bytes.NewBufferString("GO RUN STDERR: "+runerr.String()), err
	}
	*/
	return runout, nil
}

//TempFileName generates a temporary filename for use in testing or whatever
func TempFileName(prefix, suffix string) string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return filepath.Join(os.TempDir(), prefix+hex.EncodeToString(randBytes)+suffix)
}
func analyzeTrace(b bytes.Buffer) (Ttrace, error) {
	s := strings.Replace(b.String(), "\ufffd", "?", -1) // fix pigeon bug
	b.Reset()
	b.WriteString(s)
	//log.Printf("%v\n", trace.String())
	got, err := ParseReader("", &b)
	if err != nil {
		log.Fatal(err)
		return Ttrace{}, err
	}
	trace := got.(Ttrace)
	ftrace := filterTrace(trace)
	return ftrace, nil
}

func filterTrace(t Ttrace) Ttrace {
	maxIdx := 0
	var walk  func(e Tentry) []Tentry;
	walk = func(e Tentry) []Tentry {
		res := []Tentry{}
		fcalls := []Tentry{}
		for _, v := range e.Calls {
			g := walk(v)
			if len(g) != 0 {
				fcalls = append(fcalls, g...)
			}
		}
		if  ( strings.HasPrefix(e.Detail.Name, "Rule ") ||  
			strings.HasPrefix(e.Detail.Name, "ZeroOrOneExpr ") ||
			strings.HasPrefix(e.Detail.Name, "OneOrMoreExpr ") ||
			strings.HasPrefix(e.Detail.Name, "ZeroOrMoreExpr ") ) {
			e.Calls = fcalls
			if e.Match != nil {
				maxIdx = max(maxIdx, e.Match.Pos.Pos)
			}
			res = append(res, e)
		} else {
			res = fcalls
		}
		return res
	}

	r := []Tentry{}
	for _, v := range t.Entries {
		g := walk(v)
		r = append(r, g...)
	}
	return Ttrace{Entries: r, MaxIdx: maxIdx}
}

func buildResponse(trace Ttrace) ([]byte, error) {
	//log.Printf("%v\n", ftrace)
	jtrace, err := json.Marshal(trace)
	if err != nil {
		log.Printf("buildResponse: Cant marshal json\n")
		return bytes.NewBufferString("BUILD JSON: Cant marshal json").Bytes(), err
	}
	return jtrace, nil
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

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func track(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
