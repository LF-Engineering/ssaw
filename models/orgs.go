package orgs

import "time"

// Org - single organization name and last modified date
type Org struct {
	Name         string    `json:"name"`
	LastModified time.Time `json:"last_modified"`
}

// Orgs - array of Org
type Orgs struct {
	Orgs []Org `json:"orgs"`
}
