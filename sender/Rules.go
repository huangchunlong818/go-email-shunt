package sender

import (
	"context"
	"emailWarming/model"
	"emailWarming/sender/ruleMode"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"time"
)

// 标记为使用中
func (sm *SenderManager) markAsUse(rule *model.EmailSendRules) bool {
	zap.S().Infof("rule:%+v\n", *rule)
	// 规则激活状态
	if rule.Status != 1 {
		return false
	}
	curTime := time.Now().Unix()
	//时间校验
	if rule.StartedAt > curTime || rule.EndedAt < curTime {
		return false
	}
	if rule.StartedAt == 0 {
		return true
	}
	var (
		access bool
		err    error
	)
	name := fmt.Sprintf("emailWarm:ruleLock:%d", rule.Id)
	//要求是个整数
	every := (rule.EndedAt - rule.StartedAt) / int64(rule.TotalSendCount)
	zap.S().Infof("every:%#v \n", every)
	access, err = sm.Throttle(sm.rds, name, 1, every)

	if err != nil {
		zap.S().Errorf("==markAsUse==err:%#v\n", err)
		return false
	}
	return access
}

// 简化版频率限制
func (sm *SenderManager) Throttle(client *redis.Client, key string, maxRequests int64, seconds int64) (bool, error) {
	//最大请求数
	// 间隔时间
	interval := time.Duration(seconds) * time.Second
	ctx := context.Background()

	// 获取当前计数
	count, err := client.Get(ctx, key).Int64()
	if err == redis.Nil {
		// 如果键不存在，初始化并设置过期时间
		err = client.Set(ctx, key, 1, interval).Err()
		if err != nil {
			return false, err
		}
		count = 1
		return true, nil
	} else if err != nil {
		return false, err
	} else {
		// 如果计数已达到最大值，检查是否可以重置
		if count >= maxRequests {
			return false, nil
		} else {
			// 否则，增加计数
			err = client.Incr(ctx, key).Err()
			if err != nil {
				return false, nil
			}
			count++
		}
	}

	zap.S().Infof("Current count: %d\n", count)
	return true, nil
}

// 平滑加权轮询
var swrr = make(map[int]*ruleMode.SmoothWeightedRoundRobin, 0)

// 平滑加权轮询算法
func (sm *SenderManager) SmoothWeightedRoundRobin(domainsMap map[int64]*model.EmailDomains, emailType int) *model.EmailDomains {
	if _, ok := swrr[emailType]; !ok {
		servers := make([]*ruleMode.Server, 0, len(domainsMap))
		for _, v := range domainsMap {
			item := &ruleMode.Server{
				Id:     v.Id,
				Name:   v.Domain,
				Weight: int(v.Weight),
			}
			servers = append(servers, item)
		}

		swrr[emailType] = ruleMode.NewSmoothWeightedRoundRobin(servers)
	}

	server := swrr[emailType].Next()
	if domain, ok := domainsMap[server.Id]; ok {
		return domain
	}
	return nil
}

// 获取过滤后可用邮件发送domain
func (sm *SenderManager) getFilterDomains(domainIds []int) ([]*model.EmailDomains, error) {
	var emailDomains []*model.EmailDomains
	if len(sm.domains) == 0 {
		_, err := sm.getActivatedDomains()
		if err != nil {
			return nil, err
		}
	}
	var domainIdsMap = make(map[int]struct{}, len(domainIds))
	for _, id := range domainIds {
		domainIdsMap[id] = struct{}{}
	}
	for _, item := range sm.domains {
		if _, ok := domainIdsMap[int(item.Id)]; ok {
			emailDomains = append(emailDomains, item)
		}
	}
	return emailDomains, nil
}

func (sm *SenderManager) getActivatedDomains() ([]*model.EmailDomains, error) {
	var emailDomains []*model.EmailDomains
	err := sm.db.Model(&model.EmailDomains{}).Preload("Rules").
		Where("status = 1").
		Order("priority ASC").
		Find(&emailDomains).Error
	if err != nil {
		return nil, err
	}
	sm.domains = emailDomains
	return emailDomains, nil
}
