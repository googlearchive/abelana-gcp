package notice

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang/oauth2/google"

	"code.google.com/p/google-api-go-client/compute/v1"

	"appengine"
	"appengine/memcache"
	"appengine/taskqueue"
	"appengine/urlfetch"
)

func init() {
	http.Handle("/", errorHandler(bucketNotificationHandler))
	http.Handle("/notice/incoming-image", errorHandler(incomingImageHandler))
}

func bucketNotificationHandler(w http.ResponseWriter, r *http.Request) error {
	c := appengine.NewContext(r)

	// Decode the name and bucket of the notification
	var n struct {
		Name   string `json:"name"`
		Bucket string `json:"bucket"`
	}
	err := json.NewDecoder(r.Body).Decode(&n)
	if err != nil {
		return fmt.Errorf("invalid notification: %v", err)
	}

	// handle duplicated notifications, drop keys after 5 minutes
	item := &memcache.Item{
		Key:        "notify:" + n.Bucket + "/" + n.Name,
		Expiration: 5 * time.Minute,
		Value:      []byte{},
	}
	switch err := memcache.Add(c, item); err {
	case memcache.ErrNotStored:
		// the key existed already, dup notification
		c.Infof("duplicated notification for %q", item.Key)
		return nil
	case nil:
		c.Infof("first time notification for %q", item.Key)
		// first time we see this key
	default:
		c.Errorf("add notification to memcache: %v", err)
	}

	// Add a new task to the queue
	t := taskqueue.NewPOSTTask("/notice/incoming-image", map[string][]string{
		"bucket": {n.Bucket},
		"name":   {n.Name},
	})
	_, err = taskqueue.Add(c, t, "new-images")
	if err != nil {
		return fmt.Errorf("add task to queue: %v", err)
	}

	fmt.Fprintln(w, "OK")
	return nil
}

func incomingImageHandler(w http.ResponseWriter, r *http.Request) error {
	c := appengine.NewContext(r)

	ip, err := backendIP(c)
	if err != nil {
		return err
	}

	// TODO: add TLS so the backend can authenticate the request.
	// The token sent to the backend is forwarded to the photopush module,
	// we need the userinfo.email scope to be able to verify the token origin.
	config := google.NewAppEngineConfig(c, "https://www.googleapis.com/auth/userinfo.email")
	config.Transport = &urlfetch.Transport{
		Context:                       c,
		Deadline:                      time.Minute,
		AllowInvalidServerCertificate: true,
	}
	client := http.Client{Transport: config.NewTransport()}

	r.ParseForm()
	res, err := client.PostForm(fmt.Sprintf("https://%s:10443", ip), r.Form)
	if err != nil {
		return fmt.Errorf("backend: %v", err)
	}

	w.WriteHeader(res.StatusCode)
	_, err = io.Copy(w, res.Body)

	return err
}

func backendIP(c appengine.Context) (string, error) {
	const key = "backendIP"

	// check if the IP is already in memcache
	item, err := memcache.Get(c, key)
	if err == nil {
		return string(item.Value), nil
	}
	if err != memcache.ErrCacheMiss {
		c.Errorf("get %s from memcache: %v", key, err)
	}

	// create an authenticated compute API client
	config := google.NewAppEngineConfig(c, compute.ComputeReadonlyScope)
	client := &http.Client{Transport: config.NewTransport()}
	svc, err := compute.New(client)
	if err != nil {
		return "", fmt.Errorf("new compute client: %v", err)
	}

	// use the client to obtain the IP with name imagemagick
	addr, err := svc.Addresses.Get("abelana-222", "us-central1", "imagemagick").Do()
	if err != nil {
		return "", fmt.Errorf("get address: %v", err)
	}

	// try to save the IP in memcache
	item = &memcache.Item{
		Key:   key,
		Value: []byte(addr.Address),
		// TODO: how to invalidate this if the address changes?
		Expiration: time.Hour,
	}
	if err := memcache.Set(c, item); err != nil {
		c.Errorf("set %s in memcache: %v", key, err)
	}
	return addr.Address, nil
}

type errorHandler func(http.ResponseWriter, *http.Request) error

func (h errorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		appengine.NewContext(r).Errorf(err.Error())
	}
}
