package service

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestWAFRuleGroupValidationAndNormalization(t *testing.T) {
	setupServiceTestDB(t)

	group, err := CreateWAFRuleGroup(WAFRuleGroupInput{
		Name:             "edge guard",
		Enabled:          true,
		BlockStatusCode:  451,
		IPWhitelist:      []string{" 192.0.2.1 ", "192.0.2.1", "198.51.100.0/24"},
		IPBlacklist:      []string{"203.0.113.10"},
		CountryBlacklist: []string{" cn ", "CN", "us"},
	})
	if err != nil {
		t.Fatalf("CreateWAFRuleGroup failed: %v", err)
	}
	if len(group.IPWhitelist) != 2 || group.IPWhitelist[0] != "192.0.2.1" || group.IPWhitelist[1] != "198.51.100.0/24" {
		t.Fatalf("unexpected normalized ip whitelist: %#v", group.IPWhitelist)
	}
	if len(group.CountryBlacklist) != 2 || group.CountryBlacklist[0] != "CN" || group.CountryBlacklist[1] != "US" {
		t.Fatalf("unexpected normalized countries: %#v", group.CountryBlacklist)
	}

	if _, err = CreateWAFRuleGroup(WAFRuleGroupInput{
		Name:        "bad ip",
		Enabled:     true,
		IPBlacklist: []string{"not-an-ip"},
	}); err == nil {
		t.Fatal("expected invalid IP to be rejected")
	}
}

func TestWAFGlobalGroupAndBindings(t *testing.T) {
	setupServiceTestDB(t)

	groups, err := ListWAFRuleGroups()
	if err != nil {
		t.Fatalf("ListWAFRuleGroups failed: %v", err)
	}
	if len(groups) == 0 || !groups[0].IsGlobal {
		t.Fatalf("expected default global WAF rule group, got %#v", groups)
	}
	if err = DeleteWAFRuleGroup(groups[0].ID); err == nil {
		t.Fatal("expected global WAF rule group delete to be rejected")
	}

	route, err := CreateProxyRoute(ProxyRouteInput{
		SiteName:  "waf-site",
		Domains:   []string{"waf.example.com"},
		OriginURL: "https://origin.internal",
		Enabled:   true,
	})
	if err != nil {
		t.Fatalf("CreateProxyRoute failed: %v", err)
	}
	custom, err := CreateWAFRuleGroup(WAFRuleGroupInput{
		Name:            "custom",
		Enabled:         true,
		BlockStatusCode: 418,
		IPBlacklist:     []string{"203.0.113.10"},
	})
	if err != nil {
		t.Fatalf("CreateWAFRuleGroup failed: %v", err)
	}
	if _, err = ReplaceWAFRuleGroupSites(custom.ID, []uint{route.ID}); err != nil {
		t.Fatalf("ReplaceWAFRuleGroupSites failed: %v", err)
	}
	siteGroups, err := GetWAFSiteRuleGroups(route.ID)
	if err != nil {
		t.Fatalf("GetWAFSiteRuleGroups failed: %v", err)
	}
	if len(siteGroups.AppliedIDs) != 1 || siteGroups.AppliedIDs[0] != custom.ID {
		t.Fatalf("unexpected site WAF bindings: %#v", siteGroups.AppliedIDs)
	}
}

func TestPublishConfigVersionIncludesWAFSnapshotAndRuntimeConfig(t *testing.T) {
	setupServiceTestDB(t)

	route, err := CreateProxyRoute(ProxyRouteInput{
		SiteName:  "waf-publish",
		Domains:   []string{"waf-publish.example.com"},
		OriginURL: "https://origin.internal",
		Enabled:   true,
	})
	if err != nil {
		t.Fatalf("CreateProxyRoute failed: %v", err)
	}
	group, err := CreateWAFRuleGroup(WAFRuleGroupInput{
		Name:            "publish group",
		Enabled:         true,
		BlockStatusCode: 451,
		IPBlacklist:     []string{"203.0.113.0/24"},
	})
	if err != nil {
		t.Fatalf("CreateWAFRuleGroup failed: %v", err)
	}
	if _, err = ReplaceWAFSiteRuleGroups(route.ID, []uint{group.ID}); err != nil {
		t.Fatalf("ReplaceWAFSiteRuleGroups failed: %v", err)
	}
	result, err := PublishConfigVersion("root", false)
	if err != nil {
		t.Fatalf("PublishConfigVersion failed: %v", err)
	}
	if !strings.Contains(result.Version.RenderedConfig, "access_by_lua_file __OPENFLARE_LUA_DIR__/waf/check.lua;") {
		t.Fatal("expected route config to include WAF lua access hook")
	}
	if !strings.Contains(result.Version.SnapshotJSON, `"waf"`) {
		t.Fatal("expected snapshot to include waf document")
	}
	var files []SupportFile
	if err = json.Unmarshal([]byte(result.Version.SupportFilesJSON), &files); err != nil {
		t.Fatalf("decode support files failed: %v", err)
	}
	found := false
	for _, file := range files {
		if file.Path == "waf_config.json" && strings.Contains(file.Content, "203.0.113.0/24") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected waf_config.json support file, got %#v", files)
	}
}
