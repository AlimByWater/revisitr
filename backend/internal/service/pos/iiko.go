package pos

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"revisitr/internal/entity"
)

const (
	iikoDefaultBaseURL     = "https://api-ru.iiko.services/api/1"
	iikoDefaultHTTPTimeout = 30 * time.Second
	iikoTokenTTL           = 50 * time.Minute
)

// IikoProvider talks to iikoCloud Transport API.
//
// Auth: POST /access_token {apiLogin} → {token}. Token TTL 1h, cached 50m.
// All other calls require Authorization: Bearer <token>.
type IikoProvider struct {
	baseURL        string
	apiLogin       string
	orgID          string
	externalMenuID string

	httpClient *http.Client

	mu          sync.Mutex
	token       string
	tokenExpiry time.Time
}

// IikoOption configures optional dependencies for testing.
type IikoOption func(*IikoProvider)

// WithIikoHTTPClient overrides the default *http.Client (used in tests).
func WithIikoHTTPClient(c *http.Client) IikoOption {
	return func(p *IikoProvider) { p.httpClient = c }
}

// WithIikoBaseURL overrides the default base URL (used in tests).
func WithIikoBaseURL(u string) IikoOption {
	return func(p *IikoProvider) { p.baseURL = u }
}

func NewIikoProvider(cfg entity.IntegrationConfig, opts ...IikoOption) (*IikoProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("iiko: api_key (apiLogin) is required")
	}
	if cfg.OrgID == "" {
		return nil, fmt.Errorf("iiko: org_id is required")
	}

	baseURL := cfg.APIURL
	if baseURL == "" {
		baseURL = iikoDefaultBaseURL
	}

	p := &IikoProvider{
		baseURL:        baseURL,
		apiLogin:       cfg.APIKey,
		orgID:          cfg.OrgID,
		externalMenuID: cfg.ExternalMenuID,
		httpClient:     &http.Client{Timeout: iikoDefaultHTTPTimeout},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p, nil
}

// IikoAPIError is returned for non-2xx responses with the iiko payload preserved.
type IikoAPIError struct {
	Status int
	Path   string
	Body   string
}

func (e *IikoAPIError) Error() string {
	return fmt.Sprintf("iiko: %s returned %d: %s", e.Path, e.Status, e.Body)
}

// IsAuthError reports whether the iiko response was 401.
func IsAuthError(err error) bool {
	var apiErr *IikoAPIError
	if errors.As(err, &apiErr) {
		return apiErr.Status == http.StatusUnauthorized
	}
	return false
}

// getToken returns a cached token, refreshing if missing or expired.
func (p *IikoProvider) getToken(ctx context.Context) (string, error) {
	p.mu.Lock()
	if p.token != "" && time.Now().Before(p.tokenExpiry) {
		t := p.token
		p.mu.Unlock()
		return t, nil
	}
	p.mu.Unlock()

	tok, err := p.fetchToken(ctx)
	if err != nil {
		return "", err
	}

	p.mu.Lock()
	p.token = tok
	p.tokenExpiry = time.Now().Add(iikoTokenTTL)
	p.mu.Unlock()
	return tok, nil
}

// invalidateToken clears the cached token (called on 401).
func (p *IikoProvider) invalidateToken() {
	p.mu.Lock()
	p.token = ""
	p.tokenExpiry = time.Time{}
	p.mu.Unlock()
}

type iikoTokenResponse struct {
	Token         string `json:"token"`
	CorrelationID string `json:"correlationId"`
}

func (p *IikoProvider) fetchToken(ctx context.Context) (string, error) {
	body, err := json.Marshal(map[string]string{"apiLogin": p.apiLogin})
	if err != nil {
		return "", fmt.Errorf("marshal token request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/access_token", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", &IikoAPIError{Status: resp.StatusCode, Path: "/access_token", Body: string(respBody)}
	}

	var out iikoTokenResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}
	if out.Token == "" {
		return "", fmt.Errorf("iiko: empty token in response")
	}
	return out.Token, nil
}

