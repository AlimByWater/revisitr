package marketplace

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"revisitr/internal/entity"
)

var (
	ErrProductNotFound    = errors.New("product not found")
	ErrProductInactive    = errors.New("product is not active")
	ErrInsufficientStock  = errors.New("insufficient stock")
	ErrInsufficientPoints = errors.New("insufficient loyalty points")
	ErrOrderNotFound      = errors.New("order not found")
	ErrEmptyOrder         = errors.New("order must have at least one item")
	ErrWrongOrg           = errors.New("product does not belong to organization")
)

type productsRepo interface {
	CreateProduct(ctx context.Context, orgID int, req entity.CreateProductRequest) (*entity.MarketplaceProduct, error)
	GetProducts(ctx context.Context, orgID int) ([]entity.MarketplaceProduct, error)
	GetActiveProducts(ctx context.Context, orgID int) ([]entity.MarketplaceProduct, error)
	GetProductByID(ctx context.Context, id int) (*entity.MarketplaceProduct, error)
	UpdateProduct(ctx context.Context, id int, req entity.UpdateProductRequest) (*entity.MarketplaceProduct, error)
	DeleteProduct(ctx context.Context, id int) error
	DecrementStock(ctx context.Context, productID int, qty int) error
}

type ordersRepo interface {
	CreateOrder(ctx context.Context, order *entity.MarketplaceOrder) error
	GetOrders(ctx context.Context, orgID int) ([]entity.MarketplaceOrder, error)
	GetOrderByID(ctx context.Context, id int) (*entity.MarketplaceOrder, error)
	GetOrdersByClient(ctx context.Context, clientID int) ([]entity.MarketplaceOrder, error)
	UpdateOrderStatus(ctx context.Context, id int, status string) error
	GetStats(ctx context.Context, orgID int) (*entity.MarketplaceStats, error)
}

type loyaltySpender interface {
	SpendPoints(ctx context.Context, clientID, programID int, amount float64, description string) (*entity.ClientLoyalty, error)
	GetBalance(ctx context.Context, clientID, programID int) (*entity.ClientLoyalty, error)
}

type Usecase struct {
	logger   *slog.Logger
	products productsRepo
	orders   ordersRepo
	loyalty  loyaltySpender
}

func New(products productsRepo, orders ordersRepo, loyalty loyaltySpender) *Usecase {
	return &Usecase{products: products, orders: orders, loyalty: loyalty}
}

func (uc *Usecase) Init(_ context.Context, logger *slog.Logger) error {
	uc.logger = logger
	return nil
}

// ── Products ─────────────────────────────────────────────────────────────────

func (uc *Usecase) CreateProduct(ctx context.Context, orgID int, req entity.CreateProductRequest) (*entity.MarketplaceProduct, error) {
	return uc.products.CreateProduct(ctx, orgID, req)
}

func (uc *Usecase) GetProducts(ctx context.Context, orgID int) ([]entity.MarketplaceProduct, error) {
	return uc.products.GetProducts(ctx, orgID)
}

func (uc *Usecase) GetActiveProducts(ctx context.Context, orgID int) ([]entity.MarketplaceProduct, error) {
	return uc.products.GetActiveProducts(ctx, orgID)
}

func (uc *Usecase) GetProduct(ctx context.Context, orgID int, productID int) (*entity.MarketplaceProduct, error) {
	p, err := uc.products.GetProductByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrProductNotFound
	}
	if p.OrgID != orgID {
		return nil, ErrWrongOrg
	}
	return p, nil
}

func (uc *Usecase) UpdateProduct(ctx context.Context, orgID int, productID int, req entity.UpdateProductRequest) (*entity.MarketplaceProduct, error) {
	existing, err := uc.products.GetProductByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrProductNotFound
	}
	if existing.OrgID != orgID {
		return nil, ErrWrongOrg
	}
	return uc.products.UpdateProduct(ctx, productID, req)
}

