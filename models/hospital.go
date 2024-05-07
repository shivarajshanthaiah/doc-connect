package models

import (

	"gorm.io/gorm"
)

type Hospital struct {
	gorm.Model
	Name     string `json:"name"`
	Location string `json:"location"`
	Status   string `json:"status"`
}
