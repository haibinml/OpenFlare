package model

import "time"

type WAFRuleGroup struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	Name              string    `json:"name" gorm:"size:255;not null"`
	Enabled           bool      `json:"enabled" gorm:"not null;default:true"`
	IsGlobal          bool      `json:"is_global" gorm:"not null;default:false;index"`
	BlockStatusCode   int       `json:"block_status_code" gorm:"not null;default:418"`
	BlockResponseBody string    `json:"block_response_body" gorm:"type:text;not null;default:''"`
	IPWhitelist       string    `json:"ip_whitelist" gorm:"type:text;not null;default:'[]'"`
	IPBlacklist       string    `json:"ip_blacklist" gorm:"type:text;not null;default:'[]'"`
	CountryWhitelist  string    `json:"country_whitelist" gorm:"type:text;not null;default:'[]'"`
	CountryBlacklist  string    `json:"country_blacklist" gorm:"type:text;not null;default:'[]'"`
	RegionWhitelist   string    `json:"region_whitelist" gorm:"type:text;not null;default:'[]'"`
	RegionBlacklist   string    `json:"region_blacklist" gorm:"type:text;not null;default:'[]'"`
	Remark            string    `json:"remark" gorm:"size:255"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type WAFRuleGroupBinding struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	RuleGroupID  uint      `json:"rule_group_id" gorm:"not null;uniqueIndex:idx_waf_group_route"`
	ProxyRouteID uint      `json:"proxy_route_id" gorm:"not null;uniqueIndex:idx_waf_group_route;index"`
	CreatedAt    time.Time `json:"created_at"`
}

func ListWAFRuleGroups() ([]*WAFRuleGroup, error) {
	var groups []*WAFRuleGroup
	err := DB.Order("is_global desc").Order("id asc").Find(&groups).Error
	return groups, err
}

func GetWAFRuleGroupByID(id uint) (*WAFRuleGroup, error) {
	group := &WAFRuleGroup{}
	err := DB.First(group, id).Error
	return group, err
}

func GetGlobalWAFRuleGroup() (*WAFRuleGroup, error) {
	group := &WAFRuleGroup{}
	err := DB.Where("is_global = ?", true).Order("id asc").First(group).Error
	return group, err
}

func (group *WAFRuleGroup) Insert() error {
	return DB.Create(group).Error
}

func (group *WAFRuleGroup) Update() error {
	return DB.Model(&WAFRuleGroup{}).Where("id = ?", group.ID).Updates(map[string]any{
		"name":                group.Name,
		"enabled":             group.Enabled,
		"is_global":           group.IsGlobal,
		"block_status_code":   group.BlockStatusCode,
		"block_response_body": group.BlockResponseBody,
		"ip_whitelist":        group.IPWhitelist,
		"ip_blacklist":        group.IPBlacklist,
		"country_whitelist":   group.CountryWhitelist,
		"country_blacklist":   group.CountryBlacklist,
		"region_whitelist":    group.RegionWhitelist,
		"region_blacklist":    group.RegionBlacklist,
		"remark":              group.Remark,
	}).Error
}

func (group *WAFRuleGroup) Delete() error {
	return DB.Delete(group).Error
}
