package pos

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// MockProvider is an in-memory POS provider for development and demos.
type MockProvider struct {
	customers []POSCustomer
	orders    []POSOrder
	menu      *POSMenu
	mu        sync.RWMutex
}

func NewMockProvider() (*MockProvider, error) {
	p := &MockProvider{}
	p.seedData()
	return p, nil
}

func (p *MockProvider) TestConnection(_ context.Context) error {
	return nil
}

func (p *MockProvider) GetCustomers(_ context.Context, opts CustomerListOpts) ([]POSCustomer, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var filtered []POSCustomer
	for _, c := range p.customers {
		if opts.Search != "" {
			q := strings.ToLower(opts.Search)
			if !strings.Contains(strings.ToLower(c.Name), q) &&
				!strings.Contains(c.Phone, q) {
				continue
			}
		}
		filtered = append(filtered, c)
	}

	if opts.Offset >= len(filtered) {
		return nil, nil
	}
	end := opts.Offset + opts.Limit
	if end > len(filtered) || opts.Limit == 0 {
		end = len(filtered)
	}
	return filtered[opts.Offset:end], nil
}

func (p *MockProvider) GetOrders(_ context.Context, from, to time.Time) ([]POSOrder, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var result []POSOrder
	for _, o := range p.orders {
		if (o.OrderedAt.Equal(from) || o.OrderedAt.After(from)) &&
			(o.OrderedAt.Equal(to) || o.OrderedAt.Before(to)) {
			result = append(result, o)
		}
	}
	return result, nil
}

func (p *MockProvider) GetMenu(_ context.Context) (*POSMenu, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.menu, nil
}

func (p *MockProvider) GetDailyAggregates(_ context.Context, from, to time.Time) ([]POSDailyAggregate, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Aggregate orders by date
	daily := make(map[string]*POSDailyAggregate)
	for _, o := range p.orders {
		if o.Status == "cancelled" {
			continue
		}
		if o.OrderedAt.Before(from) || o.OrderedAt.After(to) {
			continue
		}
		dateKey := o.OrderedAt.Format("2006-01-02")
		agg, ok := daily[dateKey]
		if !ok {
			d, _ := time.Parse("2006-01-02", dateKey)
			agg = &POSDailyAggregate{Date: d}
			daily[dateKey] = agg
		}
		agg.Revenue += o.Total
		agg.TxCount++
		agg.GuestCount++ // 1 guest per order in mock
		if o.CustomerPhone != "" {
			agg.Phones = append(agg.Phones, o.CustomerPhone)
		}
	}

	result := make([]POSDailyAggregate, 0, len(daily))
	for _, agg := range daily {
		if agg.TxCount > 0 {
			agg.AvgCheck = agg.Revenue / float64(agg.TxCount)
		}
		result = append(result, *agg)
	}
	return result, nil
}

// --- Seed data generation ---

var (
	firstNames = []string{
		"Александр", "Мария", "Дмитрий", "Елена", "Сергей",
		"Анна", "Андрей", "Ольга", "Михаил", "Наталья",
		"Иван", "Татьяна", "Павел", "Екатерина", "Артём",
		"Юлия", "Николай", "Светлана", "Алексей", "Ирина",
	}
	lastNames = []string{
		"Иванов", "Петрова", "Сидоров", "Козлова", "Смирнов",
		"Волкова", "Новиков", "Морозова", "Лебедев", "Соколова",
		"Попов", "Кузнецова", "Фёдоров", "Егорова", "Зайцев",
		"Павлова", "Семёнов", "Голубева", "Степанов", "Виноградова",
	}
	menuData = map[string][]struct {
		name  string
		price float64
	}{
		"Кофе": {
			{"Американо", 180}, {"Капучино", 250}, {"Латте", 280},
			{"Эспрессо", 150}, {"Раф", 320}, {"Флэт Уайт", 300},
		},
		"Напитки": {
			{"Чай чёрный", 150}, {"Чай зелёный", 150}, {"Лимонад", 220},
			{"Свежевыжатый апельсин", 350}, {"Морс клюквенный", 200},
		},
		"Завтраки": {
			{"Яичница с беконом", 380}, {"Каша овсяная", 250}, {"Сырники", 320},
			{"Блины с лососем", 450}, {"Тост с авокадо", 400}, {"Гранола с йогуртом", 300},
		},
		"Основные блюда": {
			{"Паста Карбонара", 480}, {"Салат Цезарь", 420}, {"Бургер классический", 520},
			{"Том Ям", 450}, {"Стейк Рибай", 1200}, {"Ризотто с грибами", 480},
		},
		"Десерты": {
			{"Чизкейк", 350}, {"Тирамису", 380}, {"Наполеон", 300},
			{"Мороженое", 200}, {"Брауни", 280},
		},
		"Бар": {
			{"Пиво светлое 0.5", 350}, {"Вино красное бокал", 450}, {"Виски", 500},
			{"Мохито", 420}, {"Апероль Шприц", 480}, {"Маргарита", 450},
		},
	}
	waiterNames = []string{
		"Алина", "Владимир", "Кристина", "Роман", "Дарья",
	}
)

