package member

import (
	"time"

	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/user"
)

type Member struct {
    ID             int       `db:"id" json:"id"`
	FamilyID int `db:"family_id" json:"-"`
    Result         *int       `db:"result" json:"result"`
    DateOfFracture *time.Time `db:"date_of_fracture" json:"dateOfFracture"`
    FracturedAt     *time.Time `db:"fractured_at" json:"fracturedAt"`
    Operative      *user.User `db:"-" json:"operative"`
    IsReported     *bool      `db:"is_reported" json:"isReported"`
    OperativeID *int `db:"operative" json:"-"`
}