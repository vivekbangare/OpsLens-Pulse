package store

import (
	"database/sql"
)

func LoadUserPermissions(db *sql.DB, userID string, tenantID string) (map[string]bool, error) {

	rows, err := db.Query(`
		SELECT DISTINCT p.name
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		JOIN roles r ON r.id = rp.role_id
		JOIN group_roles gr ON gr.role_id = r.id
		JOIN user_groups ug ON ug.group_id = gr.group_id
		WHERE ug.user_id = $1
		  AND r.tenant_id = $2
	`, userID, tenantID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	perms := make(map[string]bool)

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			perms[name] = true
		}
	}

	return perms, nil
}
