package model

// 邮件发送规则
type EmailSendRules struct {
	Id             int64 `gorm:"column:id;type:int(11) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	DomainId       int64 `gorm:"column:domain_id;type:int(11) unsigned;default:0;NOT NULL" json:"domain_id"`
	StartedAt      int64 `gorm:"column:started_at;type:int(11) unsigned;default:0;NOT NULL" json:"started_at"`
	EndedAt        int64 `gorm:"column:ended_at;type:int(11) unsigned;default:0;NOT NULL" json:"ended_at"`
	TotalSendCount uint  `gorm:"column:total_send_count;type:int(11) unsigned;default:0;comment:预计总发送量;NOT NULL" json:"total_send_count"`
	SentCount      uint  `gorm:"column:sent_count;type:int(11) unsigned;default:0;comment:已发数量;NOT NULL" json:"sent_count"`
	Status         uint  `gorm:"column:status;type:tinyint(4) unsigned;default:0;comment:启用状态; 0: 禁用, 1: 启用;NOT NULL" json:"status"`
	CreatedAt      uint  `gorm:"column:created_at;type:int(11) unsigned;default:0;NOT NULL" json:"created_at"`
	UpdatedAt      uint  `gorm:"column:updated_at;type:int(11) unsigned;default:0;NOT NULL" json:"updated_at"`
}
