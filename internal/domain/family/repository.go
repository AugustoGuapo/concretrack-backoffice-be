package family

type Repository interface {
	SaveFamily(family *Family) (*Family, error)
}