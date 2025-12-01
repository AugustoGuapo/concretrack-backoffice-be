package project

type Repository interface {
	GetProjects(page int) ([]*Project, error)
	GetProjectByID(ID int) (*Project, error)
	SaveProject(project *Project) (*Project, error)
	GetProjectsByClientID(clientID int) ([]*Project, error)
}
