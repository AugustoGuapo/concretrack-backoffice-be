package family

import (
	"time"

	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/member"
)

type Family struct {
    ID             int            `db:"id" json:"id"`
    FamilyType     string         `db:"type" json:"family_type"`
    SamplePlace string `db:"sample_place" json:"sample_place"`
    DateOfEntry    time.Time      `db:"date_of_entry" json:"date_of_entry"`
    Radius         float64        `db:"radius" json:"radius"`
    Height         float64        `db:"height" json:"height"`
    Classification float64        `db:"classification" json:"classification"`
    ClientID       int            `db:"client_id" json:"client_id"`
    ProjectID      int            `db:"project_id" json:"project_id"`
    Members        []member.Member `db:"-" json:"members"`
}