// doRequest performs a POST to the given path with JSON body and decodes JSON response into out.
// Retries once on 401 by refreshing the token.
func (p *IikoProvider) doRequest(ctx context.Context, path string, body, out any) error {
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		token, err := p.getToken(ctx)
		if err != nil {
			return fmt.Errorf("get token: %w", err)
		}

		err = p.doRequestOnce(ctx, path, token, body, out)
		if err == nil {
			return nil
		}
		lastErr = err

		if !IsAuthError(err) {
			return err
		}
		// 401 → refresh token and retry once
		p.invalidateToken()
	}
	return lastErr
}

func (p *IikoProvider) doRequestOnce(ctx context.Context, path, token string, body, out any) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpointURL(path), reqBody)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http %s: %w", path, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &IikoAPIError{Status: resp.StatusCode, Path: path, Body: string(respBody)}
	}

	if out == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decode %s response: %w", path, err)
	}
	return nil
}

func (p *IikoProvider) endpointURL(path string) string {
	if strings.HasPrefix(path, "/api/") {
		if idx := strings.Index(p.baseURL, "/api/"); idx >= 0 {
			return strings.TrimRight(p.baseURL[:idx], "/") + path
		}
	}
	return strings.TrimRight(p.baseURL, "/") + path
}

type iikoOrganizationsResponse struct {
	Organizations []iikoOrganization `json:"organizations"`
}

type iikoOrganization struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsDeleted bool   `json:"isDeleted"`
}

func (p *IikoProvider) TestConnection(ctx context.Context) error {
	var out iikoOrganizationsResponse
	if err := p.doRequest(ctx, "/organizations", map[string]any{
		"organizationIds":      nil,
		"returnAdditionalInfo": true,
	}, &out); err != nil {
		return err
	}
	for _, org := range out.Organizations {
		if org.ID == p.orgID && !org.IsDeleted {
			return nil
		}
	}
	return fmt.Errorf("iiko: organization %s not found", p.orgID)
}

// ListOrganizations returns all organizations available to this apiLogin, used
// during onboarding so the user can pick org_id instead of pasting a UUID.
func (p *IikoProvider) ListOrganizations(ctx context.Context) ([]POSOrganization, error) {
	var out iikoOrganizationsResponse
	if err := p.doRequest(ctx, "/organizations", map[string]any{
		"organizationIds":      nil,
		"returnAdditionalInfo": true,
	}, &out); err != nil {
		return nil, err
	}
	orgs := make([]POSOrganization, 0, len(out.Organizations))
	for _, org := range out.Organizations {
		if org.IsDeleted || org.ID == "" {
			continue
		}
		orgs = append(orgs, POSOrganization{ID: org.ID, Name: org.Name})
	}
	return orgs, nil
}

// ListExternalMenus returns the external menus available to this apiLogin so the
// user can optionally pick one during onboarding. Returns nil (not an error) when
// the menu listing is not permitted for this apiLogin/tariff.
func (p *IikoProvider) ListExternalMenus(ctx context.Context) ([]POSExternalMenu, error) {
	var out iikoExternalMenuListResponse
	if err := p.doRequest(ctx, "/api/2/menu", map[string]any{}, &out); err != nil {
		if iikoLoyaltyUnavailable(err) {
			return nil, nil
		}
		return nil, err
	}
	menus := make([]POSExternalMenu, 0, len(out.ExternalMenus))
	for _, m := range out.ExternalMenus {
		if m.ID == "" {
			continue
		}
		menus = append(menus, POSExternalMenu(m))
	}
	return menus, nil
}

type iikoExternalMenuListResponse struct {
	ExternalMenus []iikoExternalMenuListItem `json:"externalMenus"`
}

type iikoExternalMenuListItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetCustomers looks up a loyalty customer by phone via iiko's loyalty API.
//
// iiko Cloud's Transport API has no bulk "list customers" endpoint; customers
// are resolved one at a time by phone or card. So GetCustomers treats opts.Search
// as a phone number and returns the matching customer (if any). When the loyalty
// (iikoCard) module is not enabled for the organization, iiko answers 400/401 —
// this is treated as "no loyalty data available" and returns an empty slice
// rather than an error, so order/aggregate sync still succeeds.
func (p *IikoProvider) GetCustomers(ctx context.Context, opts CustomerListOpts) ([]POSCustomer, error) {
	phone := strings.TrimSpace(opts.Search)
	if phone == "" {
		return nil, nil
	}

	var out iikoCustomerInfo
	err := p.doRequest(ctx, "/loyalty/iiko/customer/info", map[string]any{
		"organizationId": p.orgID,
		"type":           "phone",
		"phone":          phone,
	}, &out)
	if err != nil {
		if iikoLoyaltyUnavailable(err) {
			return nil, nil
		}
		return nil, err
	}
	if out.ID == "" {
		return nil, nil
	}
	return []POSCustomer{mapIikoCustomer(out)}, nil
}

// iikoLoyaltyUnavailable reports whether the error means the loyalty/iikoCard
// module is simply not available (not bound, wrong CRM, no rights), as opposed to
// a transient failure. Such cases degrade to "no data" instead of failing sync.
func iikoLoyaltyUnavailable(err error) bool {
	var apiErr *IikoAPIError
	if !errors.As(err, &apiErr) {
		return false
	}
	switch apiErr.Status {
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		return true
	}
	return false
}

type iikoCustomerInfo struct {
	ID             string               `json:"id"`
	Name           string               `json:"name"`
	Surname        string               `json:"surname"`
	Phone          string               `json:"phone"`
	Email          string               `json:"email"`
	Birthday       string               `json:"birthday"`
	Cards          []iikoCustomerCard   `json:"cards"`
	WalletBalances []iikoCustomerWallet `json:"walletBalances"`
}

type iikoCustomerCard struct {
	Number string `json:"number"`
}

type iikoCustomerWallet struct {
	Balance float64 `json:"balance"`
}

func mapIikoCustomer(in iikoCustomerInfo) POSCustomer {
	name := strings.TrimSpace(in.Name + " " + in.Surname)

	var balance float64
	for _, w := range in.WalletBalances {
		balance += w.Balance
	}

	var card string
	if len(in.Cards) > 0 {
		card = in.Cards[0].Number
	}

	c := POSCustomer{
		ExternalID: in.ID,
		Phone:      in.Phone,
		Name:       name,
		Email:      in.Email,
		Balance:    balance,
		CardNumber: card,
	}
	if bday := parseIikoTime(in.Birthday); !bday.IsZero() {
		c.Birthday = &bday
	}
	return c
}

// iikoDeliveriesWindow is the max span per /deliveries request. iiko rejects
// wide windows with 422 TOO_MANY_DATA_REQUESTED, so the range is chunked.
const iikoDeliveriesWindow = 24 * time.Hour

// iikoOrderWindowPad widens the deliveries query window on both ends to absorb
// the offset between the server clock (UTC) and the organization's local
// timezone. iiko evaluates deliveryDateFrom/To against each order's delivery
// date expressed in the org's local time, while we send UTC timestamps. Without
// the pad, an order whose local delivery date is within a few hours of "now"
// falls past the UTC upper bound (and, for UTC-ahead orgs, before the lower
// bound) and is silently skipped. 24h comfortably covers every real UTC offset
// (max +14/-12h). UpsertOrder is idempotent by external id, so the extra rows
// fetched from the padding are deduplicated on write.
const iikoOrderWindowPad = 24 * time.Hour

func (p *IikoProvider) GetOrders(ctx context.Context, from, to time.Time) ([]POSOrder, error) {
	deliveries, err := p.fetchDeliveriesChunked(ctx, from.Add(-iikoOrderWindowPad), to.Add(iikoOrderWindowPad))
	if err != nil {
		return nil, err
	}

	orders := make([]POSOrder, 0, len(deliveries))
	for _, orderInfo := range deliveries {
		if order, ok := mapIikoDeliveryOrder(orderInfo); ok {
			orders = append(orders, order)
		}
	}
	return orders, nil
}

