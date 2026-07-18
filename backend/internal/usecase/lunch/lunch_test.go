package lunch

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"revisitr/internal/entity"
)

func ptr[T any](value T) *T { return &value }

// --- mocks ---

type mockLunchRepo struct {
	getProgramByBotIDFn     func(ctx context.Context, botID int) (*entity.LunchProgram, error)
	getFullProgramByBotIDFn func(ctx context.Context, botID int) (*entity.LunchProgram, error)
	createProgramFn         func(ctx context.Context, p *entity.LunchProgram) error
	updateProgramFn         func(ctx context.Context, p *entity.LunchProgram) error
	createCourseFn          func(ctx context.Context, c *entity.LunchCourse) error
	updateCourseFn          func(ctx context.Context, c *entity.LunchCourse) error
	deleteCourseFn          func(ctx context.Context, id int) error
	getCourseFn             func(ctx context.Context, id int) (*entity.LunchCourse, error)
	getCourseOrgIDFn        func(ctx context.Context, courseID int) (int, error)
	createFormatFn          func(ctx context.Context, f *entity.LunchFormat) error
	updateFormatFn          func(ctx context.Context, f *entity.LunchFormat) error
	deleteFormatFn          func(ctx context.Context, id int) error
	getFormatFn             func(ctx context.Context, id int) (*entity.LunchFormat, error)
	getFormatOrgIDFn        func(ctx context.Context, formatID int) (int, error)
	replaceAvailabilityFn   func(ctx context.Context, programID int, slots []entity.LunchAvailability) error
}

func (m *mockLunchRepo) GetProgramByBotID(ctx context.Context, botID int) (*entity.LunchProgram, error) {
	if m.getProgramByBotIDFn != nil {
		return m.getProgramByBotIDFn(ctx, botID)
	}
	return nil, nil
}
func (m *mockLunchRepo) GetFullProgramByBotID(ctx context.Context, botID int) (*entity.LunchProgram, error) {
	if m.getFullProgramByBotIDFn != nil {
		return m.getFullProgramByBotIDFn(ctx, botID)
	}
	return nil, nil
}
func (m *mockLunchRepo) CreateProgram(ctx context.Context, p *entity.LunchProgram) error {
	if m.createProgramFn != nil {
		return m.createProgramFn(ctx, p)
	}
	return nil
}
func (m *mockLunchRepo) UpdateProgram(ctx context.Context, p *entity.LunchProgram) error {
	if m.updateProgramFn != nil {
		return m.updateProgramFn(ctx, p)
	}
	return nil
}
func (m *mockLunchRepo) CreateCourse(ctx context.Context, c *entity.LunchCourse) error {
	if m.createCourseFn != nil {
		return m.createCourseFn(ctx, c)
	}
	return nil
}
func (m *mockLunchRepo) UpdateCourse(ctx context.Context, c *entity.LunchCourse) error {
	if m.updateCourseFn != nil {
		return m.updateCourseFn(ctx, c)
	}
	return nil
}
func (m *mockLunchRepo) DeleteCourse(ctx context.Context, id int) error {
	if m.deleteCourseFn != nil {
		return m.deleteCourseFn(ctx, id)
	}
	return nil
}
func (m *mockLunchRepo) GetCourse(ctx context.Context, id int) (*entity.LunchCourse, error) {
	if m.getCourseFn != nil {
		return m.getCourseFn(ctx, id)
	}
	return nil, nil
}
func (m *mockLunchRepo) GetCourseOrgID(ctx context.Context, courseID int) (int, error) {
	if m.getCourseOrgIDFn != nil {
		return m.getCourseOrgIDFn(ctx, courseID)
	}
	return 0, nil
}
func (m *mockLunchRepo) CreateFormat(ctx context.Context, f *entity.LunchFormat) error {
	if m.createFormatFn != nil {
		return m.createFormatFn(ctx, f)
	}
	return nil
}
func (m *mockLunchRepo) UpdateFormat(ctx context.Context, f *entity.LunchFormat) error {
	if m.updateFormatFn != nil {
		return m.updateFormatFn(ctx, f)
	}
	return nil
}
func (m *mockLunchRepo) DeleteFormat(ctx context.Context, id int) error {
	if m.deleteFormatFn != nil {
		return m.deleteFormatFn(ctx, id)
	}
	return nil
}
func (m *mockLunchRepo) GetFormat(ctx context.Context, id int) (*entity.LunchFormat, error) {
	if m.getFormatFn != nil {
		return m.getFormatFn(ctx, id)
	}
	return nil, nil
}
func (m *mockLunchRepo) GetFormatOrgID(ctx context.Context, formatID int) (int, error) {
	if m.getFormatOrgIDFn != nil {
		return m.getFormatOrgIDFn(ctx, formatID)
	}
	return 0, nil
}
func (m *mockLunchRepo) ReplaceAvailability(ctx context.Context, programID int, slots []entity.LunchAvailability) error {
	if m.replaceAvailabilityFn != nil {
		return m.replaceAvailabilityFn(ctx, programID, slots)
	}
	return nil
}
type mockBotsGetter struct {
	getByIDFn func(ctx context.Context, id int) (*entity.Bot, error)
}

