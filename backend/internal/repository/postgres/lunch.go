package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"revisitr/internal/entity"
)

type Lunch struct {
	pg *Module
}

func NewLunch(pg *Module) *Lunch {
	return &Lunch{pg: pg}
}

func (r *Lunch) GetProgramByBotID(ctx context.Context, botID int) (*entity.LunchProgram, error) {
	var p entity.LunchProgram
	err := r.pg.DB().GetContext(ctx, &p,
		"SELECT * FROM lunch_programs WHERE bot_id = $1", botID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("lunch.GetProgramByBotID: %w", err)
	}
	return &p, nil
}

// GetFullProgramByBotID loads the program with courses (items hydrated from
// menu_items), formats (course ids ordered by position) and availability.
// Items are returned regardless of is_available — renderers filter.
func (r *Lunch) GetFullProgramByBotID(ctx context.Context, botID int) (*entity.LunchProgram, error) {
	p, err := r.GetProgramByBotID(ctx, botID)
	if err != nil || p == nil {
		return p, err
	}

	if err := r.pg.DB().SelectContext(ctx, &p.Courses,
		"SELECT * FROM lunch_courses WHERE program_id = $1 ORDER BY sort_order, id", p.ID); err != nil {
		return nil, fmt.Errorf("lunch.GetFullProgramByBotID courses: %w", err)
	}
	for i := range p.Courses {
		items, itemsErr := r.getCourseItems(ctx, p.Courses[i].ID)
		if itemsErr != nil {
			return nil, itemsErr
		}
		p.Courses[i].Items = items
	}

	if err := r.pg.DB().SelectContext(ctx, &p.Formats,
		"SELECT * FROM lunch_formats WHERE program_id = $1 ORDER BY sort_order, id", p.ID); err != nil {
		return nil, fmt.Errorf("lunch.GetFullProgramByBotID formats: %w", err)
	}
	for i := range p.Formats {
		if err := r.pg.DB().SelectContext(ctx, &p.Formats[i].CourseIDs,
			"SELECT course_id FROM lunch_format_courses WHERE format_id = $1 ORDER BY position, course_id",
			p.Formats[i].ID); err != nil {
			return nil, fmt.Errorf("lunch.GetFullProgramByBotID format courses: %w", err)
		}
	}

	if err := r.pg.DB().SelectContext(ctx, &p.Availability,
		"SELECT id, program_id, weekday, to_char(time_from, 'HH24:MI') AS time_from, to_char(time_to, 'HH24:MI') AS time_to FROM lunch_availability WHERE program_id = $1 ORDER BY weekday, time_from",
		p.ID); err != nil {
		return nil, fmt.Errorf("lunch.GetFullProgramByBotID availability: %w", err)
	}

	return p, nil
}

func (r *Lunch) getCourseItems(ctx context.Context, courseID int) ([]entity.LunchCourseItem, error) {
	rows, err := r.pg.DB().QueryContext(ctx, `
		SELECT lci.course_id, lci.menu_item_id, lci.surcharge,
		       mi.id, mi.category_id, mi.name, mi.description, mi.price,
		       mi.weight, mi.image_url, mi.is_available, mi.sort_order
		FROM lunch_course_items lci
		JOIN menu_items mi ON mi.id = lci.menu_item_id
		WHERE lci.course_id = $1
		ORDER BY mi.sort_order, mi.id`, courseID)
	if err != nil {
		return nil, fmt.Errorf("lunch.getCourseItems: %w", err)
	}
	defer rows.Close()

	items := []entity.LunchCourseItem{}
	for rows.Next() {
		var it entity.LunchCourseItem
		var mi entity.MenuItem
		if err := rows.Scan(
			&it.CourseID, &it.MenuItemID, &it.Surcharge,
			&mi.ID, &mi.CategoryID, &mi.Name, &mi.Description, &mi.Price,
			&mi.Weight, &mi.ImageURL, &mi.IsAvailable, &mi.SortOrder,
		); err != nil {
			return nil, fmt.Errorf("lunch.getCourseItems scan: %w", err)
		}
		it.MenuItem = &mi
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("lunch.getCourseItems rows: %w", err)
	}
	return items, nil
}

