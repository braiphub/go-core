package braipsession

import (
	"slices"
	"time"
)

type User struct {
	Hash        string   `json:"hash"`
	Location    string   `json:"location"` // TODO: TO BE IMPLEMENTED
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	Verified    int      `json:"verified"`
	Status      int      `json:"status"`
	Profile     int      `json:"profile"`
	Roles       []Role   `json:"roles"`
	Permissions []string `json:"permissions"`
}

func (u *User) HasRole(role Role) bool {
	return slices.Contains[[]Role, Role](u.Roles, role)
}

func (u *User) UserLocation() *time.Location {
	if u.Location == "" {
		u.Location = "America/Sao_Paulo"
	}

	loc, err := time.LoadLocation(u.Location)
	if err != nil {
		brazilLocation, err := time.LoadLocation("America/Sao_Paulo")
		if err != nil {
			panic(err)
		}

		return brazilLocation
	}

	return loc
}
