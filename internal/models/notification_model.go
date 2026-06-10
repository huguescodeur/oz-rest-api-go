package models

type LowStockAlert struct {
	ProductName string `json:"product_name"`
	ShopName    string `json:"shop_name"`
	ShopID      int    `json:"shop_id"`
	Quantity    int    `json:"quantity"`
	MinStock    int    `json:"min_stock"`
}

type PendingOrderAlert struct {
	UUID         string `json:"uuid"`
	ShopName     string `json:"shop_name"`
	HoursPending int    `json:"hours_pending"`
	TotalAmount  int64  `json:"total_amount"`
}

type NotificationSummary struct {
	LowStock      []LowStockAlert     `json:"low_stock"`
	PendingOrders []PendingOrderAlert  `json:"pending_orders"`
	TotalCount    int                  `json:"total_count"`
}

type OnboardingStatus struct {
	HasShop    bool `json:"has_shop"`
	HasProduct bool `json:"has_product"`
	HasVendeur bool `json:"has_vendeur"`
	Complete   bool `json:"complete"`
}
