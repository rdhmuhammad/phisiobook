package repository

import (
	"context"
	"fmt"
	"github.com/rdhmuhammad/phisiobook/shared/payload"

	"gorm.io/gorm"
	"iam_module/internal/core/constant"
	"iam_module/internal/core/domain"
)

type userRepo struct {
	db *gorm.DB
}

type UserListQuery struct {
	Filter    *payload.GetListQueryNoPeriod `bindQuery:"dive=true" json:"filter"`
	RoleName  string                        `json:"roleName" example:"USER"`
	StatusKey string                        `json:"statusKey" example:"active"`
}

func NewUserRepo(db *gorm.DB) UserRepo {
	return userRepo{db: db}
}

type UserRepo interface {
	UserDashboardList(ctx context.Context, query UserListQuery) ([]domain.UserListItem, int, int, error)
}

func (repo userRepo) UserDashboardList(ctx context.Context, query UserListQuery) ([]domain.UserListItem, int, int, error) {
	db := repo.db.WithContext(ctx)

	buildConditions := func(baseQuery *gorm.DB, tableName string, statusCol string) *gorm.DB {
		var status = 0
		// Apply status filter
		if query.StatusKey != "" {
			if query.StatusKey == domain.Active {
				status = 1
			}
			baseQuery = baseQuery.Where(fmt.Sprintf("%s.%s = ?", tableName, statusCol), status)
		}
		// Apply search filter
		if query.Filter.Search != "" {
			searchPattern := "%" + query.Filter.Search + "%"
			baseQuery = baseQuery.Where(
				fmt.Sprintf("%s.full_name LIKE ? OR %s.email LIKE ?", tableName, tableName),
				searchPattern, searchPattern)
		}

		tableUser := domain.User{}.TableName()
		if query.RoleName != "" && tableUser != tableName {
			baseQuery = baseQuery.Where("master_roles.name = ?", query.RoleName)
		}

		return baseQuery
	}

	mobileSql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return buildConditions(
			tx.
				Model(&domain.User{}).
				Select("`users`.id, full_name as name, email, if(is_verified=1, 'active','inactive') as status, last_active, 'USER' as role_name"),
			domain.User{}.TableName(),
			"is_verified",
		).
			Where("is_verified = 1").
			Find(&[]domain.User{})
	})

	dashboardSql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return buildConditions(
			tx.
				Model(&domain.UserAdmin{}).
				Select("`user_admins`.id, full_name as name, email, if(is_active=1, 'active','inactive') as status, last_active, master_roles.name as role_name").
				Joins("left join master_roles on master_roles.id = user_admins.role_id"),
			domain.UserAdmin{}.TableName(),
			"is_active").
			Find(&[]domain.UserAdmin{})
	})

	var finalQuery string
	switch query.RoleName {
	case constant.RoleIsUser:
		finalQuery = mobileSql
		break
	case constant.RoleIsAdmin:
		finalQuery = dashboardSql
		break
	default:
		finalQuery = fmt.Sprintf("(%s) UNION ALL (%s)", mobileSql, dashboardSql)
	}

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM (%s) as union_count", finalQuery)
	err := db.Raw(countSQL).Scan(&total).Error
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to count results: %w", err)
	}

	offset := (query.Filter.Page - 1) * query.Filter.PerPage
	totalPages := int((total + int64(query.Filter.PerPage) - 1) / int64(query.Filter.PerPage))

	// Execute union query with pagination and sorting
	finalSQL := fmt.Sprintf(`
		SELECT * FROM (%s) as union_table 
		ORDER BY last_active DESC 
		LIMIT ? OFFSET ?
	`, finalQuery)

	var results []domain.UserListItem
	err = db.Raw(finalSQL, query.Filter.PerPage, offset).Scan(&results).Error
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to execute union query: %w", err)
	}

	return results, int(total), totalPages, nil
}
