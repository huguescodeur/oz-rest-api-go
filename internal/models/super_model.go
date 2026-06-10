package models

import "time"

type ResetToken struct {
	ID        int
	UserID    int
	Token     string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

type GlobalStats struct {
	TotalAdmins   int     `json:"total_admins"`
	TotalShops    int     `json:"total_shops"`
	TotalVendeurs int     `json:"total_vendeurs"`
	TotalOrders   int     `json:"total_orders"`
	TotalRevenue  int64   `json:"total_revenue"`
	NewAdmins30d  int     `json:"new_admins_30d"`
	NewOrders30d  int     `json:"new_orders_30d"`
	Revenue30d    int64   `json:"revenue_30d"`
	TopAdmins     []AdminSummary `json:"top_admins"`
}

type AdminSummary struct {
	ID           int       `json:"id"`
	UUID         string    `json:"uuid"`
	Username     string    `json:"username"`
	Firstname    string    `json:"firstname"`
	Lastname     string    `json:"lastname"`
	Email        string    `json:"email"`
	ShopCount    int       `json:"shop_count"`
	VendeurCount int       `json:"vendeur_count"`
	OrderCount   int       `json:"order_count"`
	Revenue      int64     `json:"revenue"`
	CreatedAt    time.Time `json:"created_at"`
}

type PlatformSetting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SuperAnalytics struct {
	RevenueEvolution  []DailyRevenue  `json:"revenue_evolution"`
	OrdersByStatus    map[string]int  `json:"orders_by_status"`
	RevenueByAdmin    []AdminRevenue  `json:"revenue_by_admin"`
	RegistrationTrend []DailyCount    `json:"registration_trend"`
	PaymentBreakdown  []PaymentShare  `json:"payment_breakdown"`
}

type AdminRevenue struct {
	AdminName string `json:"admin_name"`
	Revenue   int64  `json:"revenue"`
	Orders    int    `json:"orders"`
}

type DailyCount struct {
	Day   string `json:"day"`
	Count int    `json:"count"`
}

type GlobalActivity struct {
	MovementID     int       `json:"movement_id"`
	ProductName    string    `json:"product_name"`
	ShopName       string    `json:"shop_name"`
	AdminName      string    `json:"admin_name"`
	Username       string    `json:"username"`
	QuantityChange int       `json:"quantity_change"`
	MovementType   string    `json:"movement_type"`
	Comment        string    `json:"comment"`
	CreatedAt      time.Time `json:"created_at"`
}
