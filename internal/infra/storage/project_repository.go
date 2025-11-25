package storage

import (
	"errors"
	"log"

	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/client"
	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/family"
	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/member"
	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/project"
	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/user"
	"github.com/jmoiron/sqlx"
)

var PageSize int = 20

type projectRepository struct {
	db *sqlx.DB
}

func NewProjectRepository(db *sqlx.DB) project.Repository {
	return &projectRepository{db: db}
}

func (p *projectRepository) GetProjectByID(ID int) (*project.Project, error) {
    log.Printf("[GetProjectByID] Starting. ID=%d", ID)

    // -----------------------------
    // 1. Project
    // -----------------------------
    projectRow := p.db.QueryRowx("SELECT id, name, client_id FROM projects WHERE id = ?", ID)
    project := &project.Project{}
    if err := projectRow.StructScan(project); err != nil {
        log.Printf("[GetProjectByID] Project not found or scan failed. err=%v", err)
        return nil, err
    }
    log.Printf("[GetProjectByID] Project loaded: %+v", project)

    // -----------------------------
    // 2. Client
    // -----------------------------
    clientRow := p.db.QueryRowx("SELECT id, name FROM clients WHERE id = ?", project.ClientID)
    client := &client.Client{}
    if err := clientRow.StructScan(client); err != nil {
        log.Printf("[GetProjectByID] Failed loading client %d. err=%v", project.ClientID, err)
        return nil, err
    }
    project.Client = *client
    log.Printf("[GetProjectByID] Client loaded: %+v", client)

    // -----------------------------
    // 3. Families
    // -----------------------------
    var families []family.Family
    if err := p.db.Select(&families, `
        SELECT id, type, date_of_entry, radius, height, classification, client_id, project_id
        FROM families
        WHERE project_id = ?`, project.ID); err != nil {
        log.Printf("[GetProjectByID] Failed loading families. err=%v", err)
        return nil, err
    }
    log.Printf("[GetProjectByID] Families found=%d", len(families))

    project.Families = families

    if len(families) == 0 {
        log.Printf("[GetProjectByID] No families for this project. Returning early.")
        return project, nil
    }

    // Map families by ID
    familyIDs := make([]int, len(families))
    familyMap := make(map[int]*family.Family)
    for i := range families {
        familyIDs[i] = families[i].ID
        familyMap[families[i].ID] = &families[i]
    }
    log.Printf("[GetProjectByID] Family IDs: %v", familyIDs)

    // -----------------------------
    // 4. Members
    // -----------------------------
    query, args, _ := sqlx.In(`
        SELECT id, family_id, result, date_of_fracture, fractured_at, is_reported, operative
        FROM members
        WHERE family_id IN (?)`, familyIDs)
    query = p.db.Rebind(query)

    var members []member.Member
    if err := p.db.Select(&members, query, args...); err != nil {
        log.Printf("[GetProjectByID] Failed loading members. err=%v", err)
        return nil, err
    }
    log.Printf("[GetProjectByID] Members found=%d", len(members))

    if len(members) == 0 {
        project.Families = families
        return project, nil
    }

    // -----------------------------
    // 5. Load Operatives (Batch)
    // -----------------------------
    operativeIDs := make([]int, 0)
    seen := make(map[int]bool)

    for _, m := range members {
        if m.OperativeID != nil && *m.OperativeID != 0 && !seen[*m.OperativeID] {
            operativeIDs = append(operativeIDs, *m.OperativeID)
            seen[*m.OperativeID] = true
        }
    }

    var operativeMap = map[int]*user.User{}

    if len(operativeIDs) > 0 {
        log.Printf("[GetProjectByID] Loading operatives: %v", operativeIDs)

        qOps, argsOps, _ := sqlx.In(`
            SELECT id, first_name, last_name, is_active
            FROM users
            WHERE id IN (?)`, operativeIDs)
        qOps = p.db.Rebind(qOps)

        var ops []user.User
        if err := p.db.Select(&ops, qOps, argsOps...); err != nil {
            log.Printf("[GetProjectByID] Failed loading operatives. err=%v", err)
            return nil, err
        }

        operativeMap = make(map[int]*user.User)
        for i := range ops {
            u := ops[i]
            operativeMap[u.ID] = &u
        }

        log.Printf("[GetProjectByID] Loaded %d operatives", len(operativeMap))
    }

    // -----------------------------
    // 6. Attach Members + Operatives to Families
    // -----------------------------
    for i := range members {
        if members[i].OperativeID != nil {
            op := operativeMap[*members[i].OperativeID]
            members[i].Operative = op
            familyMap[members[i].FamilyID].Members = append(familyMap[members[i].FamilyID].Members, members[i])
        }

    }

    project.Families = families
    log.Printf("[GetProjectByID] Completed assembly of project %d", project.ID)

    return project, nil
}




