package storage

import (
	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/family"
	"github.com/jmoiron/sqlx"
)

type familyRepository struct {
	db *sqlx.DB
}

func NewFamilyRepository(db *sqlx.DB) *familyRepository {
	return &familyRepository{db: db}
}

func (r *familyRepository) SaveFamily(family *family.Family) (*family.Family, error) {
	query := `
		INSERT INTO families (
			type,
			date_of_entry,
			radius,
			height,
			classification,
			client_id,
			project_id,
			sample_place
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	res, err := r.db.Exec(
		query,
		family.FamilyType,
		family.DateOfEntry,
		family.Radius,
		family.Height,
		family.Classification,
		family.ClientID,
		family.ProjectID,
		family.SamplePlace,
	)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	family.ID = int(id)
	return family, nil
}