// fetchDeliveriesChunked splits [from, to] into windows no wider than
// iikoDeliveriesWindow and concatenates the deliveries from each request,
// avoiding the iiko 422 TOO_MANY_DATA_REQUESTED limit on wide ranges.
func (p *IikoProvider) fetchDeliveriesChunked(ctx context.Context, from, to time.Time) ([]iikoDeliveryOrder, error) {
	var all []iikoDeliveryOrder
	for start := from; start.Before(to); start = start.Add(iikoDeliveriesWindow) {
		end := start.Add(iikoDeliveriesWindow)
		if end.After(to) {
			end = to
		}
		// iiko compares dates at millisecond precision and rejects a window
		// whose ends render to the same instant ("deliveryDateFrom must be less
		// than deliveryDateTo"). Skip such a degenerate trailing window — it can
		// hold no orders anyway.
		if iikoDateTime(start) == iikoDateTime(end) {
			continue
		}
		chunk, err := p.fetchDeliveries(ctx, start, end)
		if err != nil {
			return nil, err
		}
		all = append(all, chunk...)
	}
	return all, nil
}

// fetchDeliveries performs a single /deliveries request for the given window.
func (p *IikoProvider) fetchDeliveries(ctx context.Context, from, to time.Time) ([]iikoDeliveryOrder, error) {
	var out iikoDeliveriesResponse
	err := p.doRequest(ctx, "/deliveries/by_delivery_date_and_status", map[string]any{
		"organizationIds":  []string{p.orgID},
		"deliveryDateFrom": iikoDateTime(from),
		"deliveryDateTo":   iikoDateTime(to),
		"statuses":         []string{"Closed", "Delivered"},
	}, &out)
	if err != nil {
		return nil, err
	}

	var orders []iikoDeliveryOrder
	for _, orgOrders := range out.OrdersByOrganizations {
		orders = append(orders, orgOrders.Orders...)
	}
	return orders, nil
}

func (p *IikoProvider) GetMenu(ctx context.Context) (*POSMenu, error) {
	menu, err := p.getNomenclatureMenu(ctx)
	if err != nil || p.externalMenuID == "" {
		return menu, err
	}

	external, err := p.getExternalMenu(ctx)
	if err != nil {
		return nil, err
	}
	if menu == nil || len(menu.Categories) == 0 {
		return external, nil
	}
	return enrichNomenclatureMenu(menu, external), nil
}

// enrichNomenclatureMenu preserves iiko nomenclature as the canonical product
// catalogue while using the external menu only for client-facing enrichment.
func enrichNomenclatureMenu(menu, external *POSMenu) *POSMenu {
	if menu == nil || external == nil {
		return menu
	}

	externalItems := make(map[string]MenuItem)
	for _, category := range external.Categories {
		for _, item := range category.Items {
			externalItems[item.ExternalID] = item
		}
	}

	for categoryIndex := range menu.Categories {
		for itemIndex := range menu.Categories[categoryIndex].Items {
			item := &menu.Categories[categoryIndex].Items[itemIndex]
			externalItem, ok := externalItems[item.ExternalID]
			if !ok {
				continue
			}
			if externalItem.Price > 0 {
				item.Price = externalItem.Price
			}
			if externalItem.Description != "" {
				item.Description = externalItem.Description
			}
			if externalItem.ImageURL != "" {
				item.ImageURL = externalItem.ImageURL
			}
		}
	}
	return menu
}

type iikoExternalMenuResponse struct {
	ItemCategories []iikoExternalMenuCategory `json:"itemCategories"`
}

type iikoExternalMenuCategory struct {
	Name     string                 `json:"name"`
	IsHidden bool                   `json:"isHidden"`
	Deleted  bool                   `json:"deleted"`
	Items    []iikoExternalMenuItem `json:"items"`
}

type iikoExternalMenuItem struct {
	ID          string                     `json:"id"`
	ItemID      string                     `json:"itemId"`
	IikoItemID  string                     `json:"iikoItemId"`
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	ImageLinks  []string                   `json:"imageLinks"`
	IsHidden    bool                       `json:"isHidden"`
	Deleted     bool                       `json:"deleted"`
	ItemSizes   []iikoExternalMenuItemSize `json:"itemSizes"`
}

