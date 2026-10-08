package main

import "testing"

func TestDecideCourseDelivery(t *testing.T) {
	in := CourseRequest{CourseID: "math-101", LearnerID: "learner-7", Deadline: "2026-10-01"}
	tests := []struct {
		name    string
		enabled bool
		want    string
	}{{"enabled", true, "scheduled"}, {"disabled", false, "waitlisted"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decide(tt.enabled, in).Status; got != tt.want {
				t.Fatalf("status=%q want %q", got, tt.want)
			}
		})
	}
}
