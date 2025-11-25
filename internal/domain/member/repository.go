package member

type Repository interface {
	SaveMembers([]*Member) ([]*Member, error)
}