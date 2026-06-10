package store

import (
	"context"
	"fmt"

	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SuperStore interface {
	GetGlobalStats(ctx context.Context) (*models.GlobalStats, error)
	GetAdmins(ctx context.Context, limit, offset int) ([]*models.AdminSummary, int, error)
	GetSettings(ctx context.Context) (map[string]string, error)
	SetSetting(ctx context.Context, key, value string) error
	GetAnalytics(ctx context.Context, dateFrom, dateTo string) (*models.SuperAnalytics, error)
	GetGlobalLogs(ctx context.Context, dateFrom, dateTo, movType string, limit, offset int) ([]*models.GlobalActivity, int, error)
}

type superStore struct{ db *pgxpool.Pool }

func NewSuperStore(db *pgxpool.Pool) SuperStore { return &superStore{db: db} }

func (s *superStore) GetGlobalStats(ctx context.Context) (*models.GlobalStats, error) {
	st := &models.GlobalStats{}

	_ = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = 'admin' AND deleted_at IS NULL`).Scan(&st.TotalAdmins)
	_ = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM shops WHERE deleted_at IS NULL`).Scan(&st.TotalShops)
	_ = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = 'vendeur' AND deleted_at IS NULL`).Scan(&st.TotalVendeurs)
	_ = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM orders`).Scan(&st.TotalOrders)
	_ = s.db.QueryRow(ctx, `SELECT COALESCE(SUM(total_amount),0) FROM orders WHERE status = 'DELIVERED'`).Scan(&st.TotalRevenue)
	_ = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = 'admin' AND created_at >= NOW() - INTERVAL '30 days'`).Scan(&st.NewAdmins30d)
	_ = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM orders WHERE created_at >= NOW() - INTERVAL '30 days'`).Scan(&st.NewOrders30d)
	_ = s.db.QueryRow(ctx, `SELECT COALESCE(SUM(total_amount),0) FROM orders WHERE status = 'DELIVERED' AND created_at >= NOW() - INTERVAL '30 days'`).Scan(&st.Revenue30d)

	rows, err := s.db.Query(ctx, `
		SELECT u.id, u.uuid::text, u.username, u.firstname, u.lastname, u.email,
		       COUNT(DISTINCT sh.shop_id)   AS shop_count,
		       COUNT(DISTINCT v.id)          AS vendeur_count,
		       COUNT(DISTINCT o.order_id)    AS order_count,
		       COALESCE(SUM(CASE WHEN o.status = 'DELIVERED' THEN o.total_amount ELSE 0 END), 0) AS revenue
		FROM users u
		LEFT JOIN shops    sh ON sh.user_id = u.id AND sh.deleted_at IS NULL
		LEFT JOIN users     v ON v.parent_id = u.id AND v.role = 'vendeur' AND v.deleted_at IS NULL
		LEFT JOIN orders    o ON o.user_id = u.id
		WHERE u.role = 'admin' AND u.deleted_at IS NULL
		GROUP BY u.id, u.uuid, u.username, u.firstname, u.lastname, u.email
		ORDER BY revenue DESC
		LIMIT 10`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			a := models.AdminSummary{}
			_ = rows.Scan(&a.ID, &a.UUID, &a.Username, &a.Firstname, &a.Lastname, &a.Email,
				&a.ShopCount, &a.VendeurCount, &a.OrderCount, &a.Revenue)
			st.TopAdmins = append(st.TopAdmins, a)
		}
	}
	if st.TopAdmins == nil {
		st.TopAdmins = []models.AdminSummary{}
	}
	return st, nil
}

