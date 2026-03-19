package pos

import (
	"context"
	"testing"
	"time"

	"revisitr/internal/entity"
)

func TestMockProvider_TestConnection(t *testing.T) {
	p, err := NewMockProvider()
	if err != nil {
		t.Fatalf("NewMockProvider: %v", err)
	}
	if err := p.TestConnection(context.Background()); err != nil {
		t.Errorf("TestConnection should succeed, got: %v", err)
	}
}

func TestMockProvider_GetCustomers(t *testing.T) {
	p, err := NewMockProvider()
	if err != nil {
		t.Fatalf("NewMockProvider: %v", err)
	}

	customers, err := p.GetCustomers(context.Background(), CustomerListOpts{Limit: 10})
	if err != nil {
		t.Fatalf("GetCustomers: %v", err)
	}
	if len(customers) != 10 {
		t.Errorf("expected 10 customers, got %d", len(customers))
	}

	c := customers[0]
	if c.ExternalID == "" || c.Phone == "" || c.Name == "" {
		t.Errorf("customer fields should be populated: %+v", c)
	}
}

func TestMockProvider_GetCustomers_Search(t *testing.T) {
	p, _ := NewMockProvider()

	customers, err := p.GetCustomers(context.Background(), CustomerListOpts{
		Limit:  50,
		Search: "Александр",
	})
	if err != nil {
		t.Fatalf("GetCustomers: %v", err)
	}
	if len(customers) == 0 {
		t.Error("expected at least one customer matching 'Александр'")
	}
}

func TestMockProvider_GetCustomers_Pagination(t *testing.T) {
	p, _ := NewMockProvider()

	all, _ := p.GetCustomers(context.Background(), CustomerListOpts{Limit: 100})
	page1, _ := p.GetCustomers(context.Background(), CustomerListOpts{Limit: 5, Offset: 0})
	page2, _ := p.GetCustomers(context.Background(), CustomerListOpts{Limit: 5, Offset: 5})

	if len(all) != 20 {
		t.Errorf("expected 20 total customers, got %d", len(all))
	}
	if len(page1) != 5 || len(page2) != 5 {
		t.Errorf("expected 5 per page, got %d and %d", len(page1), len(page2))
	}
	if page1[0].ExternalID == page2[0].ExternalID {
		t.Error("pages should not overlap")
	}
}

func TestMockProvider_GetOrders(t *testing.T) {
	p, _ := NewMockProvider()

	now := time.Now()
	orders, err := p.GetOrders(context.Background(), now.Add(-31*24*time.Hour), now)
	if err != nil {
		t.Fatalf("GetOrders: %v", err)
	}
	if len(orders) == 0 {
		t.Error("expected orders in the last 31 days")
	}

	o := orders[0]
	if o.ExternalID == "" || o.Total <= 0 || len(o.Items) == 0 {
		t.Errorf("order fields should be populated: %+v", o)
	}
}

func TestMockProvider_GetOrders_DateFilter(t *testing.T) {
	p, _ := NewMockProvider()

	now := time.Now()
	narrow, _ := p.GetOrders(context.Background(), now.Add(-1*time.Hour), now)
	wide, _ := p.GetOrders(context.Background(), now.Add(-31*24*time.Hour), now)

	if len(narrow) > len(wide) {
		t.Error("narrow window should return fewer or equal orders than wide window")
	}
}

func TestMockProvider_GetMenu(t *testing.T) {
	p, _ := NewMockProvider()

	menu, err := p.GetMenu(context.Background())
	if err != nil {
		t.Fatalf("GetMenu: %v", err)
	}
	if menu == nil || len(menu.Categories) == 0 {
		t.Fatal("expected non-empty menu")
	}

	totalItems := 0
	for _, cat := range menu.Categories {
		if cat.Name == "" {
			t.Error("category name should not be empty")
		}
		totalItems += len(cat.Items)
		for _, item := range cat.Items {
			if item.Name == "" || item.Price <= 0 || item.ExternalID == "" {
				t.Errorf("menu item fields should be populated: %+v", item)
			}
		}
	}
	if totalItems == 0 {
		t.Error("expected menu items")
	}
}

func TestMockProvider_Deterministic(t *testing.T) {
	p1, _ := NewMockProvider()
	p2, _ := NewMockProvider()

	c1, _ := p1.GetCustomers(context.Background(), CustomerListOpts{Limit: 5})
	c2, _ := p2.GetCustomers(context.Background(), CustomerListOpts{Limit: 5})

	for i := range c1 {
		if c1[i].ExternalID != c2[i].ExternalID || c1[i].Name != c2[i].Name {
			t.Errorf("mock provider should be deterministic: customer %d differs", i)
		}
	}
}

func TestNewProvider_Factory(t *testing.T) {
	tests := []struct {
		name    string
		intType string
		wantErr bool
	}{
		{"mock", "mock", false},
		{"unknown", "unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			intg := &entity.Integration{Type: tt.intType}
			_, err := NewProvider(intg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewProvider(%s) error = %v, wantErr %v", tt.intType, err, tt.wantErr)
			}
		})
	}
}