func (p *MockProvider) seedData() {
	r := rand.New(rand.NewSource(42)) // deterministic for consistency

	// Customers
	p.customers = make([]POSCustomer, 20)
	for i := range p.customers {
		fn := firstNames[i%len(firstNames)]
		ln := lastNames[i%len(lastNames)]
		p.customers[i] = POSCustomer{
			ExternalID: fmt.Sprintf("mock-cust-%d", i+1),
			Phone:      fmt.Sprintf("+7916%07d", 1000000+i*37),
			Name:       fn + " " + ln,
			Balance:    float64(r.Intn(5000)),
			CardNumber: fmt.Sprintf("CARD-%06d", 100000+i),
		}
		if i%3 == 0 {
			bday := time.Date(1985+r.Intn(15), time.Month(1+r.Intn(12)), 1+r.Intn(28), 0, 0, 0, 0, time.UTC)
			p.customers[i].Birthday = &bday
		}
	}

	// Menu
	p.menu = &POSMenu{}
	itemIdx := 0
	for cat, items := range menuData {
		mc := MenuCategory{Name: cat}
		for _, it := range items {
			itemIdx++
			mc.Items = append(mc.Items, MenuItem{
				ExternalID: fmt.Sprintf("mock-item-%d", itemIdx),
				Name:       it.name,
				Price:      it.price,
			})
		}
		p.menu.Categories = append(p.menu.Categories, mc)
	}

	// Orders (last 30 days)
	now := time.Now()
	p.orders = make([]POSOrder, 0, 50)
	statuses := []string{"closed", "closed", "closed", "closed", "cancelled"}
	for i := 0; i < 50; i++ {
		hoursAgo := r.Intn(30 * 24)
		orderTime := now.Add(-time.Duration(hoursAgo) * time.Hour)

		numItems := 1 + r.Intn(4)
		items := make([]POSOrderItem, numItems)
		total := 0.0
		for j := 0; j < numItems; j++ {
			// pick random category and item
			catIdx := r.Intn(len(p.menu.Categories))
			cat := p.menu.Categories[catIdx]
			itemJ := r.Intn(len(cat.Items))
			mi := cat.Items[itemJ]
			qty := 1 + r.Intn(3)
			items[j] = POSOrderItem{
				Name:     mi.Name,
				Quantity: qty,
				Price:    mi.Price,
				Category: cat.Name,
			}
			total += mi.Price * float64(qty)
		}

		discount := 0.0
		if r.Intn(5) == 0 {
			discount = float64(r.Intn(int(total/5))) // up to 20%
		}

		o := POSOrder{
			ExternalID:    fmt.Sprintf("mock-order-%d", i+1),
			CustomerPhone: p.customers[r.Intn(len(p.customers))].Phone,
			Items:         items,
			Total:         total - discount,
			Discount:      discount,
			OrderedAt:     orderTime,
			Status:        statuses[r.Intn(len(statuses))],
			TableNum:      fmt.Sprintf("%d", 1+r.Intn(20)),
			WaiterName:    waiterNames[r.Intn(len(waiterNames))],
		}
		p.orders = append(p.orders, o)
	}
}
