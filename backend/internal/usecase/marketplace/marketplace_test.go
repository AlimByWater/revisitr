package marketplace

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"revisitr/internal/entity"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// ── Mocks ───────────────────────────────────────────────────────────────────

type mockProductsRepo struct {
	createProduct    func(ctx context.Context, orgID int, req entity.CreateProductRequest) (*entity.MarketplaceProduct, error)
	getProducts      func(ctx context.Context, orgID int) ([]entity.MarketplaceProduct, error)
	getActiveProducts func(ctx context.Context, orgID int) ([]entity.MarketplaceProduct, error)
	getProductByID   func(ctx context.Context, id int) (*entity.MarketplaceProduct, error)
	updateProduct    func(ctx context.Context, id int, req entity.UpdateProductRequest) (*entity.MarketplaceProduct, error)
	deleteProduct    func(ctx context.Context, id int) error
	decrementStock   func(ctx context.Context, productID int, qty int) error
}

func (m *mockProductsRepo) CreateProduct(ctx context.Context, orgID int, req entity.CreateProductRequest) (*entity.MarketplaceProduct, error) {
	return m.createProduct(ctx, orgID, req)
}
func (m *mockProductsRepo) GetProducts(ctx context.Context, orgID int) ([]entity.MarketplaceProduct, error) {
	return m.getProducts(ctx, orgID)
}
func (m *mockProductsRepo) GetActiveProducts(ctx context.Context, orgID int) ([]entity.MarketplaceProduct, error) {
	return m.getActiveProducts(ctx, orgID)
}
func (m *mockProductsRepo) GetProductByID(ctx context.Context, id int) (*entity.MarketplaceProduct, error) {
	return m.getProductByID(ctx, id)
}
func (m *mockProductsRepo) UpdateProduct(ctx context.Context, id int, req entity.UpdateProductRequest) (*entity.MarketplaceProduct, error) {
	return m.updateProduct(ctx, id, req)
}
func (m *mockProductsRepo) DeleteProduct(ctx context.Context, id int) error {
	return m.deleteProduct(ctx, id)
}
func (m *mockProductsRepo) DecrementStock(ctx context.Context, productID int, qty int) error {
	return m.decrementStock(ctx, productID, qty)
}

type mockLoyaltySpender struct {
	spendPoints func(ctx context.Context, clientID, programID int, amount float64, description string) (*entity.ClientLoyalty, error)
	getBalance  func(ctx context.Context, clientID, programID int) (*entity.ClientLoyalty, error)
}

func (m *mockLoyaltySpender) SpendPoints(ctx context.Context, clientID, programID int, amount float64, description string) (*entity.ClientLoyalty, error) {
	return m.spendPoints(ctx, clientID, programID, amount, description)
}
func (m *mockLoyaltySpender) GetBalance(ctx context.Context, clientID, programID int) (*entity.ClientLoyalty, error) {
	return m.getBalance(ctx, clientID, programID)
}

type mockOrdersRepo struct {
	createOrder       func(ctx context.Context, order *entity.MarketplaceOrder) error
	getOrders         func(ctx context.Context, orgID int) ([]entity.MarketplaceOrder, error)
	getOrderByID      func(ctx context.Context, id int) (*entity.MarketplaceOrder, error)
	getOrdersByClient func(ctx context.Context, clientID int) ([]entity.MarketplaceOrder, error)
	updateOrderStatus func(ctx context.Context, id int, status string) error
	getStats          func(ctx context.Context, orgID int) (*entity.MarketplaceStats, error)
}

func (m *mockOrdersRepo) CreateOrder(ctx context.Context, order *entity.MarketplaceOrder) error {
	return m.createOrder(ctx, order)
}
func (m *mockOrdersRepo) GetOrders(ctx context.Context, orgID int) ([]entity.MarketplaceOrder, error) {
	return m.getOrders(ctx, orgID)
}
func (m *mockOrdersRepo) GetOrderByID(ctx context.Context, id int) (*entity.MarketplaceOrder, error) {
	return m.getOrderByID(ctx, id)
}
func (m *mockOrdersRepo) GetOrdersByClient(ctx context.Context, clientID int) ([]entity.MarketplaceOrder, error) {
	return m.getOrdersByClient(ctx, clientID)
}
func (m *mockOrdersRepo) UpdateOrderStatus(ctx context.Context, id int, status string) error {
	return m.updateOrderStatus(ctx, id, status)
}
func (m *mockOrdersRepo) GetStats(ctx context.Context, orgID int) (*entity.MarketplaceStats, error) {
	return m.getStats(ctx, orgID)
}

