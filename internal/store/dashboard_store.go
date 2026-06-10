package store

import (
	"context"
	"fmt"

	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DashboardStore interface {
	GetStats(ctx context.Context, ownerID, shopID int, dateFrom, dateTo string) (*models.DashboardStats, error)
	GetNotifications(ctx context.Context, ownerID, shopID int) (*models.NotificationSummary, error)
	GetOnboardingStatus(ctx context.Context, ownerID int) (*models.OnboardingStatus, error)
}

type dashboardStore struct {
	db *pgxpool.Pool
}

func NewDashboardStore(db *pgxpool.Pool) DashboardStore {
	return &dashboardStore{db: db}
}

func buildDateCond(col, dateFrom, dateTo string) string {
	cond := ""
	if dateFrom != "" {
		cond += fmt.Sprintf(" AND DATE(%s) >= '%s'", col, dateFrom)
	}
	if dateTo != "" {
		cond += fmt.Sprintf(" AND DATE(%s) <= '%s'", col, dateTo)
	}
	return cond
}

func (s *dashboardStore) GetStats(ctx context.Context, ownerID, shopID int, dateFrom, dateTo string) (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	orderShop := ""
	stockShop := ""
	if shopID > 0 {
		orderShop = fmt.Sprintf(" AND shop_id = %d", shopID)
		stockShop = fmt.Sprintf(" AND st.shop_id = %d", shopID)
	}
	dateCond := buildDateCond("created_at", dateFrom, dateTo)
	orderCond := orderShop + dateCond

	// Revenue total / today / this month
	_ = s.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT
			COALESCE(SUM(total_amount) FILTER (WHERE status = 'DELIVERED'), 0),
			COALESCE(SUM(total_amount) FILTER (WHERE status = 'DELIVERED' AND DATE(created_at) = CURRENT_DATE), 0),
			COALESCE(SUM(total_amount) FILTER (WHERE status = 'DELIVERED' AND DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW())), 0)
		FROM orders WHERE user_id = $1%s`, orderCond), ownerID,
	).Scan(&stats.RevenueTotal, &stats.RevenueToday, &stats.RevenueThisMonth)

	// Commandes par statut
	rows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT status, COUNT(*) FROM orders WHERE user_id = $1%s GROUP BY status`, orderCond), ownerID)
	if err == nil {
		defer rows.Close()
		stats.OrdersByStatus = map[string]int{}
		for rows.Next() {
			var status string
			var count int
			_ = rows.Scan(&status, &count)
			stats.OrdersByStatus[status] = count
		}
	}

	// Total commandes
	_ = s.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*) FROM orders WHERE user_id = $1%s`, orderCond), ownerID,
	).Scan(&stats.TotalOrders)

	// Top produits vendus (on remonte 10 pour permettre le tri côté client)
	topRows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT p.product_name, SUM(oi.quantity) AS total_qty, SUM(oi.quantity * oi.unit_price) AS revenue
		FROM order_items oi
		INNER JOIN orders o ON oi.order_id = o.order_id
		INNER JOIN products p ON oi.product_id = p.product_id
		WHERE o.user_id = $1 AND o.status = 'DELIVERED'%s
		GROUP BY p.product_name
		ORDER BY total_qty DESC
		LIMIT 10`, orderCond), ownerID)
	if err == nil {
		defer topRows.Close()
		for topRows.Next() {
			tp := models.TopProduct{}
			_ = topRows.Scan(&tp.ProductName, &tp.TotalQty, &tp.Revenue)
			stats.TopProducts = append(stats.TopProducts, tp)
		}
	}

	// Valeur totale du stock (pas filtrable par date — c'est l'état actuel)
	_ = s.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COALESCE(SUM(st.quantity * p.unit_price), 0)
		FROM stocks st
		INNER JOIN products p ON st.product_id = p.product_id
		WHERE p.user_id = $1 AND p.deleted_at IS NULL%s`, stockShop), ownerID,
	).Scan(&stats.StockValue)

	// Produits en stock faible
	_ = s.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*) FROM stocks st
		INNER JOIN products p ON st.product_id = p.product_id
		WHERE p.user_id = $1 AND p.deleted_at IS NULL AND st.quantity <= st.min_stock%s`, stockShop), ownerID,
	).Scan(&stats.LowStockCount)

	// Nombre de boutiques
	if shopID > 0 {
		stats.ShopCount = 1
	} else {
		_ = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM shops WHERE user_id = $1 AND deleted_at IS NULL`, ownerID).Scan(&stats.ShopCount)
	}

	// Revenus par boutique
	shopDateCond := buildDateCond("o.created_at", dateFrom, dateTo)
	shopRowQ := fmt.Sprintf(`
		SELECT sh.shop_name, COALESCE(SUM(o.total_amount) FILTER (WHERE o.status = 'DELIVERED'%s), 0) AS revenue
		FROM shops sh
		LEFT JOIN orders o ON o.shop_id = sh.shop_id AND o.user_id = $1
		WHERE sh.user_id = $1 AND sh.deleted_at IS NULL`, shopDateCond)
	if shopID > 0 {
		shopRowQ += fmt.Sprintf(" AND sh.shop_id = %d", shopID)
	}
	shopRowQ += " GROUP BY sh.shop_name ORDER BY revenue DESC"

	shopRows, err := s.db.Query(ctx, shopRowQ, ownerID)
	if err == nil {
		defer shopRows.Close()
		for shopRows.Next() {
			sr := models.ShopRevenue{}
			_ = shopRows.Scan(&sr.ShopName, &sr.Revenue)
			stats.RevenueByShop = append(stats.RevenueByShop, sr)
		}
	}

	// Évolution journalière — plage personnalisée ou 30 jours par défaut
	evoDateCond := ""
	if dateFrom != "" || dateTo != "" {
		evoDateCond = dateCond
	} else {
		evoDateCond = " AND created_at >= NOW() - INTERVAL '30 days'"
	}
	evoRows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT TO_CHAR(DATE(created_at), 'YYYY-MM-DD') AS day,
		       COALESCE(SUM(total_amount), 0) AS revenue,
		       COUNT(*) AS orders
		FROM orders
		WHERE user_id = $1 AND status = 'DELIVERED'%s%s
		GROUP BY DATE(created_at)
		ORDER BY DATE(created_at) ASC`, orderShop, evoDateCond), ownerID)
	if err == nil {
		defer evoRows.Close()
		for evoRows.Next() {
			dr := models.DailyRevenue{}
			_ = evoRows.Scan(&dr.Day, &dr.Revenue, &dr.Orders)
			stats.RevenueEvolution = append(stats.RevenueEvolution, dr)
		}
	}

	// Répartition par paiement
	pmRows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT COALESCE(payment_method::text, 'NON DÉFINI') AS method,
		       COUNT(*) AS cnt,
		       COALESCE(SUM(total_amount), 0) AS revenue
		FROM orders
		WHERE user_id = $1 AND status = 'DELIVERED'%s
		  AND payment_method IS NOT NULL
		GROUP BY payment_method
		ORDER BY revenue DESC`, orderCond), ownerID)
	if err == nil {
		defer pmRows.Close()
		for pmRows.Next() {
			ps := models.PaymentShare{}
			_ = pmRows.Scan(&ps.Method, &ps.Count, &ps.Revenue)
			stats.PaymentBreakdown = append(stats.PaymentBreakdown, ps)
		}
	}

	return stats, nil
}

func (s *dashboardStore) GetNotifications(ctx context.Context, ownerID, shopID int) (*models.NotificationSummary, error) {
	summary := &models.NotificationSummary{
		LowStock:      []models.LowStockAlert{},
		PendingOrders: []models.PendingOrderAlert{},
	}

	shopCond := ""
	if shopID > 0 {
		shopCond = fmt.Sprintf(" AND sh.shop_id = %d", shopID)
	}

	// Stock faible
	lsRows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT p.product_name, sh.shop_name, sh.shop_id, st.quantity, st.min_stock
		FROM stocks st
		INNER JOIN products p  ON st.product_id = p.product_id
		INNER JOIN shops    sh ON st.shop_id    = sh.shop_id
		WHERE p.user_id = $1 AND p.deleted_at IS NULL AND sh.deleted_at IS NULL
		  AND st.quantity <= st.min_stock%s
		ORDER BY st.quantity ASC
		LIMIT 20`, shopCond), ownerID)
	if err == nil {
		defer lsRows.Close()
		for lsRows.Next() {
			a := models.LowStockAlert{}
			_ = lsRows.Scan(&a.ProductName, &a.ShopName, &a.ShopID, &a.Quantity, &a.MinStock)
			summary.LowStock = append(summary.LowStock, a)
		}
	}

	// Commandes en attente depuis plus de 24h
	orderShopCond := ""
	if shopID > 0 {
		orderShopCond = fmt.Sprintf(" AND o.shop_id = %d", shopID)
	}
	poRows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT o.uuid::text, COALESCE(sh.shop_name,'') AS shop_name,
		       FLOOR(EXTRACT(EPOCH FROM (NOW() - o.created_at)) / 3600)::int AS hours_pending,
		       o.total_amount
		FROM orders o
		LEFT JOIN shops sh ON sh.shop_id = o.shop_id
		WHERE o.user_id = $1 AND o.status = 'PENDING'
		  AND o.created_at < NOW() - INTERVAL '24 hours'%s
		ORDER BY o.created_at ASC
		LIMIT 20`, orderShopCond), ownerID)
	if err == nil {
		defer poRows.Close()
		for poRows.Next() {
			a := models.PendingOrderAlert{}
			_ = poRows.Scan(&a.UUID, &a.ShopName, &a.HoursPending, &a.TotalAmount)
			summary.PendingOrders = append(summary.PendingOrders, a)
		}
	}

	summary.TotalCount = len(summary.LowStock) + len(summary.PendingOrders)
	return summary, nil
}

func (s *dashboardStore) GetOnboardingStatus(ctx context.Context, ownerID int) (*models.OnboardingStatus, error) {
	st := &models.OnboardingStatus{}

	var shopCount, productCount, vendeurCount int
	_ = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM shops    WHERE user_id  = $1 AND deleted_at IS NULL`, ownerID).Scan(&shopCount)
	_ = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM products WHERE user_id  = $1 AND deleted_at IS NULL`, ownerID).Scan(&productCount)
	_ = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM users    WHERE parent_id = $1 AND deleted_at IS NULL`, ownerID).Scan(&vendeurCount)

	st.HasShop    = shopCount    > 0
	st.HasProduct = productCount > 0
	st.HasVendeur = vendeurCount > 0
	st.Complete   = st.HasShop && st.HasProduct && st.HasVendeur
	return st, nil
}
