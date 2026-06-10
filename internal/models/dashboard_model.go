package models

type DashboardStats struct {
	RevenueTotal      int64          `json:"revenue_total"`
	RevenueToday      int64          `json:"revenue_today"`
	RevenueThisMonth  int64          `json:"revenue_this_month"`
	TotalOrders       int            `json:"total_orders"`
	OrdersByStatus    map[string]int `json:"orders_by_status"`
	TopProducts       []TopProduct   `json:"top_products"`
	StockValue        int64          `json:"stock_value"`
	LowStockCount     int            `json:"low_stock_count"`
	ShopCount         int            `json:"shop_count"`
	RevenueByShop     []ShopRevenue  `json:"revenue_by_shop"`
	RevenueEvolution  []DailyRevenue `json:"revenue_evolution"`
	PaymentBreakdown  []PaymentShare `json:"payment_breakdown"`
}

type TopProduct struct {
	ProductName string `json:"product_name"`
	TotalQty    int    `json:"total_qty"`
	Revenue     int64  `json:"revenue"`
}

type ShopRevenue struct {
	ShopName string `json:"shop_name"`
	Revenue  int64  `json:"revenue"`
}

type DailyRevenue struct {
	Day     string `json:"day"`
	Revenue int64  `json:"revenue"`
	Orders  int    `json:"orders"`
}

type PaymentShare struct {
	Method  string `json:"method"`
	Count   int    `json:"count"`
	Revenue int64  `json:"revenue"`
}
