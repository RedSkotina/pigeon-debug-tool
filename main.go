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
	m.Config.MaxMessageSize = 10048576

	r.Use(static.Serve("/", static.LocalFile("./public", true)))

	api := r.Group("/api")
	{
		api.GET("/ws", func(c *gin.Context) {
			m.HandleRequest(c.Writer, c.Request)
		})
	}

	m.HandleMessage(func(s *melody.Session, bmsg []byte) {
		defer track(time.Now(), "HandleMessage")

		analyzePEG := func(msg Msg) TTrace {
			parser, err := generateParser(msg.Grammar)
			if err != nil {
				return TTrace{Errors: parser.String()}
			}
			out, err := runParser(parser, msg.TestString)
			if err != nil {
				return TTrace{Errors: out.String()}
			}
			trace, err := analyzeTrace(out)
			if err != nil {
				return TTrace{Errors: err.Error()}
			}
			trace.Errors = ""
			return trace
		}

		var msg Msg
		if err := json.Unmarshal(bmsg, &msg); err != nil {
			log.Printf("json.Unmarshall: %v", err)
		}
		trace := analyzePEG(msg)
		resp, err := buildResponse(trace)
		if err != nil {
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
		return *bytes.NewBufferString("generateParser pigeon: " + pigerr.String()), err
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
		return *bytes.NewBufferString("generateParser goimports: " + imperr.String()), err
	}
	return impout, nil
}

const mainFunc = `
func main() {
	_, err := ParseReader("stdin", os.Stdin, Debug(true))
	if err != nil {
		log.Fatal(err)
	}
}
`

func runParser(source bytes.Buffer, test string) (bytes.Buffer, error) {
	tmpfilename := TempFileName("pigeon", ".go")
	err := ioutil.WriteFile(tmpfilename, source.Bytes(), 0644)
	if err != nil {
		log.Printf("runParser(WriteFile): %v", err.Error())
		//runout.Write(runerr.Bytes())
		return *bytes.NewBufferString("runParser(WriteFile): " + err.Error()), err
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

func analyzeTrace(b bytes.Buffer) (TTrace, error) {
	s := strings.Replace(b.String(), "\ufffd", "?", -1) // fix pigeon bug
	b.Reset()
	b.WriteString(s)
	//log.Printf("%v\n", trace.String())
	got, err := ParseReader("", &b)
	if err != nil {
		log.Fatal(err)
		return TTrace{}, err
	}
	trace := got.(TTrace)
	ftrace := filterTrace(trace)
	vtrace := virtualIdxTrace(ftrace)
	return vtrace, nil
}

func filterTrace(t TTrace) TTrace {
	var walk func(e TEntry) []TEntry
	walk = func(e TEntry) []TEntry {
		res := []TEntry{}
		fcalls := []TEntry{}
		for _, v := range e.Calls {
			g := walk(v)
			if len(g) != 0 {
				fcalls = append(fcalls, g...)
			}
		}
		if strings.HasPrefix(e.Detail.Name, "Rule ") ||
			strings.HasPrefix(e.Detail.Name, "ZeroOrOneExpr ") ||
			strings.HasPrefix(e.Detail.Name, "OneOrMoreExpr ") ||
			strings.HasPrefix(e.Detail.Name, "ZeroOrMoreExpr ") {
			e.Calls = fcalls
			res = append(res, e)
		} else {
			res = fcalls
		}
		return res
	}

	res := []TEntry{}
	for _, v := range t.Entries {
		g := walk(v)
		res = append(res, g...)
	}
	return TTrace{Entries: res}
}

func virtualIdxTrace(t TTrace) TTrace {
	var walk func(e TEntry) (int, []TEntry)
	walk = func(e TEntry) (int, []TEntry) {
		res := []TEntry{}
		vIdx := 0
		fcalls := []TEntry{}
		for _, v := range e.Calls {
			d, g := walk(v)
			vIdx = max(vIdx, d)
			if len(g) != 0 {
				fcalls = append(fcalls, g...)
			}
		}
		if e.IsMatch {
			vIdx = e.Detail.Idx2 // matched rules have real idx2
		} else {
			e.Detail.Idx2 = vIdx // assign virtual idx2 to non matched rules
		}
		e.Calls = fcalls
		res = append(res, e)
		return vIdx, res
	}

	res := []TEntry{}
	for _, v := range t.Entries {
		_, g := walk(v)
		res = append(res, g...)
	}
	return TTrace{Entries: res}
}

func buildResponse(trace TTrace) ([]byte, error) {
	jtrace, err := json.Marshal(trace)
	if err != nil {
		log.Printf("buildResponse: Cant marshal json\n")
		return bytes.NewBufferString("BUILD JSON: Cant marshal json").Bytes(), err
	}
	//log.Printf("%s\n", jtrace)

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

//TempFileName generates a temporary filename for use in testing or whatever
func TempFileName(prefix, suffix string) string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return filepath.Join(os.TempDir(), prefix+hex.EncodeToString(randBytes)+suffix)
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
