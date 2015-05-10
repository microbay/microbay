package model

import (
	"github.com/SHMEDIALIMITED/apigo/server/backends"
)

type Resource struct {
	Auth     string            `json:"auth"`
	Path     string            `json:"path"`
	Micros   []Micro           `json:"micros"`
	Backends backends.Backends // Load balancer
}
