package notice

import (
	"encoding/json"
	"fmt"
	"net/http"

	"appengine"
	"appengine/taskqueue"
)

func init() { http.HandleFunc("/", handler) }

func handler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	// Decode the name and bucket of the notification
	var n struct {
		Name   string `json:"name"`
		Bucket string `json:"bucket"`
	}
	err := json.NewDecoder(r.Body).Decode(&n)
	if err != nil {
		c.Errorf("invalid notification: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Add a new task to the queue
	t := &taskqueue.Task{
		Method:  "PULL",
		Payload: []byte(n.Bucket + "/" + n.Name),
	}
	_, err = taskqueue.Add(c, t, "abelana-in")
	if err != nil {
		c.Errorf("add task to queue: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "OK")
}
