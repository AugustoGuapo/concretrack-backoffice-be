// Package project contains all the logic and domain for the project usecases
package project

import (
	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/client"
	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/family"
)

type Project struct {
    ID       int             `db:"id" json:"id"`
    Name     string          `db:"name" json:"name"`
	ClientID int 			 `db:"client_id" json:"-"`
    Client   client.Client   `db:"-" json:"client"`
    Families []family.Family `db:"-" json:"families"`
}
