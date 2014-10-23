package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"strings"

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

	ctx := cloud.NewContext(projectID, client)
	err := processImage(ctx, bucket, name)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func processImage(ctx context.Context, bucket, name string) error {
	path, err := downloadImage(ctx, bucket, name)
	if err != nil {
		return fmt.Errorf("download: %v", err)
	}

	for suffix, size := range sizes {
		go func(suffix, size string) {
			log.Println("process size", size)
			// convert image
			tmp, err := ioutil.TempFile("", "")
			if err != nil {
				log.Printf("create target file: %v", err)
				return
			}
			tmp.Close()

			cmd := exec.Command("convert", path, "-adaptive-resize", size, tmp.Name())
			out, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("convert: %v\n%s", err, out)
			}

			// compute target name
			target, ext := name, ""
			if ps := strings.SplitN(target, ".", 2); len(ps) == 2 {
				target, ext = ps[0], "."+ps[1]
			}
			target = fmt.Sprintf("%s_%s_%s%s", target, size, suffix, ext)

			err = retry(uploadRetries, func() error {
				return uploadImage(ctx, outputBucket, target, tmp.Name())
			})
			if err != nil {
				log.Printf("upload %v failed %v times: %v", target, uploadRetries, err)
			}
		}(suffix, size)
	}

	return nil
}

func downloadImage(ctx context.Context, bucket, name string) (string, error) {
	log.Println("download", bucket, name)
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
	log.Println("upload", bucket, name)
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

	// block until the write operation is done
	_, err = w.Object()
	return err
}

func decodePayload(s string) (bucket, name string, err error) {
	p, err := ioutil.ReadAll(
		base64.NewDecoder(base64.StdEncoding,
			strings.NewReader(s)))
	if err != nil {
		return "", "", err
	}

	ps := strings.SplitN(string(p), "/", 2)
	if len(ps) != 2 {
		return "", "", fmt.Errorf("invalid name format: %q", p)
	}
	return ps[0], ps[1], nil
}

// Utility stuff

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

func retry(n int, f func() error) error {
	err := f()
	for i := 1; err != nil && i < n; i++ {
		err = f()
	}
	return err
}