func (m *mockBotsGetter) GetByID(ctx context.Context, id int) (*entity.Bot, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return &entity.Bot{ID: id, OrgID: 1}, nil
}

func newTestUsecase(t *testing.T, repo *mockLunchRepo, bots *mockBotsGetter) *Usecase {
	t.Helper()
	if repo == nil {
		repo = &mockLunchRepo{}
	}
	if bots == nil {
		bots = &mockBotsGetter{}
	}
	uc := New(repo, bots)
	if err := uc.Init(context.Background(), slog.Default()); err != nil {
		t.Fatalf("init usecase: %v", err)
	}
	return uc
}

// --- program ---

func TestGetProgramLazyCreates(t *testing.T) {
	created := false
	repo := &mockLunchRepo{
		createProgramFn: func(_ context.Context, p *entity.LunchProgram) error {
			created = true
			p.ID = 42
			return nil
		},
	}
	uc := newTestUsecase(t, repo, nil)

	p, err := uc.GetProgram(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Error("expected program to be lazily created")
	}
	if p.ID != 42 || p.BotID != 10 || p.IsActive {
		t.Errorf("unexpected program: %+v (must be inactive with bot id set)", p)
	}
}

func TestGetProgramForeignBot(t *testing.T) {
	bots := &mockBotsGetter{
		getByIDFn: func(_ context.Context, id int) (*entity.Bot, error) {
			return &entity.Bot{ID: id, OrgID: 99}, nil
		},
	}
	uc := newTestUsecase(t, nil, bots)

	if _, err := uc.GetProgram(context.Background(), 1, 10); !errors.Is(err, ErrNotOwner) {
		t.Errorf("expected ErrNotOwner, got %v", err)
	}
}

// --- courses (FR-A6) ---

func programRepo() *mockLunchRepo {
	return &mockLunchRepo{
		getFullProgramByBotIDFn: func(_ context.Context, botID int) (*entity.LunchProgram, error) {
			return &entity.LunchProgram{ID: 5, BotID: botID}, nil
		},
	}
}

func TestCreateCourseWithoutCategory(t *testing.T) {
	uc := newTestUsecase(t, programRepo(), nil)
	_, err := uc.CreateCourse(context.Background(), 1, 10, entity.CreateLunchCourseRequest{
		Code:  "1",
		Title: "Первое",
	})
	if !errors.Is(err, ErrValidation) {
		t.Errorf("course without category: expected ErrValidation, got %v", err)
	}
}

