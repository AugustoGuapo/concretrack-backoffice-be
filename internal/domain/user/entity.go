package user

type User struct {
    ID        int    `db:"id" json:"id"`
    FirstName string `db:"first_name" json:"firstName"`
    LastName  string `db:"last_name" json:"lastName"`
    Role      string `db:"role" json:"role"`
    Username  string `db:"username" json:"username"`
    Password  string `db:"password" json:"password"`
    IsActive  bool   `db:"is_active" json:"isActive"`
}

