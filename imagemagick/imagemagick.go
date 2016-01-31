package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"runtime"
	"strings"

	"github.com/gographics/imagick/imagick"
	auth "github.com/google/google-api-go-client/oauth2/v2"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"
)

const (
	projectID     = "abelana-222"
	outputBucket  = "abelana"
	listenAddress = "0.0.0.0:8080"
	pushURL       = "https://endpoints-dot-abelana-222.appspot.com/photopush/"
	authEmail     = "abelana-222@appspot.gserviceaccount.com"
)

var (
	// map with the suffixes and sizes to generate
	sizes = map[string]struct{ x, y uint }{
		"a": {480, 800},
		"b": {768, 768},
		"c": {1080, 1080},
		"d": {1440, 1440},
		"e": {1200, 1200},
		"f": {1536, 1536},
		"g": {720, 720},
		"h": {640, 640},
		"i": {750, 750},
	}

	ctx    context.Context
	client *http.Client
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	client = &http.Client{Transport: &oauth2.Transport{
		Source: google.ComputeTokenSource(""),
	}}

	ctx = cloud.NewContext(projectID, client)

	http.HandleFunc("/healthcheck", func(http.ResponseWriter, *http.Request) {})
	http.HandleFunc("/", notificationHandler)
	log.Println("server listening on", listenAddress)

	if err := http.ListenAndServe(listenAddress, nil); err != nil {
		log.Fatal(err)
	}
}

func notificationHandler(w http.ResponseWriter, r *http.Request) {
	bucket, name := r.PostFormValue("bucket"), r.PostFormValue("name")
	if bucket == "" || name == "" {
		http.Error(w, "missing bucket or name", http.StatusBadRequest)
		return
	}

	if ok, err := authorized(r.Header.Get("Authorization")); !ok {
		if err != nil {
			log.Printf("authorize: %v", err)
		}
		http.Error(w, "you're not authorized", http.StatusForbidden)
		return
	}

	start := time.Now()
	defer func() { log.Printf("%v: processed in %v", name, time.Since(start)) }()

	if err := processImage(bucket, name); err != nil {
		// TODO: should this remove uploaded images?
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := notifyDone(name); err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func authorized(token string) (ok bool, err error) {
	if fs := strings.Fields(token); len(fs) == 2 && fs[0] == "Bearer" {
		token = fs[1]
	} else {
		return false, nil
	}

	svc, err := auth.New(http.DefaultClient)
	if err != nil {
		return false, err
	}
	tok, err := svc.Tokeninfo().AccessToken(token).Do()
	return err == nil && tok.Email == authEmail, err
}

func processImage(bucket, name string) error {
	r, err := storage.NewReader(ctx, bucket, name)
	if err != nil {
		return fmt.Errorf("storage reader: %v", err)
	}
	img, err := ioutil.ReadAll(r)
	r.Close()
	if err != nil {
		return fmt.Errorf("read image: %v", err)
	}

	wand := imagick.NewMagickWand()
	defer wand.Destroy()

	wand.ReadImageBlob(img)
	if err := wand.SetImageFormat("WEBP"); err != nil {
		return fmt.Errorf("set WEBP format: %v", err)
	}

	errc := make(chan error, len(sizes))
	for suffix, size := range sizes {
		go func(wand *imagick.MagickWand, suffix string, x, y uint) {
			errc <- func() error {
				defer wand.Destroy()

				if err := wand.AdaptiveResizeImage(size.x, size.y); err != nil {
					return fmt.Errorf("resize: %v", err)
				}

				target := name
				if sep := strings.LastIndex(target, "."); sep >= 0 {
					target = target[:sep]
				}
				target = fmt.Sprintf("%s_%s.webp", target, suffix)

				w := storage.NewWriter(ctx, outputBucket, target)
				if _, err := w.Write(wand.GetImageBlob()); err != nil {
					return fmt.Errorf("new writer: %v", err)
				}
				if err := w.Close(); err != nil {
					return fmt.Errorf("close object writer: %v", err)
				}
				return nil
			}()
		}(wand.Clone(), suffix, size.x, size.y)
	}

	for _ = range sizes {
		if err := <-errc; err != nil {
			return err
		}
	}
	return nil
}

func notifyDone(name string) (err error) {
	req, err := http.NewRequest("POST", pushURL+name, &bytes.Buffer{})
	if err != nil {
		return err
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