type iikoExternalMenuItemSize struct {
	Price  *float64                `json:"price"`
	Prices []iikoExternalMenuPrice `json:"prices"`
}

type iikoExternalMenuPrice struct {
	Price *float64 `json:"price"`
}

func (p *IikoProvider) getExternalMenu(ctx context.Context) (*POSMenu, error) {
	var out iikoExternalMenuResponse
	err := p.doRequest(ctx, "/api/2/menu/by_id", map[string]any{
		"externalMenuId":  p.externalMenuID,
		"organizationIds": []string{p.orgID},
	}, &out)
	if err != nil {
		return nil, err
	}

	menu := &POSMenu{}
	for _, cat := range out.ItemCategories {
		if cat.Deleted || cat.IsHidden || cat.Name == "" {
			continue
		}
		mc := MenuCategory{Name: cat.Name}
		for _, item := range cat.Items {
			if item.Deleted || item.IsHidden || item.Name == "" {
				continue
			}
			externalID := item.IikoItemID
			if externalID == "" {
				externalID = item.ItemID
			}
			if externalID == "" {
				externalID = item.ID
			}
			mc.Items = append(mc.Items, MenuItem{
				ExternalID:  externalID,
				Name:        item.Name,
				Price:       externalMenuPrice(item.ItemSizes),
				Description: item.Description,
				ImageURL:    firstImageLink(item.ImageLinks),
			})
		}
		if len(mc.Items) > 0 {
			menu.Categories = append(menu.Categories, mc)
		}
	}
	return menu, nil
}

func firstImageLink(links []string) string {
	if len(links) == 0 {
		return ""
	}
	return links[0]
}

func externalMenuPrice(sizes []iikoExternalMenuItemSize) float64 {
	for _, size := range sizes {
		if size.Price != nil {
			return *size.Price
		}
		for _, price := range size.Prices {
			if price.Price != nil {
				return *price.Price
			}
		}
	}
	return 0
}

type iikoNomenclatureResponse struct {
	Groups   []iikoNomenclatureGroup   `json:"groups"`
	Products []iikoNomenclatureProduct `json:"products"`
}

type iikoNomenclatureGroup struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsDeleted bool   `json:"isDeleted"`
}

type iikoNomenclatureProduct struct {
	ID               string                  `json:"id"`
	Name             string                  `json:"name"`
	Description      string                  `json:"description"`
	ParentGroup      string                  `json:"parentGroup"`
	IsDeleted        bool                    `json:"isDeleted"`
	Type             string                  `json:"type"`
	Price            *float64                `json:"price"`
	DefaultSalePrice *float64                `json:"defaultSalePrice"`
	SizePrices       []iikoNomenclaturePrice `json:"sizePrices"`
}

type iikoNomenclaturePrice struct {
	Price        *float64 `json:"price"`
	CurrentPrice *float64 `json:"currentPrice"`
}

func (p *IikoProvider) getNomenclatureMenu(ctx context.Context) (*POSMenu, error) {
	var out iikoNomenclatureResponse
	if err := p.doRequest(ctx, "/nomenclature", map[string]string{"organizationId": p.orgID}, &out); err != nil {
		return nil, err
	}

	groups := make(map[string]string, len(out.Groups))
	for _, group := range out.Groups {
		if group.IsDeleted || group.ID == "" || group.Name == "" {
			continue
		}
		groups[group.ID] = group.Name
	}

	byCategory := make(map[string][]MenuItem)
	for _, product := range out.Products {
		if product.IsDeleted || product.ID == "" || product.Name == "" {
			continue
		}
		category := groups[product.ParentGroup]
		if category == "" {
			category = "Uncategorized"
		}
		byCategory[category] = append(byCategory[category], MenuItem{
			ExternalID:  product.ID,
			Name:        product.Name,
			Price:       nomenclaturePrice(product),
			Description: product.Description,
		})
	}

	names := make([]string, 0, len(byCategory))
	for name := range byCategory {
		names = append(names, name)
	}
	sort.Strings(names)

	menu := &POSMenu{}
	for _, name := range names {
		items := byCategory[name]
		sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
		groupID := ""
		for id, groupName := range groups {
			if groupName == name {
				groupID = id
				break
			}
		}
		menu.Categories = append(menu.Categories, MenuCategory{ExternalID: groupID, Name: name, Items: items})
	}
	return menu, nil
}

