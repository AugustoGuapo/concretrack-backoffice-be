package family

import (
	"time"

	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/member"
)

type Family struct {
    ID             int            `db:"id" json:"id"`
    FamilyType     string         `db:"type" json:"familyType"`
    DateOfEntry    time.Time      `db:"date_of_entry" json:"dateOfEntry"`
    Radius         float64        `db:"radius" json:"radius"`
    Height         float64        `db:"height" json:"height"`
    Classification float64        `db:"classification" json:"classification"`
    ClientID       int            `db:"client_id" json:"clientId"`
    ProjectID      int            `db:"project_id" json:"projectId"`
    Members        []member.Member `db:"-" json:"members"`
}