func (p *projectRepository) GetProjects(page int) ([]*project.Project, error) {
    log.Printf("[GetProjects] Starting. page=%d", page)

    if page < 1 {
        log.Printf("[GetProjects] Invalid page number: %d", page)
        return nil, errors.New("page can't be less than one")
    }

    offset := (page - 1) * PageSize
    log.Printf("[GetProjects] PageSize=%d offset=%d", PageSize, offset)

    // ---------------------------
    // 1. Proyectos base
    // ---------------------------
    var projects []project.Project
    err := p.db.Select(&projects, `
        SELECT id, name, client_id
        FROM projects
        ORDER BY id
        LIMIT ? OFFSET ?`,
        PageSize, offset)
    if err != nil {
        log.Printf("[GetProjects] Failed loading projects. err=%v", err)
        return nil, err
    }
    log.Printf("[GetProjects] Projects found=%d", len(projects))

    if len(projects) == 0 {
        log.Printf("[GetProjects] No projects in this page.")
        return []*project.Project{}, nil
    }

    projMap := make(map[int]*project.Project)
    projectIDs := make([]int, 0, len(projects))

    for i := range projects {
        proj := &projects[i]
        projMap[proj.ID] = proj
        projectIDs = append(projectIDs, proj.ID)
    }
    log.Printf("[GetProjects] Project IDs=%v", projectIDs)

    // ---------------------------
    // 2. Clientes
    // ---------------------------
    clientIDs := make([]int, 0, len(projects))
    for _, p := range projects {
        clientIDs = append(clientIDs, p.ClientID)
    }
    log.Printf("[GetProjects] Client IDs=%v", clientIDs)

    query, args, err := sqlx.In(`
        SELECT id, name
        FROM clients
        WHERE id IN (?)`, clientIDs)
    if err != nil {
        log.Printf("[GetProjects] sqlx.In error clients. err=%v", err)
        return nil, err
    }
    query = p.db.Rebind(query)

    var clients []client.Client
    if err := p.db.Select(&clients, query, args...); err != nil {
        log.Printf("[GetProjects] Failed loading clients. err=%v", err)
        return nil, err
    }
    log.Printf("[GetProjects] Clients found=%d", len(clients))

    cliMap := make(map[int]*client.Client)
    for i := range clients {
        c := &clients[i]
        cliMap[c.ID] = c
    }

    for _, p := range projMap {
        p.Client = *cliMap[p.ClientID]
    }

    // ---------------------------
    // 3. Familias
    // ---------------------------
    query, args, err = sqlx.In(`
        SELECT id, type, date_of_entry, radius, height, classification, client_id, project_id
        FROM families
        WHERE project_id IN (?)`, projectIDs)
    if err != nil {
        log.Printf("[GetProjects] sqlx.In error families. err=%v", err)
        return nil, err
    }
    query = p.db.Rebind(query)

    var families []family.Family
    if err := p.db.Select(&families, query, args...); err != nil {
        log.Printf("[GetProjects] Failed loading families. err=%v", err)
        return nil, err
    }
    log.Printf("[GetProjects] Families found=%d", len(families))

    famMap := make(map[int]*family.Family)
    familyIDs := make([]int, 0, len(families))

    for i := range families {
        f := &families[i]
        famMap[f.ID] = f
        familyIDs = append(familyIDs, f.ID)

        projMap[f.ProjectID].Families = append(projMap[f.ProjectID].Families, *f)
    }
    log.Printf("[GetProjects] Family IDs=%v", familyIDs)

    if len(families) == 0 {
        log.Printf("[GetProjects] No families. Returning early.")
        out := make([]*project.Project, 0, len(projects))
        for i := range projects {
            out = append(out, &projects[i])
        }
        return out, nil
    }

    // ---------------------------
    // 4. Miembros
    // ---------------------------
    query, args, err = sqlx.In(`
        SELECT id, family_id, result, date_of_fracture, is_reported
        FROM members
        WHERE family_id IN (?)`, familyIDs)
    if err != nil {
        log.Printf("[GetProjects] sqlx.In error members. err=%v", err)
        return nil, err
    }
    query = p.db.Rebind(query)

    var members []member.Member
    if err := p.db.Select(&members, query, args...); err != nil {
        log.Printf("[GetProjects] Failed loading members. err=%v", err)
        return nil, err
    }
    log.Printf("[GetProjects] Members found=%d", len(members))

    for _, m := range members {
        famMap[m.FamilyID].Members = append(famMap[m.FamilyID].Members, m)
    }

    for projID := range projMap {
        for i := range projMap[projID].Families {
            famID := projMap[projID].Families[i].ID
            projMap[projID].Families[i] = *famMap[famID]
        }
    }

    // ---------------------------
    // Final
    // ---------------------------
    out := make([]*project.Project, 0, len(projects))
    for i := range projects {
        out = append(out, &projects[i])
    }
    log.Printf("[GetProjects] Returning %d projects", len(out))

    return out, nil
}