func TestCreateCourseOK(t *testing.T) {
	repo := programRepo()
	repo.createCourseFn = func(_ context.Context, c *entity.LunchCourse) error {
		c.ID = 7
		return nil
	}
	uc := newTestUsecase(t, repo, nil)

	course, err := uc.CreateCourse(context.Background(), 1, 10, entity.CreateLunchCourseRequest{
		Code:           "1",
		Title:          "Первое",
		MenuCategoryID: 3,
		Items:          []entity.LunchCourseItemRequest{{MenuItemID: 100, Surcharge: 50}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if course.ProgramID != 5 || len(course.Items) != 1 || course.Items[0].Surcharge != 50 {
		t.Errorf("unexpected course: %+v", course)
	}
}

func TestUpdateCourseForeignOrg(t *testing.T) {
	repo := &mockLunchRepo{
		getCourseOrgIDFn: func(_ context.Context, _ int) (int, error) { return 99, nil },
	}
	uc := newTestUsecase(t, repo, nil)

	err := uc.UpdateCourse(context.Background(), 1, 7, entity.UpdateLunchCourseRequest{Title: ptr("X")})
	if !errors.Is(err, ErrNotOwner) {
		t.Errorf("expected ErrNotOwner, got %v", err)
	}
}

func TestDeleteCourseNotFound(t *testing.T) {
	uc := newTestUsecase(t, &mockLunchRepo{}, nil)
	if err := uc.DeleteCourse(context.Background(), 1, 7); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// --- formats (FR-A6) ---

func formatRepo(courseItems int) *mockLunchRepo {
	repo := programRepo()
	repo.getCourseFn = func(_ context.Context, id int) (*entity.LunchCourse, error) {
		items := make([]entity.LunchCourseItem, courseItems)
		return &entity.LunchCourse{ID: id, ProgramID: 5, Title: "Первое", Items: items}, nil
	}
	return repo
}

func TestCreateFormatEmptyCourse(t *testing.T) {
	uc := newTestUsecase(t, formatRepo(0), nil)
	_, err := uc.CreateFormat(context.Background(), 1, 10, entity.CreateLunchFormatRequest{
		Name:      "1+2",
		PriceMode: entity.LunchPriceFixed,
		BasePrice: 350,
		CourseIDs: []int{1},
	})
	if !errors.Is(err, ErrValidation) {
		t.Errorf("format with empty course: expected ErrValidation, got %v", err)
	}
}

func TestCreateFormatFixedZeroPrice(t *testing.T) {
	uc := newTestUsecase(t, formatRepo(2), nil)
	_, err := uc.CreateFormat(context.Background(), 1, 10, entity.CreateLunchFormatRequest{
		Name:      "1+2",
		PriceMode: entity.LunchPriceFixed,
		BasePrice: 0,
		CourseIDs: []int{1},
	})
	if !errors.Is(err, ErrValidation) {
		t.Errorf("fixed format with zero price: expected ErrValidation, got %v", err)
	}
}

func TestCreateFormatNoCourses(t *testing.T) {
	uc := newTestUsecase(t, formatRepo(2), nil)
	_, err := uc.CreateFormat(context.Background(), 1, 10, entity.CreateLunchFormatRequest{
		Name:      "1+2",
		PriceMode: entity.LunchPriceSumOfItems,
	})
	if !errors.Is(err, ErrValidation) {
		t.Errorf("format without courses: expected ErrValidation, got %v", err)
	}
}

func TestCreateFormatUnknownPriceMode(t *testing.T) {
	uc := newTestUsecase(t, formatRepo(2), nil)
	_, err := uc.CreateFormat(context.Background(), 1, 10, entity.CreateLunchFormatRequest{
		Name:      "1+2",
		PriceMode: "percent",
		CourseIDs: []int{1},
	})
	if !errors.Is(err, ErrValidation) {
		t.Errorf("unknown price mode: expected ErrValidation, got %v", err)
	}
}

func TestCreateFormatForeignCourse(t *testing.T) {
	repo := programRepo()
	repo.getCourseFn = func(_ context.Context, id int) (*entity.LunchCourse, error) {
		return &entity.LunchCourse{ID: id, ProgramID: 777, Items: make([]entity.LunchCourseItem, 1)}, nil
	}
	uc := newTestUsecase(t, repo, nil)
	_, err := uc.CreateFormat(context.Background(), 1, 10, entity.CreateLunchFormatRequest{
		Name:      "1+2",
		PriceMode: entity.LunchPriceSumOfItems,
		CourseIDs: []int{1},
	})
	if !errors.Is(err, ErrValidation) {
		t.Errorf("course from another program: expected ErrValidation, got %v", err)
	}
}

func TestCreateFormatOK(t *testing.T) {
	repo := formatRepo(2)
	repo.createFormatFn = func(_ context.Context, f *entity.LunchFormat) error {
		f.ID = 11
		return nil
	}
	uc := newTestUsecase(t, repo, nil)

	format, err := uc.CreateFormat(context.Background(), 1, 10, entity.CreateLunchFormatRequest{
		Name:      "Первое + Второе",
		PriceMode: entity.LunchPriceFixed,
		BasePrice: 350,
		CourseIDs: []int{1, 2},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if format.ID != 11 || format.ProgramID != 5 {
		t.Errorf("unexpected format: %+v", format)
	}
}

func TestUpdateFormatForeignOrg(t *testing.T) {
	repo := &mockLunchRepo{
		getFormatOrgIDFn: func(_ context.Context, _ int) (int, error) { return 99, nil },
	}
	uc := newTestUsecase(t, repo, nil)

	err := uc.UpdateFormat(context.Background(), 1, 11, entity.UpdateLunchFormatRequest{Name: ptr("X")})
	if !errors.Is(err, ErrNotOwner) {
		t.Errorf("expected ErrNotOwner, got %v", err)
	}
}

// --- availability ---

func TestSetAvailabilityValidation(t *testing.T) {
	uc := newTestUsecase(t, programRepo(), nil)
	ctx := context.Background()

	cases := []struct {
		name string
		slot entity.LunchAvailability
	}{
		{"bad weekday", entity.LunchAvailability{Weekday: 8, TimeFrom: "12:00", TimeTo: "16:00"}},
		{"bad time", entity.LunchAvailability{Weekday: 1, TimeFrom: "noon", TimeTo: "16:00"}},
		{"inverted range", entity.LunchAvailability{Weekday: 1, TimeFrom: "16:00", TimeTo: "12:00"}},
	}
	for _, tc := range cases {
		err := uc.SetAvailability(ctx, 1, 10, []entity.LunchAvailability{tc.slot})
		if !errors.Is(err, ErrValidation) {
			t.Errorf("%s: expected ErrValidation, got %v", tc.name, err)
		}
	}
}

func TestSetAvailabilityOK(t *testing.T) {
	var savedProgramID int
	repo := programRepo()
	repo.replaceAvailabilityFn = func(_ context.Context, programID int, _ []entity.LunchAvailability) error {
		savedProgramID = programID
		return nil
	}
	uc := newTestUsecase(t, repo, nil)

	err := uc.SetAvailability(context.Background(), 1, 10, []entity.LunchAvailability{
		{Weekday: 1, TimeFrom: "12:00", TimeTo: "16:00"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if savedProgramID != 5 {
		t.Errorf("saved to program %d, want 5", savedProgramID)
	}
}
