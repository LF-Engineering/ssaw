package affiliation

import "time"

// Org - single organization name and last modified date
type Org struct {
	Name         string    `json:"name"`
	LastModified time.Time `json:"last_modified"`
}

// Profile - single add/update data
type Profile struct {
	UUID         string    `json:"uuid"`
	LastModified time.Time `json:"last_modified"`
	Op           string    `json:"op"`
	// FIXME
}

// Data - holds organizations and profiles to add/update
type Data struct {
	Orgs     []Org     `json:"orgs"`
	Profiles []Profile `json:"profiles"`
}
