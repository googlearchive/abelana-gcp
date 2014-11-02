package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"runtime"
	"strings"

	"code.google.com/p/go.net/context"
	auth "code.google.com/p/google-api-go-client/oauth2/v2"

	"github.com/gographics/imagick/imagick"
	"github.com/golang/oauth2/google"
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
	account = flag.String("account", "service-account.json", "path to service account JSON file")

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
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	ctx = cloud.NewContext(projectID, &http.Client{
		Transport: google.NewComputeEngineConfig("").NewTransport(),
	})

	config, err := google.NewServiceAccountJSONConfig(*account, "https://www.googleapis.com/auth/userinfo.email")
	if err != nil {
		log.Fatal(err)
	}
	client = &http.Client{Transport: config.NewTransport()}

	http.HandleFunc("/healthcheck", func(http.ResponseWriter, *http.Request) {})
	http.HandleFunc("/", notificationHandler)
	log.Println("server about to start listening on", listenAddress)
	err = http.ListenAndServe(listenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func notificationHandler(w http.ResponseWriter, r *http.Request) {
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

	start := time.Now()
	defer func() { log.Printf("done in %v", time.Since(start)) }()

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
	wand.SetGravity(imagick.GRAVITY_CENTER)

	errc := make(chan error, len(sizes))
	for suffix, size := range sizes {
		go func(wand *imagick.MagickWand, suffix string, x, y uint) {
			errc <- func() error {
				defer wand.Destroy()

				/* Fix later, crop the images to be square.
				if err := wand.CropImage(x, y, 0, 0); err != nil {
					return fmt.Errorf("crop: %v", err)
				}
				*/
				if err := wand.AdaptiveResizeImage(size.x, size.y); err != nil {
					return fmt.Errorf("resize: %v", err)
				}

				target := name
				if sep := strings.LastIndex(target, "."); sep >= 0 {
					target = target[:sep]
				}
				target = fmt.Sprintf("%s_%s.webp", target, suffix)

				w := storage.NewWriter(ctx, outputBucket, target, nil)
				if _, err := w.Write(wand.GetImageBlob()); err != nil {
					return fmt.Errorf("new writer: %v", err)
				}
				if err := w.Close(); err != nil {
					return fmt.Errorf("close object writer: %v", err)
				}
				if _, err := w.Object(); err != nil {
					return fmt.Errorf("write op: %v", err)
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
	tok, err := svc.Tokeninfo().Access_token(token).Do()
	return err == nil && tok.Email == authEmail, err
}

func notifyDone(name, token string) (err error) {
	// drop the file extension
	name = name[:strings.LastIndex(name, ".")]
	req, err := http.NewRequest("POST", pushURL+name, &bytes.Buffer{})
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("photo push: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("photo push status: %v", res.Status)
	}
	return nil
}
