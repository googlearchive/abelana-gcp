package notice

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/golang/oauth2/google"

	"appengine"
	"appengine/memcache"
	"appengine/taskqueue"
	"appengine/urlfetch"
)

const backendAddress = "http://146.148.74.233:8080"

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
	config := google.NewAppEngineConfig(c, "https://www.googleapis.com/auth/userinfo.email")
	config.Transport = &urlfetch.Transport{
		Context:                       c,
		Deadline:                      time.Minute,
		AllowInvalidServerCertificate: true,
	}
	client := http.Client{Transport: config.NewTransport()}

	r.ParseForm()
	res, err := client.PostForm(backendAddress, r.Form)
	if err != nil {
		return fmt.Errorf("backend: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		b, err := httputil.DumpResponse(res, true)
		if err != nil {
			return fmt.Errorf("dump response: %v", err)
		}
		c.Errorf("backend failed with code %v:\n%s", res.Status, b)
	}

	w.WriteHeader(res.StatusCode)
	_, err = io.Copy(w, res.Body)
	return err
}

type errorHandler func(http.ResponseWriter, *http.Request) error

func (h errorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		appengine.NewContext(r).Errorf(err.Error())
	}
}
