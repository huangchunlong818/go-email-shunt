package sender

import (
	"errors"
	"strings"

	"github.com/huangchunlong818/go-email-shunt/model"
	"github.com/huangchunlong818/go-email-shunt/utils"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SenderManager管理电子邮件发送域的选择
type SenderManager struct {
	rds                  *redis.Client
	db                   *gorm.DB
	domains              []*model.EmailDomains
	emailTypeWithDomains map[int][]*model.EmailDomains
	ruleModel            int             //域名选择算法模型 1-redis限流 2-平滑加权轮询
	emailSendInfoList    []EmailSendInfo //用于电子邮件发送的配置规则, 邮件类型, 对应域ID, from邮箱
	nonShuntingDomainIds []int           //不进行域名预热管理的域ID
	shuntingDomainIds    []int           //进行域名预热管理的域ID, gmail才会进
}

func NewSenderManager(redisCli *redis.Client, db *gorm.DB, ruleModel int, nonShuntingDomainIds []int, shuntingDomainIds []int, emailSendInfoList []EmailSendInfo) SenderManager {
	sm := SenderManager{
		rds:                  redisCli,
		db:                   db,
		ruleModel:            ruleModel,
		domains:              make([]*model.EmailDomains, 0),
		emailTypeWithDomains: make(map[int][]*model.EmailDomains, 0),
	}
	if len(nonShuntingDomainIds) > 0 {
		sm.nonShuntingDomainIds = nonShuntingDomainIds
	} else {
		sm.nonShuntingDomainIds = []int{13}
	}

	if len(shuntingDomainIds) > 0 {
		sm.shuntingDomainIds = shuntingDomainIds
	} else {
		sm.shuntingDomainIds = []int{1, 2, 13}
	}

	if len(sm.emailSendInfoList) > 0 {
		sm.emailSendInfoList = emailSendInfoList
	} else {
		sm.emailSendInfoList = []EmailSendInfo{
			{
				EmailTypes: []int{1, 2},
				DomainIds:  []int{1, 9, 13},
				EmailInfo: EmailInfo{
					FromEmail: "no-reply@parcelpanel.net",
				},
			},
			{
				EmailTypes: []int{1, 2},
				DomainIds:  []int{2, 3, 4, 8},
				EmailInfo: EmailInfo{
					FromEmail: "no-reply@parcelpanel.com",
				},
			},
		}
	}
	return sm
}

// Send 发送一封邮件
func (sm *SenderManager) Send(message *MailContent) (*EmailSendResult, error) {
	domain, rule := sm.SelectSendingDomain(message)
	if domain == nil {
		return nil, errors.New("no sending domain selected")
	}
	zap.S().Infof("GET domain:%+v\n", *domain)
	zap.S().Infof("GET rule:%+v\n", *rule)
	//获取邮件配置
	emailInfo := sm.GetDomainEmailInfo(domain, message.EmailType)
	if emailInfo != nil {
		// 设置当前发送域的发件邮箱
		if emailInfo.FromEmail != "" {
			message.From.Email = emailInfo.FromEmail
		}
	}
	sender := sm.GetSenderByDomain(domain)
	return nil, nil
	// 发送邮件
	result, err := sender.Send(message)
	if err != nil {
		return nil, err
	}
	if result.Success {
		err = sm.db.Model(&model.EmailSendRules{}).
			Where("domain_id = ?", rule.DomainId).
			Update("sent_count", gorm.Expr("sent_count + 1")).Error
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// GetSendingDomainIdsByToAddr 根据收件人的电子邮件地址选择域ID
func (sm *SenderManager) GetSendingDomainIdsByToAddr(toEmail string) []int {
	nonShuntingDomainIds := sm.nonShuntingDomainIds
	shuntingDomainIds := sm.shuntingDomainIds

	toEmailDomain := strings.Split(toEmail, "@")[1]

	if toEmailDomain == "gmail.com" {
		return shuntingDomainIds
	}

	return nonShuntingDomainIds
}

// GetEmailSendInfoList 返回用于电子邮件发送的配置规则
func (sm *SenderManager) GetEmailSendInfoList() []EmailSendInfo {
	return sm.emailSendInfoList
}

// GetEmailTypeWithDomain 返回电子邮件类型到其相应域的映射
func (sm *SenderManager) GetEmailTypeWithDomain() map[int][]*model.EmailDomains {
	if len(sm.emailTypeWithDomains) > 0 {
		return sm.emailTypeWithDomains
	}
	sendInfoConfig := sm.GetEmailSendInfoList()
	typeWithDomainIds := map[int][]int{}

	for _, v := range sendInfoConfig {
		for _, emailType := range v.EmailTypes {
			if _, ok := typeWithDomainIds[emailType]; !ok {
				typeWithDomainIds[emailType] = []int{}
			}
			typeWithDomainIds[emailType] = append(typeWithDomainIds[emailType], v.DomainIds...)
		}
	}
	// domains 是有排序的，为了保证顺序，只能统一收集 domainIds 再使用 only 进行取出
	for emailType, domainIds := range typeWithDomainIds {
		emaidomains, err := sm.getFilterDomains(domainIds)
		if err != nil {
			return nil
		}
		sm.emailTypeWithDomains[emailType] = emaidomains
	}
	return sm.emailTypeWithDomains
}

// FindDomainsByEmailType finds domains by email type.
func (sm *SenderManager) FindDomainsByEmailType(emailType int) []*model.EmailDomains {
	typeDomains := sm.GetEmailTypeWithDomain()
	if domains, ok := typeDomains[emailType]; ok {
		return domains
	}
	return nil
}

// SelectSendingDomain 按电子邮件类型选择域
func (sm *SenderManager) SelectSendingDomain(message *MailContent) (*model.EmailDomains, *model.EmailSendRules) {
	// 获取收件邮箱对应的发件域
	domainIds := sm.GetSendingDomainIdsByToAddr(message.To.Email)
	var domainIdsMap = make(map[int]struct{}, len(domainIds))
	for _, v := range domainIds {
		domainIdsMap[v] = struct{}{}
	}
	// 获取邮件类型对应的发件域
	domains := sm.FindDomainsByEmailType(message.EmailType)
	zap.S().Infof("domainIdsMap:%+v \n", domainIdsMap)
	var accessDomains = map[int64]*model.EmailDomains{}
	for _, domain := range domains {
		//zap.S().Infof("domain:%+v\n", domain)
		if _, ok := domainIdsMap[int(domain.Id)]; !ok {
			continue
		}
		accessDomains[domain.Id] = domain
		if sm.ruleModel == 1 { //域名选择算法模型 1-redis限流 2-平滑加权轮询
			// 取出发送域对应的规则列表
			for _, rule := range domain.Rules {
				// 将发送规则标记为使用中
				if !sm.markAsUse(rule) {
					// 标记失败，继续尝试下一个规则
					continue
				}
				return domain, rule
			}
		}
	}
	if sm.ruleModel == 2 { //域名选择算法模型 1-redis限流 2-平滑加权轮询
		emailDomain := sm.SmoothWeightedRoundRobin(accessDomains, message.EmailType)
		if len(emailDomain.Rules) > 0 {
			return emailDomain, emailDomain.Rules[0]
		}
	}

	return nil, nil
}

// GetDomainEmailInfo 获取域和电子邮件类型的电子邮件配置信息
func (sm *SenderManager) GetDomainEmailInfo(domain *model.EmailDomains, emailType int) *EmailInfo {
	sendInfoConfig := sm.GetEmailSendInfoList()
	for _, item := range sendInfoConfig {
		if utils.InSlice(int(domain.Id), item.DomainIds) && utils.InSlice(emailType, item.EmailTypes) {
			return &item.EmailInfo
		}
	}
	return nil
}

// GetSenderByDomain 按域获取邮件发送实例
func (sm *SenderManager) GetSenderByDomain(domain *model.EmailDomains) EmailSenderInterface {
	switch domain.ServiceProvider {
	case 1: //mailgun
		return NewMailgunSender(&EmailDomain{
			Domain: domain.Domain,
			APIKey: domain.ApiKey,
		})
	case 2:
		return NewSendgridSender(&EmailDomain{
			Domain: domain.Domain,
			APIKey: domain.ApiKey,
		})
	}
	return nil
}
