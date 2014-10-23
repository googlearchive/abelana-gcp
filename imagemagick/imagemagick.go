package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"code.google.com/p/go.net/context"
	auth "code.google.com/p/google-api-go-client/oauth2/v2"

	"github.com/golang/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"
)

const (
	projectID     = "abelana-222"
	outputBucket  = "abelana"
	uploadRetries = 5
	listenAddress = "0.0.0.0:10443"
	pushURL       = "https://endpoints-dot-abelana-222.appspot.com/photopush/"
	authEmail     = "abelana-222@appspot.gserviceaccount.com"
)

var (
	certPath = flag.String("cert", "cert.pem", "path to the certificate file")
	keyPath  = flag.String("key", "key.pem", "path to the key file")
	debug    = flag.Int("debug", 0, "debug logging level")

	ctx context.Context

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
)

func main() {
	flag.Parse()

	// Initialize API context
	conf := google.NewComputeEngineConfig("")
	var transport http.RoundTripper = conf.NewTransport()
	if *debug > 0 {
		transport = loggingTransport{transport, *debug > 1}
	}
	ctx = cloud.NewContext(projectID, &http.Client{Transport: transport})

	http.HandleFunc("/", handler)
	err := http.ListenAndServeTLS(listenAddress, "cert.pem", "key.pem", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
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

	err := processImage(ctx, bucket, name)
	if err != nil {
		// TODO: should this remove uploaded images?
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if err := notifyDone(name, token); err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func authorized(token string) (bool, error) {
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

func processImage(ctx context.Context, bucket, name string) error {
	path, err := downloadImage(ctx, bucket, name)
	if err != nil {
		return fmt.Errorf("download: %v", err)
	}

	errc := make(chan error, len(sizes))
	for suffix, size := range sizes {
		go func(suffix, size string) {
			target, ext := name, ""
			if ps := strings.SplitN(target, ".", 2); len(ps) == 2 {
				target, ext = ps[0], "."+ps[1]
			}
			target = fmt.Sprintf("%s_%s%s", target, suffix, ext)

			tmp := filepath.Join(os.TempDir(), target)
			cmd := exec.Command("convert", path, "-adaptive-resize", size, tmp)
			out, err := cmd.CombinedOutput()
			if err != nil {
				errc <- fmt.Errorf("convert: %v\n%s", err, out)
				return
			}
			errc <- uploadImage(ctx, outputBucket, name, tmp)
		}(suffix, size)
	}

	// wait for all the uploads to finish
	for _ = range sizes {
		if err := <-errc; err != nil {
			return err
		}
	}

	return nil
}

func downloadImage(ctx context.Context, bucket, name string) (string, error) {
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

func uploadImage(ctx context.Context, bucket, name, path string) error {
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

func notifyDone(name, token string) error {
	req, err := http.NewRequest("POST", pushURL+name, &bytes.Buffer{})
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)

	client := http.Client{Transport: http.DefaultTransport}
	if *debug > 0 {
		client.Transport = loggingTransport{client.Transport, *debug > 1}
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

type loggingTransport struct {
	rt   http.RoundTripper
	body bool
}

func (l loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if b, err := httputil.DumpRequest(req, l.body); err != nil {
		log.Printf("dump request: %v", err)
	} else {
		log.Printf("%s", b)
	}

	res, err := l.rt.RoundTrip(req)
	if err != nil {
		log.Printf("roundtrip error: %v", err)
		return res, err
	}

	if b, err := httputil.DumpResponse(res, l.body); err != nil {
		log.Printf("dump response: %v", err)
	} else {
		log.Printf("%s", b)
	}
	return res, err
}