func (r *Lunch) CreateProgram(ctx context.Context, p *entity.LunchProgram) error {
	err := r.pg.DB().QueryRowContext(ctx, `
		INSERT INTO lunch_programs (bot_id, name, description, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`,
		p.BotID, p.Name, p.Description, p.IsActive,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("lunch.CreateProgram: %w", err)
	}
	return nil
}

func (r *Lunch) UpdateProgram(ctx context.Context, p *entity.LunchProgram) error {
	_, err := r.pg.DB().ExecContext(ctx, `
		UPDATE lunch_programs
		SET name = $2, description = $3, is_active = $4, updated_at = now()
		WHERE id = $1`,
		p.ID, p.Name, p.Description, p.IsActive)
	if err != nil {
		return fmt.Errorf("lunch.UpdateProgram: %w", err)
	}
	return nil
}

func (r *Lunch) CreateCourse(ctx context.Context, c *entity.LunchCourse) error {
	tx, err := r.pg.DB().BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("lunch.CreateCourse begin: %w", err)
	}
	defer tx.Rollback()

	if err := tx.QueryRowContext(ctx, `
		INSERT INTO lunch_courses (program_id, code, title, menu_category_id, sort_order)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		c.ProgramID, c.Code, c.Title, c.MenuCategoryID, c.SortOrder,
	).Scan(&c.ID); err != nil {
		return fmt.Errorf("lunch.CreateCourse insert: %w", err)
	}

	for _, item := range c.Items {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO lunch_course_items (course_id, menu_item_id, surcharge)
			VALUES ($1, $2, $3)`,
			c.ID, item.MenuItemID, item.Surcharge); err != nil {
			return fmt.Errorf("lunch.CreateCourse item %d: %w", item.MenuItemID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("lunch.CreateCourse commit: %w", err)
	}
	return nil
}

