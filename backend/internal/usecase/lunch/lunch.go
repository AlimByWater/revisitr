package lunch

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"revisitr/internal/entity"
)

var (
	ErrNotFound   = errors.New("lunch entity not found")
	ErrNotOwner   = errors.New("not the owner")
	ErrValidation = errors.New("validation failed")
)

type lunchRepo interface {
	GetProgramByBotID(ctx context.Context, botID int) (*entity.LunchProgram, error)
	GetFullProgramByBotID(ctx context.Context, botID int) (*entity.LunchProgram, error)
	CreateProgram(ctx context.Context, p *entity.LunchProgram) error
	UpdateProgram(ctx context.Context, p *entity.LunchProgram) error
	CreateCourse(ctx context.Context, c *entity.LunchCourse) error
	UpdateCourse(ctx context.Context, c *entity.LunchCourse) error
	DeleteCourse(ctx context.Context, id int) error
	GetCourse(ctx context.Context, id int) (*entity.LunchCourse, error)
	GetCourseOrgID(ctx context.Context, courseID int) (int, error)
	CreateFormat(ctx context.Context, f *entity.LunchFormat) error
	UpdateFormat(ctx context.Context, f *entity.LunchFormat) error
	DeleteFormat(ctx context.Context, id int) error
	GetFormat(ctx context.Context, id int) (*entity.LunchFormat, error)
	GetFormatOrgID(ctx context.Context, formatID int) (int, error)
	ReplaceAvailability(ctx context.Context, programID int, slots []entity.LunchAvailability) error
}

type botsGetter interface {
	GetByID(ctx context.Context, id int) (*entity.Bot, error)
}

type Usecase struct {
	logger *slog.Logger
	repo   lunchRepo
	bots   botsGetter
}

func New(repo lunchRepo, bots botsGetter) *Usecase {
	return &Usecase{repo: repo, bots: bots}
}

func (uc *Usecase) Init(_ context.Context, logger *slog.Logger) error {
	uc.logger = logger
	return nil
}

// checkBot verifies the bot exists and belongs to the org.
func (uc *Usecase) checkBot(ctx context.Context, orgID, botID int) error {
	bot, err := uc.bots.GetByID(ctx, botID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if bot.OrgID != orgID {
		return ErrNotOwner
	}
	return nil
}

// GetProgram returns the bot's lunch program with courses, formats and
// availability, lazily creating an empty inactive program on first access.
func (uc *Usecase) GetProgram(ctx context.Context, orgID, botID int) (*entity.LunchProgram, error) {
	if err := uc.checkBot(ctx, orgID, botID); err != nil {
		return nil, err
	}
	p, err := uc.repo.GetFullProgramByBotID(ctx, botID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		p = &entity.LunchProgram{
			BotID: botID,
			Name:  "Бизнес-ланч",
		}
		if err := uc.repo.CreateProgram(ctx, p); err != nil {
			return nil, fmt.Errorf("create lunch program: %w", err)
		}
		p.Courses = []entity.LunchCourse{}
		p.Formats = []entity.LunchFormat{}
		p.Availability = []entity.LunchAvailability{}
	}
	return p, nil
}

func (uc *Usecase) UpsertProgram(ctx context.Context, orgID, botID int, req entity.UpsertLunchProgramRequest) (*entity.LunchProgram, error) {
	p, err := uc.GetProgram(ctx, orgID, botID)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		if *req.Name == "" {
			return nil, fmt.Errorf("%w: name is required", ErrValidation)
		}
		p.Name = *req.Name
	}
	if req.Description != nil {
		p.Description = *req.Description
	}
	if req.IsActive != nil {
		p.IsActive = *req.IsActive
	}
	if err := uc.repo.UpdateProgram(ctx, p); err != nil {
		return nil, fmt.Errorf("update lunch program: %w", err)
	}
	return p, nil
}

func validateCourse(c *entity.LunchCourse) error {
	if c.Title == "" {
		return fmt.Errorf("%w: course title is required", ErrValidation)
	}
	if c.Code == "" {
		return fmt.Errorf("%w: course code is required", ErrValidation)
	}
	if c.MenuCategoryID == 0 {
		return fmt.Errorf("%w: course requires a menu category", ErrValidation)
	}
	return nil
}