func (s *superStore) GetAdmins(ctx context.Context, limit, offset int) ([]*models.AdminSummary, int, error) {
	var total int
	_ = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = 'admin' AND deleted_at IS NULL`).Scan(&total)

	rows, err := s.db.Query(ctx, `
		SELECT u.id, u.uuid::text, u.username, u.firstname, u.lastname, u.email,
		       u.created_at,
		       COUNT(DISTINCT sh.shop_id)  AS shop_count,
		       COUNT(DISTINCT v.id)         AS vendeur_count,
		       COUNT(DISTINCT o.order_id)   AS order_count,
		       COALESCE(SUM(CASE WHEN o.status = 'DELIVERED' THEN o.total_amount ELSE 0 END), 0) AS revenue
		FROM users u
		LEFT JOIN shops  sh ON sh.user_id = u.id AND sh.deleted_at IS NULL
		LEFT JOIN users   v ON v.parent_id = u.id AND v.role = 'vendeur' AND v.deleted_at IS NULL
		LEFT JOIN orders  o ON o.user_id = u.id
		WHERE u.role = 'admin' AND u.deleted_at IS NULL
		GROUP BY u.id, u.uuid, u.username, u.firstname, u.lastname, u.email, u.created_at
		ORDER BY u.created_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var admins []*models.AdminSummary
	for rows.Next() {
		a := &models.AdminSummary{}
		if err := rows.Scan(&a.ID, &a.UUID, &a.Username, &a.Firstname, &a.Lastname, &a.Email,
			&a.CreatedAt, &a.ShopCount, &a.VendeurCount, &a.OrderCount, &a.Revenue); err != nil {
			return nil, 0, err
		}
		admins = append(admins, a)
	}
	if admins == nil {
		admins = []*models.AdminSummary{}
	}
	return admins, total, rows.Err()
}

func (s *superStore) GetSettings(ctx context.Context) (map[string]string, error) {
	rows, err := s.db.Query(ctx, `SELECT key, value FROM platform_settings ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		m[k] = v
	}
	return m, rows.Err()
}

func (s *superStore) GetAnalytics(ctx context.Context, dateFrom, dateTo string) (*models.SuperAnalytics, error) {
	a := &models.SuperAnalytics{
		OrdersByStatus: map[string]int{},
	}

	dateCond := ""
	if dateFrom != "" {
		dateCond += fmt.Sprintf(" AND DATE(o.created_at) >= '%s'", dateFrom)
	}
	if dateTo != "" {
		dateCond += fmt.Sprintf(" AND DATE(o.created_at) <= '%s'", dateTo)
	}
	evoDateCond := dateCond
	if dateFrom == "" && dateTo == "" {
		evoDateCond = " AND o.created_at >= NOW() - INTERVAL '30 days'"
	}

	// Évolution journalière du CA (toute la plateforme)
	evoRows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT TO_CHAR(DATE(o.created_at), 'YYYY-MM-DD'),
		       COALESCE(SUM(o.total_amount), 0),
		       COUNT(*)
		FROM orders o
		WHERE o.status = 'DELIVERED'%s
		GROUP BY DATE(o.created_at)
		ORDER BY DATE(o.created_at) ASC`, evoDateCond))
	if err == nil {
		defer evoRows.Close()
		for evoRows.Next() {
			d := models.DailyRevenue{}
			_ = evoRows.Scan(&d.Day, &d.Revenue, &d.Orders)
			a.RevenueEvolution = append(a.RevenueEvolution, d)
		}
	}

	// Commandes par statut (global)
	statusRows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT o.status, COUNT(*) FROM orders o WHERE 1=1%s GROUP BY o.status`, dateCond))
	if err == nil {
		defer statusRows.Close()
		for statusRows.Next() {
			var st string
			var cnt int
			_ = statusRows.Scan(&st, &cnt)
			a.OrdersByStatus[st] = cnt
		}
	}

	// CA par admin (top 15)
	adminRows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT u.username,
		       COALESCE(SUM(CASE WHEN o.status = 'DELIVERED' THEN o.total_amount ELSE 0 END), 0) AS revenue,
		       COUNT(o.order_id) AS orders
		FROM users u
		LEFT JOIN orders o ON o.user_id = u.id%s
		WHERE u.role = 'admin' AND u.deleted_at IS NULL
		GROUP BY u.id, u.username
		ORDER BY revenue DESC
		LIMIT 15`, dateCond))
	if err == nil {
		defer adminRows.Close()
		for adminRows.Next() {
			ar := models.AdminRevenue{}
			_ = adminRows.Scan(&ar.AdminName, &ar.Revenue, &ar.Orders)
			a.RevenueByAdmin = append(a.RevenueByAdmin, ar)
		}
	}

	// Tendance inscriptions admins (30j ou période)
	regCond := ""
	if dateFrom != "" {
		regCond += fmt.Sprintf(" AND DATE(created_at) >= '%s'", dateFrom)
	}
	if dateTo != "" {
		regCond += fmt.Sprintf(" AND DATE(created_at) <= '%s'", dateTo)
	}
	if dateFrom == "" && dateTo == "" {
		regCond = " AND created_at >= NOW() - INTERVAL '30 days'"
	}
	regRows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT TO_CHAR(DATE(created_at), 'YYYY-MM-DD'), COUNT(*)
		FROM users
		WHERE role = 'admin' AND deleted_at IS NULL%s
		GROUP BY DATE(created_at)
		ORDER BY DATE(created_at) ASC`, regCond))
	if err == nil {
		defer regRows.Close()
		for regRows.Next() {
			dc := models.DailyCount{}
			_ = regRows.Scan(&dc.Day, &dc.Count)
			a.RegistrationTrend = append(a.RegistrationTrend, dc)
		}
	}

	// Répartition paiements (global)
	pmRows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT COALESCE(payment_method::text, 'NON DÉFINI'), COUNT(*), COALESCE(SUM(total_amount), 0)
		FROM orders o
		WHERE status = 'DELIVERED' AND payment_method IS NOT NULL%s
		GROUP BY payment_method
		ORDER BY SUM(total_amount) DESC`, dateCond))
	if err == nil {
		defer pmRows.Close()
		for pmRows.Next() {
			ps := models.PaymentShare{}
			_ = pmRows.Scan(&ps.Method, &ps.Count, &ps.Revenue)
			a.PaymentBreakdown = append(a.PaymentBreakdown, ps)
		}
	}

	return a, nil
}