func (r *Lunch) UpdateCourse(ctx context.Context, c *entity.LunchCourse) error {
	tx, err := r.pg.DB().BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("lunch.UpdateCourse begin: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		UPDATE lunch_courses
		SET code = $2, title = $3, menu_category_id = $4, sort_order = $5
		WHERE id = $1`,
		c.ID, c.Code, c.Title, c.MenuCategoryID, c.SortOrder); err != nil {
		return fmt.Errorf("lunch.UpdateCourse update: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		"DELETE FROM lunch_course_items WHERE course_id = $1", c.ID); err != nil {
		return fmt.Errorf("lunch.UpdateCourse clear items: %w", err)
	}
	for _, item := range c.Items {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO lunch_course_items (course_id, menu_item_id, surcharge)
			VALUES ($1, $2, $3)`,
			c.ID, item.MenuItemID, item.Surcharge); err != nil {
			return fmt.Errorf("lunch.UpdateCourse item %d: %w", item.MenuItemID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("lunch.UpdateCourse commit: %w", err)
	}
	return nil
}

func (r *Lunch) DeleteCourse(ctx context.Context, id int) error {
	_, err := r.pg.DB().ExecContext(ctx, "DELETE FROM lunch_courses WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("lunch.DeleteCourse: %w", err)
	}
	return nil
}

func (r *Lunch) GetCourse(ctx context.Context, id int) (*entity.LunchCourse, error) {
	var c entity.LunchCourse
	err := r.pg.DB().GetContext(ctx, &c, "SELECT * FROM lunch_courses WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("lunch.GetCourse: %w", err)
	}
	items, err := r.getCourseItems(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	c.Items = items
	return &c, nil
}

func (r *Lunch) GetCourseOrgID(ctx context.Context, courseID int) (int, error) {
	var orgID int
	err := r.pg.DB().GetContext(ctx, &orgID, `
		SELECT b.org_id FROM lunch_courses lc
		JOIN lunch_programs lp ON lp.id = lc.program_id
		JOIN bots b ON b.id = lp.bot_id
		WHERE lc.id = $1`, courseID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("lunch.GetCourseOrgID: %w", err)
	}
	return orgID, nil
}

func (r *Lunch) CreateFormat(ctx context.Context, f *entity.LunchFormat) error {
	tx, err := r.pg.DB().BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("lunch.CreateFormat begin: %w", err)
	}
	defer tx.Rollback()

	if err := tx.QueryRowContext(ctx, `
		INSERT INTO lunch_formats (program_id, name, price_mode, base_price, is_active, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		f.ProgramID, f.Name, f.PriceMode, f.BasePrice, f.IsActive, f.SortOrder,
	).Scan(&f.ID); err != nil {
		return fmt.Errorf("lunch.CreateFormat insert: %w", err)
	}

	for pos, courseID := range f.CourseIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO lunch_format_courses (format_id, course_id, position)
			VALUES ($1, $2, $3)`,
			f.ID, courseID, pos); err != nil {
			return fmt.Errorf("lunch.CreateFormat course %d: %w", courseID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("lunch.CreateFormat commit: %w", err)
	}
	return nil
}

func (r *Lunch) UpdateFormat(ctx context.Context, f *entity.LunchFormat) error {
	tx, err := r.pg.DB().BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("lunch.UpdateFormat begin: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		UPDATE lunch_formats
		SET name = $2, price_mode = $3, base_price = $4, is_active = $5, sort_order = $6
		WHERE id = $1`,
		f.ID, f.Name, f.PriceMode, f.BasePrice, f.IsActive, f.SortOrder); err != nil {
		return fmt.Errorf("lunch.UpdateFormat update: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		"DELETE FROM lunch_format_courses WHERE format_id = $1", f.ID); err != nil {
		return fmt.Errorf("lunch.UpdateFormat clear courses: %w", err)
	}
	for pos, courseID := range f.CourseIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO lunch_format_courses (format_id, course_id, position)
			VALUES ($1, $2, $3)`,
			f.ID, courseID, pos); err != nil {
			return fmt.Errorf("lunch.UpdateFormat course %d: %w", courseID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("lunch.UpdateFormat commit: %w", err)
	}
	return nil
}

func (r *Lunch) DeleteFormat(ctx context.Context, id int) error {
	_, err := r.pg.DB().ExecContext(ctx, "DELETE FROM lunch_formats WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("lunch.DeleteFormat: %w", err)
	}
	return nil
}

func (r *Lunch) GetFormat(ctx context.Context, id int) (*entity.LunchFormat, error) {
	var f entity.LunchFormat
	err := r.pg.DB().GetContext(ctx, &f, "SELECT * FROM lunch_formats WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("lunch.GetFormat: %w", err)
	}
	if err := r.pg.DB().SelectContext(ctx, &f.CourseIDs,
		"SELECT course_id FROM lunch_format_courses WHERE format_id = $1 ORDER BY position, course_id",
		f.ID); err != nil {
		return nil, fmt.Errorf("lunch.GetFormat courses: %w", err)
	}
	return &f, nil
}

func (r *Lunch) GetFormatOrgID(ctx context.Context, formatID int) (int, error) {
	var orgID int
	err := r.pg.DB().GetContext(ctx, &orgID, `
		SELECT b.org_id FROM lunch_formats lf
		JOIN lunch_programs lp ON lp.id = lf.program_id
		JOIN bots b ON b.id = lp.bot_id
		WHERE lf.id = $1`, formatID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("lunch.GetFormatOrgID: %w", err)
	}
	return orgID, nil
}

func (r *Lunch) ReplaceAvailability(ctx context.Context, programID int, slots []entity.LunchAvailability) error {
	tx, err := r.pg.DB().BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("lunch.ReplaceAvailability begin: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		"DELETE FROM lunch_availability WHERE program_id = $1", programID); err != nil {
		return fmt.Errorf("lunch.ReplaceAvailability clear: %w", err)
	}
	for _, slot := range slots {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO lunch_availability (program_id, weekday, time_from, time_to)
			VALUES ($1, $2, $3, $4)`,
			programID, slot.Weekday, slot.TimeFrom, slot.TimeTo); err != nil {
			return fmt.Errorf("lunch.ReplaceAvailability insert: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("lunch.ReplaceAvailability commit: %w", err)
	}
	return nil
}