func (uc *Usecase) CreateCourse(ctx context.Context, orgID, botID int, req entity.CreateLunchCourseRequest) (*entity.LunchCourse, error) {
	p, err := uc.GetProgram(ctx, orgID, botID)
	if err != nil {
		return nil, err
	}
	course := &entity.LunchCourse{
		ProgramID:      p.ID,
		Code:           req.Code,
		Title:          req.Title,
		MenuCategoryID: req.MenuCategoryID,
		SortOrder:      req.SortOrder,
		Items:          courseItemsFromRequest(req.Items),
	}
	if err := validateCourse(course); err != nil {
		return nil, err
	}
	if err := uc.repo.CreateCourse(ctx, course); err != nil {
		return nil, fmt.Errorf("create lunch course: %w", err)
	}
	return course, nil
}

func (uc *Usecase) UpdateCourse(ctx context.Context, orgID, courseID int, req entity.UpdateLunchCourseRequest) error {
	if err := uc.checkCourseOwner(ctx, orgID, courseID); err != nil {
		return err
	}
	course, err := uc.repo.GetCourse(ctx, courseID)
	if err != nil {
		return err
	}
	if course == nil {
		return ErrNotFound
	}
	if req.Code != nil {
		course.Code = *req.Code
	}
	if req.Title != nil {
		course.Title = *req.Title
	}
	if req.MenuCategoryID != nil {
		course.MenuCategoryID = *req.MenuCategoryID
	}
	if req.SortOrder != nil {
		course.SortOrder = *req.SortOrder
	}
	if req.Items != nil {
		course.Items = courseItemsFromRequest(req.Items)
	}
	if err := validateCourse(course); err != nil {
		return err
	}
	if err := uc.repo.UpdateCourse(ctx, course); err != nil {
		return fmt.Errorf("update lunch course: %w", err)
	}
	return nil
}

func (uc *Usecase) DeleteCourse(ctx context.Context, orgID, courseID int) error {
	if err := uc.checkCourseOwner(ctx, orgID, courseID); err != nil {
		return err
	}
	return uc.repo.DeleteCourse(ctx, courseID)
}

func (uc *Usecase) checkCourseOwner(ctx context.Context, orgID, courseID int) error {
	courseOrg, err := uc.repo.GetCourseOrgID(ctx, courseID)
	if err != nil {
		return err
	}
	if courseOrg == 0 {
		return ErrNotFound
	}
	if courseOrg != orgID {
		return ErrNotOwner
	}
	return nil
}

func courseItemsFromRequest(items []entity.LunchCourseItemRequest) []entity.LunchCourseItem {
	result := make([]entity.LunchCourseItem, 0, len(items))
	for _, item := range items {
		result = append(result, entity.LunchCourseItem{
			MenuItemID: item.MenuItemID,
			Surcharge:  item.Surcharge,
		})
	}
	return result
}

func validPriceMode(mode string) bool {
	switch mode {
	case entity.LunchPriceFixed, entity.LunchPriceSumOfItems, entity.LunchPriceBasePlusSurcharge:
		return true
	}
	return false
}

// validateFormat enforces FR-A6: a format needs at least one course, every
// referenced course must exist in the program and have at least one item,
// and fixed pricing requires a positive base price.
func (uc *Usecase) validateFormat(ctx context.Context, f *entity.LunchFormat) error {
	if f.Name == "" {
		return fmt.Errorf("%w: format name is required", ErrValidation)
	}
	if !validPriceMode(f.PriceMode) {
		return fmt.Errorf("%w: unknown price mode %q", ErrValidation, f.PriceMode)
	}
	if f.PriceMode == entity.LunchPriceFixed && f.BasePrice <= 0 {
		return fmt.Errorf("%w: fixed price must be positive", ErrValidation)
	}
	if len(f.CourseIDs) == 0 {
		return fmt.Errorf("%w: format requires at least one course", ErrValidation)
	}
	for _, courseID := range f.CourseIDs {
		course, err := uc.repo.GetCourse(ctx, courseID)
		if err != nil {
			return err
		}
		if course == nil || course.ProgramID != f.ProgramID {
			return fmt.Errorf("%w: course %d does not belong to the program", ErrValidation, courseID)
		}
		if len(course.Items) == 0 {
			return fmt.Errorf("%w: course %q has no items", ErrValidation, course.Title)
		}
	}
	return nil
}