func nomenclaturePrice(product iikoNomenclatureProduct) float64 {
	if product.Price != nil {
		return *product.Price
	}
	if product.DefaultSalePrice != nil {
		return *product.DefaultSalePrice
	}
	for _, price := range product.SizePrices {
		if price.Price != nil {
			return *price.Price
		}
		if price.CurrentPrice != nil {
			return *price.CurrentPrice
		}
	}
	return 0
}

type iikoDeliveriesResponse struct {
	OrdersByOrganizations []iikoOrdersByOrganization `json:"ordersByOrganizations"`
}

type iikoOrdersByOrganization struct {
	OrganizationID string              `json:"organizationId"`
	Orders         []iikoDeliveryOrder `json:"orders"`
}

type iikoDeliveryOrder struct {
	ID    string            `json:"id"`
	PosID string            `json:"posId"`
	Order *iikoDeliveryBody `json:"order"`
}

type iikoDeliveryBody struct {
	Phone       string                `json:"phone"`
	Status      string                `json:"status"`
	Sum         float64               `json:"sum"`
	WhenCreated string                `json:"whenCreated"`
	WhenClosed  string                `json:"whenClosed"`
	Items       []iikoDeliveryItem    `json:"items"`
	Discounts   []iikoDiscountItem    `json:"discounts"`
	Customer    *iikoDeliveryCustomer `json:"customer"`
}

type iikoDeliveryCustomer struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
}

type iikoDiscountItem struct {
	Sum float64 `json:"sum"`
}

type iikoDeliveryItem struct {
	Type               string                 `json:"type"`
	Deleted            any                    `json:"deleted"`
	Amount             float64                `json:"amount"`
	Product            *iikoDeliveryProduct   `json:"product"`
	Price              float64                `json:"price"`
	Cost               float64                `json:"cost"`
	ResultSum          float64                `json:"resultSum"`
	PrimaryComponent   *iikoDeliveryComponent `json:"primaryComponent"`
	SecondaryComponent *iikoDeliveryComponent `json:"secondaryComponent"`
}

type iikoDeliveryProduct struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type iikoDeliveryComponent struct {
	Product   *iikoDeliveryProduct `json:"product"`
	Price     float64              `json:"price"`
	Cost      float64              `json:"cost"`
	ResultSum float64              `json:"resultSum"`
}

func mapIikoDeliveryOrder(in iikoDeliveryOrder) (POSOrder, bool) {
	if in.ID == "" || in.Order == nil {
		return POSOrder{}, false
	}

	orderedAt := parseIikoTime(in.Order.WhenClosed)
	if orderedAt.IsZero() {
		orderedAt = parseIikoTime(in.Order.WhenCreated)
	}
	if orderedAt.IsZero() {
		orderedAt = time.Now()
	}

	items := make([]POSOrderItem, 0, len(in.Order.Items))
	for _, item := range in.Order.Items {
		items = append(items, mapIikoDeliveryItems(item)...)
	}

	return POSOrder{
		ExternalID:    in.ID,
		CustomerPhone: in.Order.Phone,
		Items:         items,
		Total:         in.Order.Sum,
		Discount:      iikoDiscountTotal(in.Order.Discounts),
		OrderedAt:     orderedAt,
		Status:        strings.ToLower(in.Order.Status),
	}, true
}

