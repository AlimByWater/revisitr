package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"revisitr/internal/entity"
)

type Users struct {
	pg *Module
}

func NewUsers(pg *Module) *Users {
	return &Users{pg: pg}
}

func (r *Users) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (email, phone, name, password_hash, role, org_id)
		VALUES (:email, :phone, :name, :password_hash, :role, :org_id)
		RETURNING id, created_at`

	rows, err := r.pg.DB().NamedQueryContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("users.Create: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&user.ID, &user.CreatedAt); err != nil {
			return fmt.Errorf("users.Create scan: %w", err)
		}
	}

	return nil
}

func (r *Users) GetByID(ctx context.Context, id int) (*entity.User, error) {
	var user entity.User
	err := r.pg.DB().GetContext(ctx, &user, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("users.GetByID: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("users.GetByID: %w", err)
	}
	return &user, nil
}

func (r *Users) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.pg.DB().GetContext(ctx, &user, "SELECT * FROM users WHERE email = $1", email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("users.GetByEmail: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("users.GetByEmail: %w", err)
	}
	return &user, nil
}

func (r *Users) Update(ctx context.Context, user *entity.User) error {
	query := `UPDATE users SET name = :name, phone = :phone, email = :email WHERE id = :id`
	result, err := r.pg.DB().NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("users.Update: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("users.Update rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("users.Update: %w", sql.ErrNoRows)
	}

	return nil
}

func (r *Users) CreateOrganization(ctx context.Context, org *entity.Organization) error {
	query := `
		INSERT INTO organizations (name, owner_id)
		VALUES (:name, :owner_id)
		RETURNING id, created_at`

	rows, err := r.pg.DB().NamedQueryContext(ctx, query, org)
	if err != nil {
		return fmt.Errorf("users.CreateOrganization: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&org.ID, &org.CreatedAt); err != nil {
			return fmt.Errorf("users.CreateOrganization scan: %w", err)
		}
	}

	return nil
}

func (r *Users) UpdateOrganizationOwner(ctx context.Context, orgID, ownerID int) error {
	result, err := r.pg.DB().ExecContext(ctx, "UPDATE organizations SET owner_id = $1 WHERE id = $2", ownerID, orgID)
	if err != nil {
		return fmt.Errorf("users.UpdateOrganizationOwner: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("users.UpdateOrganizationOwner rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("users.UpdateOrganizationOwner: %w", sql.ErrNoRows)
	}

	return nil
}
