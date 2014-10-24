package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"code.google.com/p/go.net/context"
	auth "code.google.com/p/google-api-go-client/oauth2/v2"

	"github.com/campoy/httplog"
	"github.com/golang/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"
)

const (
	projectID     = "abelana-222"
	outputBucket  = "abelana"
	listenAddress = "0.0.0.0:10443"
	pushURL       = "https://endpoints-dot-abelana-222.appspot.com/photopush/"
	authEmail     = "abelana-222@appspot.gserviceaccount.com"
	nConverters   = 25
)

var (
	certPath = flag.String("cert", "cert.pem", "path to the certificate file")
	keyPath  = flag.String("key", "key.pem", "path to the key file")
	debug    = flag.Int("debug", 0, "debug logging level")

	client *http.Client
	ctx    context.Context

	// map with the suffixes and sizes to generate
	sizes = map[string]string{
		"a": "480x800",
		"b": "768x768",
		"c": "1080x1080",
		"d": "1440x1440",
		"e": "1200x1200",
		"f": "1536x1536",
		"g": "720x720",
		"h": "640x640",
		"i": "750x750",
	}

	converterTokens = make(chan bool, nConverters)
)

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	// Initialize API context
	conf := google.NewComputeEngineConfig("")
	var transport http.RoundTripper = conf.NewTransport()
	if *debug > 0 {
		transport = httplog.Transport{transport, *debug > 1, log.Printf}
	}
	client = &http.Client{Transport: transport}
	// NewContext modifies the original client, so passing a new one.
	ctx = cloud.NewContext(projectID, &http.Client{Transport: transport})

	http.HandleFunc("/", handler)
	err := http.ListenAndServeTLS(listenAddress, "cert.pem", "key.pem", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() { log.Printf("handler-time: %v", time.Since(start)) }()

	bucket, name := r.PostFormValue("bucket"), r.PostFormValue("name")
	if bucket == "" || name == "" {
		http.Error(w, "missing bucket or name", http.StatusBadRequest)
		return
	}

	token := r.Header.Get("Authorization")
	if ok, err := authorized(token); !ok {
		if err != nil {
			log.Printf("authorize: %v", err)
		}
		http.Error(w, "you're not authorized", http.StatusForbidden)
		return
	}

	if err := processImage(bucket, name); err != nil {
		// TODO: should this remove uploaded images?
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := notifyDone(name, token); err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func authorized(token string) (ok bool, err error) {
	start := time.Now()
	defer func() { log.Printf("authorized-time: %v %v %v", time.Since(start), ok, err) }()

	if fs := strings.Fields(token); len(fs) == 2 && fs[0] == "Bearer" {
		token = fs[1]
	} else {
		return false, nil
	}

	svc, err := auth.New(http.DefaultClient)
	if err != nil {
		return false, err
	}
	tok, err := svc.Tokeninfo().Access_token(token).Do()
	if err != nil {
		return false, err
	}

	return tok.Email == authEmail, nil
}

func processImage(bucket, name string) (err error) {
	start := time.Now()
	defer func() { log.Printf("process-time: %v %v", time.Since(start), err) }()

	path, err := downloadImage(bucket, name)
	if err != nil {
		return fmt.Errorf("download: %v", err)
	}

	timeout := make(chan bool)
	time.AfterFunc(time.Minute, func() { close(timeout) })

	errc := make(chan error, len(sizes))
	for suffix := range sizes {
		go func(suffix string) {
			target, ext := name, ""
			if sep := strings.LastIndex(target, "."); sep >= 0 {
				target, ext = target[:sep], target[sep:]
			}
			target = fmt.Sprintf("%s_%s%s", target, suffix, ext)

			tmp := filepath.Join(os.TempDir(), target)

			// ask for permission to execute
			if err := convert(tmp, suffix, path, timeout); err != nil {
				errc <- err
			}
			errc <- uploadImage(outputBucket, target, tmp)
			log.Printf("upload-time: %v", time.Since(start))
		}(suffix)
	}

	// wait for all the uploads to finish
	for _ = range sizes {
		if err := <-errc; err != nil {
			return err
		}
	}
	return nil
}

func convert(tmp, suffix, path string, timeout <-chan bool) (err error) {
	start := time.Now()
	defer func() { log.Printf("convert-time: %s %v %v", time.Since(start), suffix, err) }()

	select {
	case converterTokens <- true:
		defer func() { <-converterTokens }()
	case <-timeout:
		return fmt.Errorf("timeout")
	}

	cmd := exec.Command("convert", path, "-adaptive-resize", sizes[suffix], tmp)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("convert: %v\n%s", err, out)
	}
	return nil
}

func downloadImage(bucket, name string) (path string, err error) {
	start := time.Now()
	defer func() { log.Printf("download-time: %v %q %v", time.Since(start), path, err) }()

	r, err := storage.NewReader(ctx, bucket, name)
	if err != nil {
		return "", fmt.Errorf("storage reader: %v", err)
	}
	defer r.Close()

	f, err := ioutil.TempFile("", "")
	if err != nil {
		return "", fmt.Errorf("temp file: %v", err)
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	if err != nil {
		return "", fmt.Errorf("copy storage to temp: %v", err)
	}
	return f.Name(), nil
}

func uploadImage(bucket, name, path string) (err error) {
	start := time.Now()
	defer func() {
		log.Printf("upload-time: %v %q %v", time.Since(start), name, err)
	}()

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %q: %v", path, err)
	}
	defer f.Close()

	w := storage.NewWriter(ctx, bucket, name, nil)
	_, err = io.Copy(w, f)
	if err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("close object writer: %v", err)
	}

	_, err = w.Object()
	return err
}

func notifyDone(name, token string) (err error) {
	start := time.Now()
	defer func() {
		log.Printf("notify-time: %v %q %v", time.Since(start), name, err)
	}()

	req, err := http.NewRequest("POST", pushURL+name, &bytes.Buffer{})
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", token)

	client := http.Client{Transport: http.DefaultTransport}
	if *debug > 0 {
		client.Transport = httplog.Transport{client.Transport, *debug > 1, log.Printf}
	}

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("photo push: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("photo push status: %v", res.Status)
	}
	return nil
}
