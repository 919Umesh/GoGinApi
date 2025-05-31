package models

type Venue struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Size     string `json:"size"`
	Image    string `json:"image"`
}