func mapIikoDeliveryItems(item iikoDeliveryItem) []POSOrderItem {
	if item.Deleted != nil {
		return nil
	}
	switch item.Type {
	case "Product", "Service":
		if item.Product == nil || item.Product.Name == "" {
			return nil
		}
		return []POSOrderItem{{
			ExternalID: item.Product.ID,
			Name:       item.Product.Name,
			Quantity:   positiveInt(item.Amount),
			Price:      iikoItemPrice(item.Price, item.Cost, item.ResultSum, item.Amount),
		}}
	case "Compound":
		items := make([]POSOrderItem, 0, 2)
		if mapped, ok := mapIikoDeliveryComponent(item.Amount, item.PrimaryComponent); ok {
			items = append(items, mapped)
		}
		if mapped, ok := mapIikoDeliveryComponent(item.Amount, item.SecondaryComponent); ok {
			items = append(items, mapped)
		}
		return items
	default:
		return nil
	}
}

func mapIikoDeliveryComponent(amount float64, component *iikoDeliveryComponent) (POSOrderItem, bool) {
	if component == nil || component.Product == nil || component.Product.Name == "" {
		return POSOrderItem{}, false
	}
	return POSOrderItem{
		ExternalID: component.Product.ID,
		Name:       component.Product.Name,
		Quantity:   positiveInt(amount),
		Price:      iikoItemPrice(component.Price, component.Cost, component.ResultSum, amount),
	}, true
}

func iikoItemPrice(price, cost, resultSum, amount float64) float64 {
	if price != 0 {
		return price
	}
	if amount > 0 {
		if resultSum != 0 {
			return resultSum / amount
		}
		if cost != 0 {
			return cost / amount
		}
	}
	return 0
}

func positiveInt(v float64) int {
	if v < 1 {
		return 1
	}
	return int(v)
}

func iikoDiscountTotal(discounts []iikoDiscountItem) float64 {
	var total float64
	for _, discount := range discounts {
		total += discount.Sum
	}
	return total
}

func iikoDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.000")
}

func parseIikoTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	layouts := []string{
		"2006-01-02 15:04:05.000",
		"2006-01-02 15:04:05",
		time.RFC3339Nano,
		time.RFC3339,
	}
	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t
		}
	}
	return time.Time{}
}

// GetDailyAggregates computes per-day revenue, average check, transaction count,
// guest count and customer phones from closed/delivered orders. iiko Cloud OLAP
// reports are not available via the Transport API for most tariffs, so aggregates
// are derived from the same /deliveries data used by GetOrders.
func (p *IikoProvider) GetDailyAggregates(ctx context.Context, from, to time.Time) ([]POSDailyAggregate, error) {
	deliveries, err := p.fetchDeliveriesChunked(ctx, from, to)
	if err != nil {
		return nil, err
	}

	type dayAgg struct {
		revenue float64
		txCount int
		phones  map[string]struct{}
	}
	byDay := make(map[string]*dayAgg)
	dayOrder := make([]string, 0)

	for _, orderInfo := range deliveries {
		order, ok := mapIikoDeliveryOrder(orderInfo)
		if !ok {
			continue
		}
		key := order.OrderedAt.Format("2006-01-02")
		agg := byDay[key]
		if agg == nil {
			agg = &dayAgg{phones: make(map[string]struct{})}
			byDay[key] = agg
			dayOrder = append(dayOrder, key)
		}
		agg.revenue += order.Total
		agg.txCount++
		if order.CustomerPhone != "" {
			agg.phones[order.CustomerPhone] = struct{}{}
		}
	}

	sort.Strings(dayOrder)
	result := make([]POSDailyAggregate, 0, len(dayOrder))
	for _, key := range dayOrder {
		agg := byDay[key]
		date, _ := time.Parse("2006-01-02", key)
		var avgCheck float64
		if agg.txCount > 0 {
			avgCheck = agg.revenue / float64(agg.txCount)
		}
		phones := make([]string, 0, len(agg.phones))
		for phone := range agg.phones {
			phones = append(phones, phone)
		}
		sort.Strings(phones)
		result = append(result, POSDailyAggregate{
			Date:       date,
			Revenue:    agg.revenue,
			AvgCheck:   avgCheck,
			TxCount:    agg.txCount,
			GuestCount: len(phones),
			Phones:     phones,
		})
	}
	return result, nil
}
