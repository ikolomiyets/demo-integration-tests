package main

import "time"

type Artifact struct {
	ArtifactName        string  `json:"artifact_name"`
	ArtifactDescription *string `json:"artifact_description,omitempty"`
	Type                string  `json:"type"`
	GroupId             *string `json:"group_id,omitempty"`
	ArtifactId          *string `json:"artifact_id,omitempty"`
}

type Version struct {
	*Artifact
	Version string `json:"version"`
}

type StageArtifacts struct {
	Stage     string    `json:"stage"`
	Artifacts []Version `json:"artifacts"`
}

type Policy struct {
	PolicyNumber string    `json:"policyNumber"`
	StartDate    time.Time `json:"startDate"`
	Insured      []int     `json:"insured"`
}

type PoliciesResponse struct {
	First    int      `json:"first"`
	Next     int      `json:"next"`
	Policies []Policy `json:"policies"`
}

type Customer struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Address   string `json:"address"`
}

type Error struct {
	Message string `json:"message"`
}
