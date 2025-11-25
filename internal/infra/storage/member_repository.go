package storage

import (
	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/member"
	"github.com/jmoiron/sqlx"
)

type MemberRepository struct {
	db *sqlx.DB
}

func NewMemberRepository(db *sqlx.DB) *MemberRepository {
	return &MemberRepository{db: db}
}

func (r *MemberRepository) SaveMembers(members []*member.Member) ([]*member.Member, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO members (
			family_id,
			result,
			date_of_fracture,
			fractured_at,
			is_reported,
			fracture_days,
			operative
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	for _, m := range members {
		res, err := tx.Exec(
			query,
			m.FamilyID,
			m.Result,
			m.DateOfFracture,
			m.FracturedAt,
			m.IsReported,
			m.FractureDays,
			m.OperativeID,
		)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		id, err := res.LastInsertId()
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		m.ID = int(id)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return members, nil
}