func (uc *Usecase) DeleteProduct(ctx context.Context, orgID int, productID int) error {
	existing, err := uc.products.GetProductByID(ctx, productID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrProductNotFound
	}
	if existing.OrgID != orgID {
		return ErrWrongOrg
	}
	return uc.products.DeleteProduct(ctx, productID)
}

// ── Orders ───────────────────────────────────────────────────────────────────

func (uc *Usecase) PlaceOrder(ctx context.Context, orgID int, req entity.PlaceOrderRequest) (*entity.MarketplaceOrder, error) {
	if len(req.Items) == 0 {
		return nil, ErrEmptyOrder
	}

	var totalPoints int
	var orderItems entity.MarketplaceOrderItems

	for _, item := range req.Items {
		product, err := uc.products.GetProductByID(ctx, item.ProductID)
		if err != nil {
			return nil, err
		}
		if product == nil {
			return nil, ErrProductNotFound
		}
		if product.OrgID != orgID {
			return nil, ErrWrongOrg
		}
		if !product.IsActive {
			return nil, ErrProductInactive
		}
		if product.Stock != nil && *product.Stock < item.Quantity {
			return nil, ErrInsufficientStock
		}

		linePoints := product.PricePoints * item.Quantity
		totalPoints += linePoints

		orderItems = append(orderItems, entity.MarketplaceOrderItem{
			ProductID:   product.ID,
			ProductName: product.Name,
			Quantity:    item.Quantity,
			Points:      linePoints,
		})
	}

	// Check loyalty balance before proceeding
	balance, err := uc.loyalty.GetBalance(ctx, req.ClientID, req.ProgramID)
	if err != nil {
		return nil, fmt.Errorf("check loyalty balance: %w", err)
	}
	if balance.Balance < float64(totalPoints) {
		return nil, ErrInsufficientPoints
	}

	// Decrement stock for limited items
	for _, item := range req.Items {
		product, _ := uc.products.GetProductByID(ctx, item.ProductID)
		if product != nil && product.Stock != nil {
			if err := uc.products.DecrementStock(ctx, item.ProductID, item.Quantity); err != nil {
				return nil, ErrInsufficientStock
			}
		}
	}

	// Deduct loyalty points
	desc := fmt.Sprintf("Маркетплейс: заказ на %d баллов", totalPoints)
	if _, err := uc.loyalty.SpendPoints(ctx, req.ClientID, req.ProgramID, float64(totalPoints), desc); err != nil {
		return nil, fmt.Errorf("spend loyalty points: %w", err)
	}

	order := &entity.MarketplaceOrder{
		OrgID:       orgID,
		ClientID:    req.ClientID,
		Status:      "confirmed",
		TotalPoints: totalPoints,
		Items:       orderItems,
		Note:        req.Note,
	}

	if err := uc.orders.CreateOrder(ctx, order); err != nil {
		return nil, err
	}

	uc.logger.Info("marketplace order placed",
		"org_id", orgID, "client_id", req.ClientID,
		"total_points", totalPoints, "items", len(req.Items))

	return order, nil
}

func (uc *Usecase) GetOrders(ctx context.Context, orgID int) ([]entity.MarketplaceOrder, error) {
	return uc.orders.GetOrders(ctx, orgID)
}

func (uc *Usecase) GetOrder(ctx context.Context, orgID int, orderID int) (*entity.MarketplaceOrder, error) {
	o, err := uc.orders.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, ErrOrderNotFound
	}
	if o.OrgID != orgID {
		return nil, ErrWrongOrg
	}
	return o, nil
}

func (uc *Usecase) UpdateOrderStatus(ctx context.Context, orgID int, orderID int, status string) error {
	o, err := uc.orders.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}
	if o == nil {
		return ErrOrderNotFound
	}
	if o.OrgID != orgID {
		return ErrWrongOrg
	}
	return uc.orders.UpdateOrderStatus(ctx, orderID, status)
}

func (uc *Usecase) GetStats(ctx context.Context, orgID int) (*entity.MarketplaceStats, error) {
	return uc.orders.GetStats(ctx, orgID)
}
