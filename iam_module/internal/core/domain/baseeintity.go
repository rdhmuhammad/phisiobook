package domain

import (
	"fmt"
	"os"
	"time"
)

type BaseEntity struct {
	ID        uint      `gorm:"primaryKey;type:BIGINT UNSIGNED AUTO_INCREMENT" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime:false" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime:false" json:"updatedAt"`
	CreatedBy string    `gorm:"column:created_by;type:varchar(100)" json:"createdBy"`
	UpdatedBy string    `gorm:"column:updated_by;type:varchar(100)" json:"updatedBy"`
}

func (receiver *BaseEntity) SetCreated(actor string) {
	if actor != "" {
		receiver.CreatedBy = actor
		receiver.UpdatedBy = actor
	}
	receiver.CreatedAt = time.Now().UTC()
	receiver.UpdatedAt = time.Now().UTC()
}

func (receiver *BaseEntity) GetCreatedAt() time.Time {
	return receiver.CreatedAt
}

func (receiver *BaseEntity) GetUpdatedAt() time.Time {
	return receiver.UpdatedAt
}

func (receiver *BaseEntity) SetUpdated(actor string) {
	if actor != "" {
		receiver.UpdatedBy = actor
	}
	receiver.UpdatedAt = time.Now().UTC()
}

func (receiver *BaseEntity) SetPrimaryKey(id uint) {
	receiver.ID = id
}

// ===================== GET IMAGE ===========================

func (receiver *BaseEntity) downloadFile(value string) string {
	if value == "" {
		return ""
	}

	return fmt.Sprintf("%s/api/v1/download?fileName=%s",
		os.Getenv("BACKEND_URL"),
		value,
	)
}
