package member

import (
	"time"

	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/user"
)

type Member struct {
	ID             int        `db:"id" json:"id"`
	FamilyID       int        `db:"family_id" json:"family_id"`
	Result         *float64   `db:"result" json:"result"`
	DateOfFracture *time.Time `db:"date_of_fracture" json:"date_of_fracture"`
	FracturedAt    *time.Time `db:"fractured_at" json:"fractured_at"`
	Operative      *user.User `db:"-" json:"operative"`
	IsReported     *bool      `db:"is_reported" json:"is_reported"`
	FractureDays   *int       `db:"fracture_days" json:"fracture_days"`
	OperativeID    *int       `db:"operative" json:"-"`
	FractureType   *string    `db:"fracture_type" json:"fracture_type"`
}