func (s *superStore) GetGlobalLogs(ctx context.Context, dateFrom, dateTo, movType string, limit, offset int) ([]*models.GlobalActivity, int, error) {
	cond := ""
	if dateFrom != "" {
		cond += fmt.Sprintf(" AND DATE(sm.created_at) >= '%s'", dateFrom)
	}
	if dateTo != "" {
		cond += fmt.Sprintf(" AND DATE(sm.created_at) <= '%s'", dateTo)
	}
	if movType != "" {
		cond += fmt.Sprintf(" AND sm.movement_type = '%s'", movType)
	}

	var total int
	_ = s.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*) FROM stock_movements sm WHERE 1=1%s`, cond)).Scan(&total)

	rows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT sm.movement_id,
		       p.product_name,
		       sh.shop_name,
		       admin_u.username AS admin_name,
		       COALESCE(actor.username, '') AS username,
		       sm.quantity_change, sm.movement_type, COALESCE(sm.comment,''), sm.created_at
		FROM stock_movements sm
		INNER JOIN products p   ON sm.product_id = p.product_id
		INNER JOIN shops    sh  ON sm.shop_id    = sh.shop_id
		INNER JOIN users    admin_u ON admin_u.id = p.user_id
		LEFT  JOIN users    actor   ON actor.id   = sm.user_id
		WHERE 1=1%s
		ORDER BY sm.created_at DESC
		LIMIT $1 OFFSET $2`, cond), limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*models.GlobalActivity
	for rows.Next() {
		g := &models.GlobalActivity{}
		if err := rows.Scan(
			&g.MovementID, &g.ProductName, &g.ShopName, &g.AdminName,
			&g.Username, &g.QuantityChange, &g.MovementType, &g.Comment, &g.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		logs = append(logs, g)
	}
	if logs == nil {
		logs = []*models.GlobalActivity{}
	}
	return logs, total, rows.Err()
}

func (s *superStore) SetSetting(ctx context.Context, key, value string) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO platform_settings (key, value, updated_at) VALUES ($1, $2, NOW())
		 ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()`,
		key, value)
	return err
}
