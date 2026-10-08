package main

import (
	"encoding/json"
	"net/http"
)

type CourseRequest struct {
	CourseID  string `json:"course_id"`
	LearnerID string `json:"learner_id"`
	Deadline  string `json:"deadline"`
}
type Delivery struct {
	CourseID  string `json:"course_id"`
	LearnerID string `json:"learner_id"`
	Status    string `json:"status"`
	Deadline  string `json:"deadline"`
}

func decide(enabled bool, in CourseRequest) Delivery {
	status := "scheduled"
	if !enabled {
		status = "waitlisted"
	}
	return Delivery{in.CourseID, in.LearnerID, status, in.Deadline}
}

func handler(c *Client, w http.ResponseWriter, r *http.Request) {
	var in CourseRequest
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.CourseID == "" || in.LearnerID == "" {
		http.Error(w, "invalid course request", http.StatusBadRequest)
		return
	}
	enabled, err := c.FlagValue("edtech-course-delivery")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	out := decide(enabled, in)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func main() {
	c, err := NewClient()
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/deliver", func(w http.ResponseWriter, r *http.Request) { handler(c, w, r) })
	http.ListenAndServe(":8080", nil)
}