func (uc *Usecase) CreateFormat(ctx context.Context, orgID, botID int, req entity.CreateLunchFormatRequest) (*entity.LunchFormat, error) {
	p, err := uc.GetProgram(ctx, orgID, botID)
	if err != nil {
		return nil, err
	}
	format := &entity.LunchFormat{
		ProgramID: p.ID,
		Name:      req.Name,
		PriceMode: req.PriceMode,
		BasePrice: req.BasePrice,
		IsActive:  req.IsActive,
		SortOrder: req.SortOrder,
		CourseIDs: req.CourseIDs,
	}
	if err := uc.validateFormat(ctx, format); err != nil {
		return nil, err
	}
	if err := uc.repo.CreateFormat(ctx, format); err != nil {
		return nil, fmt.Errorf("create lunch format: %w", err)
	}
	return format, nil
}

func (uc *Usecase) UpdateFormat(ctx context.Context, orgID, formatID int, req entity.UpdateLunchFormatRequest) error {
	if err := uc.checkFormatOwner(ctx, orgID, formatID); err != nil {
		return err
	}
	format, err := uc.repo.GetFormat(ctx, formatID)
	if err != nil {
		return err
	}
	if format == nil {
		return ErrNotFound
	}
	if req.Name != nil {
		format.Name = *req.Name
	}
	if req.PriceMode != nil {
		format.PriceMode = *req.PriceMode
	}
	if req.BasePrice != nil {
		format.BasePrice = *req.BasePrice
	}
	if req.IsActive != nil {
		format.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		format.SortOrder = *req.SortOrder
	}
	if req.CourseIDs != nil {
		format.CourseIDs = req.CourseIDs
	}
	if err := uc.validateFormat(ctx, format); err != nil {
		return err
	}
	if err := uc.repo.UpdateFormat(ctx, format); err != nil {
		return fmt.Errorf("update lunch format: %w", err)
	}
	return nil
}

func (uc *Usecase) DeleteFormat(ctx context.Context, orgID, formatID int) error {
	if err := uc.checkFormatOwner(ctx, orgID, formatID); err != nil {
		return err
	}
	return uc.repo.DeleteFormat(ctx, formatID)
}

func (uc *Usecase) checkFormatOwner(ctx context.Context, orgID, formatID int) error {
	formatOrg, err := uc.repo.GetFormatOrgID(ctx, formatID)
	if err != nil {
		return err
	}
	if formatOrg == 0 {
		return ErrNotFound
	}
	if formatOrg != orgID {
		return ErrNotOwner
	}
	return nil
}

func (uc *Usecase) SetAvailability(ctx context.Context, orgID, botID int, slots []entity.LunchAvailability) error {
	p, err := uc.GetProgram(ctx, orgID, botID)
	if err != nil {
		return err
	}
	for _, slot := range slots {
		if slot.Weekday < 1 || slot.Weekday > 7 {
			return fmt.Errorf("%w: weekday must be 1-7, got %d", ErrValidation, slot.Weekday)
		}
		from, errFrom := parseMinutes(slot.TimeFrom)
		to, errTo := parseMinutes(slot.TimeTo)
		if errFrom != nil || errTo != nil {
			return fmt.Errorf("%w: time must be HH:MM", ErrValidation)
		}
		if from >= to {
			return fmt.Errorf("%w: time_from must be before time_to", ErrValidation)
		}
	}
	if err := uc.repo.ReplaceAvailability(ctx, p.ID, slots); err != nil {
		return fmt.Errorf("set lunch availability: %w", err)
	}
	return nil
}