func defaultLoyalty() *mockLoyaltySpender {
	return &mockLoyaltySpender{
		getBalance: func(_ context.Context, _, _ int) (*entity.ClientLoyalty, error) {
			return &entity.ClientLoyalty{Balance: 100000}, nil
		},
		spendPoints: func(_ context.Context, _, _ int, _ float64, _ string) (*entity.ClientLoyalty, error) {
			return &entity.ClientLoyalty{}, nil
		},
	}
}

func newTestUC(products *mockProductsRepo, orders *mockOrdersRepo) *Usecase {
	uc := New(products, orders, defaultLoyalty())
	_ = uc.Init(context.Background(), testLogger())
	return uc
}

func newTestUCWithLoyalty(products *mockProductsRepo, orders *mockOrdersRepo, loyalty *mockLoyaltySpender) *Usecase {
	uc := New(products, orders, loyalty)
	_ = uc.Init(context.Background(), testLogger())
	return uc
}

func intPtr(v int) *int { return &v }

// ── Product Tests ───────────────────────────────────────────────────────────

func TestCreateProduct(t *testing.T) {
	products := &mockProductsRepo{
		createProduct: func(_ context.Context, orgID int, req entity.CreateProductRequest) (*entity.MarketplaceProduct, error) {
			return &entity.MarketplaceProduct{ID: 1, OrgID: orgID, Name: req.Name, PricePoints: req.PricePoints}, nil
		},
	}
	uc := newTestUC(products, &mockOrdersRepo{})
	p, err := uc.CreateProduct(context.Background(), 1, entity.CreateProductRequest{
		Name: "Кофе", PricePoints: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "Кофе" || p.PricePoints != 100 {
		t.Fatalf("unexpected product: %+v", p)
	}
}

func TestGetProduct_NotFound(t *testing.T) {
	products := &mockProductsRepo{
		getProductByID: func(_ context.Context, _ int) (*entity.MarketplaceProduct, error) {
			return nil, nil
		},
	}
	uc := newTestUC(products, &mockOrdersRepo{})
	_, err := uc.GetProduct(context.Background(), 1, 999)
	if !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}

func TestGetProduct_WrongOrg(t *testing.T) {
	products := &mockProductsRepo{
		getProductByID: func(_ context.Context, _ int) (*entity.MarketplaceProduct, error) {
			return &entity.MarketplaceProduct{ID: 1, OrgID: 2}, nil
		},
	}
	uc := newTestUC(products, &mockOrdersRepo{})
	_, err := uc.GetProduct(context.Background(), 1, 1)
	if !errors.Is(err, ErrWrongOrg) {
		t.Fatalf("expected ErrWrongOrg, got %v", err)
	}
}

func TestDeleteProduct_NotFound(t *testing.T) {
	products := &mockProductsRepo{
		getProductByID: func(_ context.Context, _ int) (*entity.MarketplaceProduct, error) {
			return nil, nil
		},
	}
	uc := newTestUC(products, &mockOrdersRepo{})
	err := uc.DeleteProduct(context.Background(), 1, 999)
	if !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}

func TestDeleteProduct_WrongOrg(t *testing.T) {
	products := &mockProductsRepo{
		getProductByID: func(_ context.Context, _ int) (*entity.MarketplaceProduct, error) {
			return &entity.MarketplaceProduct{ID: 1, OrgID: 2}, nil
		},
	}
	uc := newTestUC(products, &mockOrdersRepo{})
	err := uc.DeleteProduct(context.Background(), 1, 1)
	if !errors.Is(err, ErrWrongOrg) {
		t.Fatalf("expected ErrWrongOrg, got %v", err)
	}
}

// ── Order Tests ─────────────────────────────────────────────────────────────

func TestPlaceOrder_EmptyItems(t *testing.T) {
	uc := newTestUC(&mockProductsRepo{}, &mockOrdersRepo{})
	_, err := uc.PlaceOrder(context.Background(), 1, entity.PlaceOrderRequest{
		ClientID: 10, ProgramID: 1, Items: []entity.PlaceOrderItem{},
	})
	if !errors.Is(err, ErrEmptyOrder) {
		t.Fatalf("expected ErrEmptyOrder, got %v", err)
	}
}

func TestPlaceOrder_ProductNotFound(t *testing.T) {
	products := &mockProductsRepo{
		getProductByID: func(_ context.Context, _ int) (*entity.MarketplaceProduct, error) {
			return nil, nil
		},
	}
	uc := newTestUC(products, &mockOrdersRepo{})
	_, err := uc.PlaceOrder(context.Background(), 1, entity.PlaceOrderRequest{
		ClientID: 10, ProgramID: 1, Items: []entity.PlaceOrderItem{{ProductID: 999, Quantity: 1}},
	})
	if !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}

func TestPlaceOrder_ProductInactive(t *testing.T) {
	products := &mockProductsRepo{
		getProductByID: func(_ context.Context, _ int) (*entity.MarketplaceProduct, error) {
			return &entity.MarketplaceProduct{ID: 1, OrgID: 1, IsActive: false, PricePoints: 50}, nil
		},
	}
	uc := newTestUC(products, &mockOrdersRepo{})
	_, err := uc.PlaceOrder(context.Background(), 1, entity.PlaceOrderRequest{
		ClientID: 10, ProgramID: 1, Items: []entity.PlaceOrderItem{{ProductID: 1, Quantity: 1}},
	})
	if !errors.Is(err, ErrProductInactive) {
		t.Fatalf("expected ErrProductInactive, got %v", err)
	}
}

func TestPlaceOrder_InsufficientStock(t *testing.T) {
	stock := 2
	products := &mockProductsRepo{
		getProductByID: func(_ context.Context, _ int) (*entity.MarketplaceProduct, error) {
			return &entity.MarketplaceProduct{ID: 1, OrgID: 1, IsActive: true, PricePoints: 50, Stock: &stock}, nil
		},
	}
	uc := newTestUC(products, &mockOrdersRepo{})
	_, err := uc.PlaceOrder(context.Background(), 1, entity.PlaceOrderRequest{
		ClientID: 10, ProgramID: 1, Items: []entity.PlaceOrderItem{{ProductID: 1, Quantity: 5}},
	})
	if !errors.Is(err, ErrInsufficientStock) {
		t.Fatalf("expected ErrInsufficientStock, got %v", err)
	}
}

func TestPlaceOrder_WrongOrg(t *testing.T) {
	products := &mockProductsRepo{
		getProductByID: func(_ context.Context, _ int) (*entity.MarketplaceProduct, error) {
			return &entity.MarketplaceProduct{ID: 1, OrgID: 2, IsActive: true, PricePoints: 50}, nil
		},
	}
	uc := newTestUC(products, &mockOrdersRepo{})
	_, err := uc.PlaceOrder(context.Background(), 1, entity.PlaceOrderRequest{
		ClientID: 10, ProgramID: 1, Items: []entity.PlaceOrderItem{{ProductID: 1, Quantity: 1}},
	})
	if !errors.Is(err, ErrWrongOrg) {
		t.Fatalf("expected ErrWrongOrg, got %v", err)
	}
}

func TestPlaceOrder_InsufficientPoints(t *testing.T) {
	products := &mockProductsRepo{
		getProductByID: func(_ context.Context, id int) (*entity.MarketplaceProduct, error) {
			return &entity.MarketplaceProduct{
				ID: id, OrgID: 1, Name: "Кофе", IsActive: true, PricePoints: 500, Stock: nil,
			}, nil
		},
	}
	loyalty := &mockLoyaltySpender{
		getBalance: func(_ context.Context, _, _ int) (*entity.ClientLoyalty, error) {
			return &entity.ClientLoyalty{Balance: 100}, nil
		},
	}
	uc := newTestUCWithLoyalty(products, &mockOrdersRepo{}, loyalty)
	_, err := uc.PlaceOrder(context.Background(), 1, entity.PlaceOrderRequest{
		ClientID: 10, ProgramID: 1, Items: []entity.PlaceOrderItem{{ProductID: 1, Quantity: 1}},
	})
	if !errors.Is(err, ErrInsufficientPoints) {
		t.Fatalf("expected ErrInsufficientPoints, got %v", err)
	}
}

func TestPlaceOrder_Success_Unlimited(t *testing.T) {
	products := &mockProductsRepo{
		getProductByID: func(_ context.Context, id int) (*entity.MarketplaceProduct, error) {
			return &entity.MarketplaceProduct{
				ID: id, OrgID: 1, Name: "Кофе", IsActive: true, PricePoints: 50, Stock: nil,
			}, nil
		},
	}
	var spentAmount float64
	loyalty := &mockLoyaltySpender{
		getBalance: func(_ context.Context, _, _ int) (*entity.ClientLoyalty, error) {
			return &entity.ClientLoyalty{Balance: 1000}, nil
		},
		spendPoints: func(_ context.Context, _, _ int, amount float64, _ string) (*entity.ClientLoyalty, error) {
			spentAmount = amount
			return &entity.ClientLoyalty{Balance: 1000 - amount}, nil
		},
	}
	var createdOrder *entity.MarketplaceOrder
	orders := &mockOrdersRepo{
		createOrder: func(_ context.Context, order *entity.MarketplaceOrder) error {
			order.ID = 1
			createdOrder = order
			return nil
		},
	}
	uc := newTestUCWithLoyalty(products, orders, loyalty)
	result, err := uc.PlaceOrder(context.Background(), 1, entity.PlaceOrderRequest{
		ClientID: 10, ProgramID: 1,
		Items: []entity.PlaceOrderItem{{ProductID: 5, Quantity: 3}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalPoints != 150 {
		t.Fatalf("expected 150 points, got %d", result.TotalPoints)
	}
	if spentAmount != 150 {
		t.Fatalf("expected loyalty spend of 150, got %.0f", spentAmount)
	}
	if len(createdOrder.Items) != 1 || createdOrder.Items[0].ProductName != "Кофе" {
		t.Fatalf("unexpected items: %+v", createdOrder.Items)
	}
}

func TestPlaceOrder_Success_WithStock(t *testing.T) {
	stock := 10
	var decremented bool
	products := &mockProductsRepo{
		getProductByID: func(_ context.Context, _ int) (*entity.MarketplaceProduct, error) {
			return &entity.MarketplaceProduct{
				ID: 1, OrgID: 1, Name: "Торт", IsActive: true, PricePoints: 200, Stock: &stock,
			}, nil
		},
		decrementStock: func(_ context.Context, _ int, _ int) error {
			decremented = true
			return nil
		},
	}
	orders := &mockOrdersRepo{
		createOrder: func(_ context.Context, order *entity.MarketplaceOrder) error {
			order.ID = 1
			return nil
		},
	}
	uc := newTestUCWithLoyalty(products, orders, defaultLoyalty())
	result, err := uc.PlaceOrder(context.Background(), 1, entity.PlaceOrderRequest{
		ClientID: 10, ProgramID: 1,
		Items: []entity.PlaceOrderItem{{ProductID: 1, Quantity: 2}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalPoints != 400 {
		t.Fatalf("expected 400 points, got %d", result.TotalPoints)
	}
	if !decremented {
		t.Fatal("stock should have been decremented")
	}
}

func TestGetOrder_NotFound(t *testing.T) {
	orders := &mockOrdersRepo{
		getOrderByID: func(_ context.Context, _ int) (*entity.MarketplaceOrder, error) {
			return nil, nil
		},
	}
	uc := newTestUC(&mockProductsRepo{}, orders)
	_, err := uc.GetOrder(context.Background(), 1, 999)
	if !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestUpdateOrderStatus_WrongOrg(t *testing.T) {
	orders := &mockOrdersRepo{
		getOrderByID: func(_ context.Context, _ int) (*entity.MarketplaceOrder, error) {
			return &entity.MarketplaceOrder{ID: 1, OrgID: 2}, nil
		},
	}
	uc := newTestUC(&mockProductsRepo{}, orders)
	err := uc.UpdateOrderStatus(context.Background(), 1, 1, "completed")
	if !errors.Is(err, ErrWrongOrg) {
		t.Fatalf("expected ErrWrongOrg, got %v", err)
	}
}
