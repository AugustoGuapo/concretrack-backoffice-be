package project

type Repository interface {
	GetProjects(page int) ([]*Project, error)
	GetProjectByID(ID int)(*Project, error)
}