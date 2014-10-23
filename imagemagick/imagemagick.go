package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"code.google.com/p/go.net/context"

	"github.com/golang/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"
)

const (
	projectID     = "abelana-222"
	outputBucket  = "abelana"
	uploadRetries = 5
	listenAddress = "0.0.0.0:10443"
)

var (
	certPath = flag.String("cert", "cert.pem", "path to the certificate file")
	keyPath  = flag.String("key", "key.pem", "path to the key file")

	client *http.Client

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
	// Initialize API clients
	conf := google.NewComputeEngineConfig("")
	client = &http.Client{Transport: conf.NewTransport()}

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

	start := time.Now()
	err := processImage(cloud.NewContext(projectID, client), bucket, name)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Printf("image processed in %v", time.Since(start))
}

func processImage(ctx context.Context, bucket, name string) error {
	path, err := downloadImage(ctx, bucket, name)
	if err != nil {
		return fmt.Errorf("download: %v", err)
	}

	type partial struct {
		name string
		err  error
	}
	partials := make(chan partial, len(sizes))

	for suffix, size := range sizes {
		go func(suffix, size string) {
			target, ext := name, ""
			if ps := strings.SplitN(target, ".", 2); len(ps) == 2 {
				target, ext = ps[0], "."+ps[1]
			}
			target = fmt.Sprintf("%s_%s%s", target, suffix, ext)

			// convert image
			tmp := filepath.Join(os.TempDir(), target)
			cmd := exec.Command("convert", path, "-adaptive-resize", size, tmp)
			out, err := cmd.CombinedOutput()
			if err != nil {
				partials <- partial{err: fmt.Errorf("convert: %v\n%s", err, out)}
				return
			}
			partials <- partial{name: target}
		}(suffix, size)
	}

	for _ = range sizes {
		p := <-partials
		if p.err != nil {
			return p.err
		}
		err := uploadImage(ctx, outputBucket, p.name, filepath.Join(os.TempDir(), p.name))
		if err != nil {
			return err
		}
	}

	u, _ := url.Parse("https://endpoints-dot-abelana-222.appspot.com/photopush/" + name)
	res, err := client.Do(&http.Request{Method: "PUT", URL: u})
	if err != nil {
		return fmt.Errorf("photo push: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("photo push status: %v", res.Status)
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