func (r *projectRepository) SaveProject(p *project.Project) (*project.Project, error) {
    result, err := r.db.Exec(`
        INSERT INTO projects (name, client_id)
        VALUES (?, ?)
    `, p.Name, p.ClientID)
    log.Printf("%+v", p)

    if err != nil {
        return nil, err
    }

    id, err := result.LastInsertId()
    log.Print(id)
    if err != nil {
        return nil, err
    }

    created, err := r.GetProjectByID(int(id))

    if err != nil {
        return nil, err
    }

    return created, nil
}

func (p *projectRepository) GetProjectsByClientID(clientID int) ([]*project.Project, error) {
    log.Printf("[GetProjectsByClientID] Starting. clientID=%d", clientID)

    if clientID < 1 {
        log.Printf("[GetProjectsByClientID] Invalid clientID=%d", clientID)
        return nil, errors.New("clientID must be greater than zero")
    }

    // ---------------------------
    // 1. Proyectos base
    // ---------------------------
    var projects []project.Project
    err := p.db.Select(&projects, `
        SELECT id, name, client_id
        FROM projects
        WHERE client_id = ?
        ORDER BY id`, clientID)

    if err != nil {
        log.Printf("[GetProjectsByClientID] Failed loading projects. err=%v", err)
        return nil, err
    }

    log.Printf("[GetProjectsByClientID] Projects found=%d", len(projects))

    if len(projects) == 0 {
        return []*project.Project{}, nil
    }

    projMap := make(map[int]*project.Project)
    projectIDs := make([]int, 0, len(projects))

    for i := range projects {
        proj := &projects[i]
        projMap[proj.ID] = proj
        projectIDs = append(projectIDs, proj.ID)
    }

    // ---------------------------
    // 2. Cargar el cliente
    // ---------------------------
    var cli client.Client
    err = p.db.Get(&cli, `SELECT id, name FROM clients WHERE id = ?`, clientID)
    if err != nil {
        log.Printf("[GetProjectsByClientID] Failed loading client. err=%v", err)
        return nil, err
    }

    // Asignar cliente a cada proyecto
    for _, p := range projMap {
        p.Client = cli
    }

    // ---------------------------
    // 3. Familias
    // ---------------------------
    query, args, err := sqlx.In(`
        SELECT id, type, date_of_entry, radius, height, classification, client_id, project_id, sample_place
        FROM families
        WHERE project_id IN (?)`, projectIDs)
    if err != nil {
        log.Printf("[GetProjectsByClientID] sqlx.In error families. err=%v", err)
        return nil, err
    }

    query = p.db.Rebind(query)

    var families []family.Family
    if err := p.db.Select(&families, query, args...); err != nil {
        log.Printf("[GetProjectsByClientID] Failed loading families. err=%v", err)
        return nil, err
    }

    log.Printf("[GetProjectsByClientID] Families found=%d", len(families))

    famMap := make(map[int]*family.Family)
    familyIDs := make([]int, 0, len(families))

    for i := range families {
        f := &families[i]
        famMap[f.ID] = f
        familyIDs = append(familyIDs, f.ID)

        projMap[f.ProjectID].Families = append(projMap[f.ProjectID].Families, *f)
    }

    if len(families) == 0 {
        // No hay familias → retornar proyectos tal cual
        out := make([]*project.Project, 0, len(projects))
        for i := range projects {
            out = append(out, &projects[i])
        }
        return out, nil
    }

    // ---------------------------
    // 4. Miembros
    // ---------------------------
    query, args, err = sqlx.In(`
        SELECT id, family_id, result, date_of_fracture, is_reported
        FROM members
        WHERE family_id IN (?)`, familyIDs)
    if err != nil {
        log.Printf("[GetProjectsByClientID] sqlx.In error members. err=%v", err)
        return nil, err
    }

    query = p.db.Rebind(query)

    var members []member.Member
    if err := p.db.Select(&members, query, args...); err != nil {
        log.Printf("[GetProjectsByClientID] Failed loading members. err=%v", err)
        return nil, err
    }

    log.Printf("[GetProjectsByClientID] Members found=%d", len(members))

    for _, m := range members {
        famMap[m.FamilyID].Members = append(famMap[m.FamilyID].Members, m)
    }

    // Reconstruir estructuras anidadas
    for projID := range projMap {
        for i := range projMap[projID].Families {
            famID := projMap[projID].Families[i].ID
            projMap[projID].Families[i] = *famMap[famID]
        }
    }

    // ---------------------------
    // Final
    // ---------------------------
    out := make([]*project.Project, 0, len(projects))
    for i := range projects {
        out = append(out, &projects[i])
    }

    log.Printf("[GetProjectsByClientID] Returning %d projects", len(out))
    return out, nil
}


