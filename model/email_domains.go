package model

// 邮件域名
type EmailDomains struct {
	Id              int64  `gorm:"column:id;type:int(11) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	ServiceProvider int64  `gorm:"column:service_provider;type:tinyint(4) unsigned;comment:服务商; 1: Mailgun, 2: Sendgrid;NOT NULL" json:"service_provider"`
	Domain          string `gorm:"column:domain;type:varchar(64);comment:发件域;NOT NULL" json:"domain"`
	ApiKeyId        string `gorm:"column:api_key_id;type:varchar(64);comment:API KEY-ID;NOT NULL" json:"api_key_id"`
	ApiKey          string `gorm:"column:api_key;type:varchar(128);comment:API 密钥;NOT NULL" json:"api_key"`
	Priority        uint   `gorm:"column:priority;type:tinyint(4) unsigned;default:0;comment:优先级, 越小越优先;NOT NULL" json:"priority"`
	Weight          uint   `gorm:"column:weight;type:tinyint(4) unsigned;default:1;comment:轮训算法权重" json:"weight"`
	Status          uint   `gorm:"column:status;type:tinyint(4) unsigned;default:0;comment:状态; 0: 禁用, 1: 启用;NOT NULL" json:"status"`
	CreatedAt       uint   `gorm:"column:created_at;type:int(11) unsigned;default:0;NOT NULL" json:"created_at"`
	UpdatedAt       uint   `gorm:"column:updated_at;type:int(11) unsigned;default:0;NOT NULL" json:"updated_at"`

	Rules []*EmailSendRules `json:"rules" gorm:"foreignKey:domain_id;references:id;"`
}
