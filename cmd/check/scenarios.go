package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

const (
	demoApiaryID = "apiary-north"
	demoHiveID   = "hive-01"
)

func seedApiary(c *client, id, name, district, keeper string) error {
	_, err := c.expect("POST", "/apiaries", map[string]any{
		"id":       id,
		"name":     name,
		"district": district,
		"keeper":   keeper,
	}, 201, `"id":"`+id+`"`)
	return err
}

func seedHive(c *client, apiaryID, hiveID, model string) error {
	_, err := c.expect("POST", "/apiaries/"+apiaryID+"/hives", map[string]any{
		"id":    hiveID,
		"model": model,
	}, 201, `"state":"colonized"`)
	return err
}

func seedDefault(c *client) error {
	if err := seedApiary(c, demoApiaryID, "北岭蜂场", "北岭片区", "周岚"); err != nil {
		return err
	}
	return seedHive(c, demoApiaryID, demoHiveID, "langstroth")
}

func checkRegisterApiary(c *client) error {
	if _, err := c.expect("POST", "/apiaries", map[string]any{
		"id": "apiary-north", "name": "北岭蜂场", "district": "北岭片区", "keeper": "周岚",
	}, 201, `"id":"apiary-north"`, `"name":"北岭蜂场"`); err != nil {
		return err
	}
	if _, err := c.expect("GET", "/apiaries/apiary-north", nil, 200,
		`"district":"北岭片区"`, `"keeper":"周岚"`); err != nil {
		return err
	}
	if _, err := c.expect("POST", "/apiaries", map[string]any{
		"id": "apiary-north", "name": "重复蜂场", "district": "北岭片区", "keeper": "周岚",
	}, 409, `"code":"conflict"`); err != nil {
		return err
	}
	if _, err := c.expect("POST", "/apiaries", map[string]any{
		"id": "apiary-south", "name": "", "district": "南坡片区", "keeper": "李澈",
	}, 400, `"code":"invalid_request"`); err != nil {
		return err
	}
	return nil
}

func checkHiveInstall(c *client) error {
	if err := seedApiary(c, demoApiaryID, "北岭蜂场", "北岭片区", "周岚"); err != nil {
		return err
	}
	if _, err := c.expect("POST", "/apiaries/apiary-north/hives", map[string]any{
		"id": "hive-01", "model": "langstroth",
	}, 201, `"state":"colonized"`, `"apiaryId":"apiary-north"`); err != nil {
		return err
	}
	if _, err := c.expect("GET", "/hives/hive-01", nil, 200, `"model":"langstroth"`); err != nil {
		return err
	}
	if _, err := c.expect("POST", "/apiaries/apiary-north/hives", map[string]any{
		"id": "hive-01", "model": "warre",
	}, 409, `"code":"conflict"`); err != nil {
		return err
	}
	if _, err := c.expect("POST", "/apiaries/apiary-missing/hives", map[string]any{
		"id": "hive-09", "model": "warre",
	}, 404, `"code":"not_found"`); err != nil {
		return err
	}
	if _, err := c.expect("POST", "/apiaries/apiary-north/hives", map[string]any{
		"id": "hive-02", "model": "dadant",
	}, 400, `"code":"invalid_request"`); err != nil {
		return err
	}
	return nil
}

func checkRecordInspection(c *client) error {
	if err := seedDefault(c); err != nil {
		return err
	}
	if _, err := c.expect("POST", "/hives/hive-01/inspections", map[string]any{
		"observedAt": "2026-04-10T09:00:00Z", "colonyStrength": 9, "broodPattern": "patchy",
		"temper": "calm", "miteCount": 3, "note": "",
	}, 400, `"code":"invalid_request"`); err != nil {
		return err
	}
	if _, err := c.expect("POST", "/hives/hive-01/inspections", map[string]any{
		"observedAt": "2026-04-10T09:00:00Z", "colonyStrength": 3, "broodPattern": "patchy",
		"temper": "aggressive", "miteCount": 12, "note": "西侧巢门观察",
	}, 201, `"inspection-000001"`, `"mite_load"`, `"temper_risk"`); err != nil {
		return err
	}
	if _, err := c.expect("GET", "/hives/hive-01/inspections", nil, 200, `"note":"西侧巢门观察"`); err != nil {
		return err
	}
	if _, err := c.expect("POST", "/hives/hive-01/state", map[string]any{"target": "retired"}, 200,
		`"to":"retired"`); err != nil {
		return err
	}
	if _, err := c.expect("POST", "/hives/hive-01/inspections", map[string]any{
		"observedAt": "2026-04-11T09:00:00Z", "colonyStrength": 3, "broodPattern": "solid",
		"temper": "calm", "miteCount": 1, "note": "",
	}, 409, `"code":"invalid_state"`); err != nil {
		return err
	}
	return nil
}

func checkReplaceQueen(c *client) error {
	if err := seedDefault(c); err != nil {
		return err
	}
	if _, err := c.expect("POST", "/hives/hive-01/queens", map[string]any{
		"origin": "bred", "introducedAt": "2026-04-02T08:00:00Z", "reason": "aging",
	}, 201, `"queen-000001"`, `"markColor":"white"`, `"state":"laying"`); err != nil {
		return err
	}
	if _, err := c.expect("POST", "/hives/hive-01/queens", map[string]any{
		"origin": "cells", "introducedAt": "2026-04-20T08:00:00Z", "reason": "aggression",
	}, 201, `"queen-000002"`, `"markColor":"white"`); err != nil {
		return err
	}
	if _, err := c.expect("GET", "/hives/hive-01/queens", nil, 200,
		`"queen-000001"`, `"superseded"`, `"queen-000002"`); err != nil {
		return err
	}
	if _, err := c.expect("POST", "/hives/hive-01/queens", map[string]any{
		"origin": "imported", "introducedAt": "2026-05-01T08:00:00Z", "reason": "aging",
	}, 400, `"code":"invalid_request"`); err != nil {
		return err
	}
	if _, err := c.expect("POST", "/hives/hive-01/state", map[string]any{"target": "retired"}, 200,
		`"to":"retired"`); err != nil {
		return err
	}
	if _, err := c.expect("POST", "/hives/hive-01/queens", map[string]any{
		"origin": "bred", "introducedAt": "2026-05-02T08:00:00Z", "reason": "loss",
	}, 409, `"code":"invalid_state"`); err != nil {
		return err
	}
	return nil
}

func checkAdvanceHiveState(c *client) error {
	if err := seedDefault(c); err != nil {
		return err
	}
	if _, err := c.expect("POST", "/hives/hive-01/state", map[string]any{"target": "production"}, 200,
		`"from":"colonized"`, `"to":"production"`); err != nil {
		return err
	}
	if _, err := c.expect("POST", "/hives/hive-01/state", map[string]any{"target": "colonized"}, 409,
		`"code":"invalid_state"`); err != nil {
		return err
	}
	if _, err := c.expect("POST", "/hives/hive-01/state", map[string]any{"target": "production"}, 409,
		`"code":"invalid_state"`); err != nil {
		return err
	}
	if _, err := c.expect("POST", "/hives/hive-01/state", map[string]any{"target": "retired"}, 200,
		`"to":"retired"`); err != nil {
		return err
	}
	if _, err := c.expect("POST", "/hives/hive-01/state", map[string]any{"target": "production"}, 409,
		`"code":"invalid_state"`); err != nil {
		return err
	}
	if _, err := c.expect("POST", "/hives/hive-missing/state", map[string]any{"target": "production"}, 404,
		`"code":"not_found"`); err != nil {
		return err
	}
	return nil
}

func checkReadApiaryStatus(c *client) error {
	if err := seedApiary(c, demoApiaryID, "北岭蜂场", "北岭片区", "周岚"); err != nil {
		return err
	}
	for _, hiveID := range []string{"hive-01", "hive-02", "hive-03"} {
		if err := seedHive(c, demoApiaryID, hiveID, "langstroth"); err != nil {
			return err
		}
	}
	if _, err := c.expect("POST", "/hives/hive-02/state", map[string]any{"target": "production"}, 200,
		`"to":"production"`); err != nil {
		return err
	}
	if _, err := c.expect("POST", "/hives/hive-01/inspections", map[string]any{
		"observedAt": "2026-04-10T09:00:00Z", "colonyStrength": 4, "broodPattern": "solid",
		"temper": "calm", "miteCount": 12, "note": "",
	}, 201, `"mite_load"`); err != nil {
		return err
	}
	report, err := c.expect("GET", "/apiaries/apiary-north/status?asOf=2026-04-12T00:00:00Z", nil, 200,
		`"hiveCount":3`, `"colonized":2`, `"production":1`, `"layingQueens":0`)
	if err != nil {
		return err
	}
	for _, fragment := range []string{
		`"needsInspection":["hive-02","hive-03"]`,
		`"concernHives":["hive-01"]`,
	} {
		if !strings.Contains(report.text(), fragment) {
			return fmt.Errorf("状态汇总缺少 %q：%s", fragment, report.text())
		}
	}
	if err := report.mustOrder(`"hiveId":"hive-01"`, `"hiveId":"hive-02"`); err != nil {
		return err
	}
	if _, err := c.expect("GET", "/apiaries/apiary-missing/status", nil, 404, `"code":"not_found"`); err != nil {
		return err
	}
	if _, err := c.expect("GET", "/apiaries/apiary-north/status?asOf=not-a-time", nil, 400,
		`"code":"invalid_request"`); err != nil {
		return err
	}
	return nil
}

// checkDistrictThresholds 校验协调人按片区覆盖关注项门槛：
// 覆盖只影响设定之后提交的巡检；已保存巡检与状态汇总保持原样；
// 覆盖缺失或非法时显式报错，不会悄悄退回默认门槛。
func checkDistrictThresholds(c *client) error {
	if err := seedApiary(c, "apiary-north", "北岭蜂场", "北岭片区", "周岚"); err != nil {
		return err
	}
	if err := seedApiary(c, "apiary-east", "东坡蜂场", "东坡片区", "林砚"); err != nil {
		return err
	}
	if err := seedHive(c, "apiary-north", "hive-01", "langstroth"); err != nil {
		return err
	}
	if err := seedHive(c, "apiary-east", "hive-02", "langstroth"); err != nil {
		return err
	}
	northPath := "/districts/" + url.PathEscape("北岭片区") + "/concern-thresholds"

	// 设定前两个片区都使用默认门槛。
	if _, err := c.expect("GET", northPath, nil, 200,
		`"source":"default"`, `"miteLoadThreshold":9`, `"lowStrengthCeiling":2`); err != nil {
		return err
	}
	if _, err := c.expect("GET", "/districts/"+url.PathEscape("东坡片区")+"/concern-thresholds", nil, 200,
		`"source":"default"`); err != nil {
		return err
	}

	// 调整前提交一次边界巡检：螨计数 12 在默认门槛 9 下记 mite_load，
	// 群势 2 在默认上限 2 下记 low_strength。
	if _, err := c.expect("POST", "/hives/hive-01/inspections", map[string]any{
		"observedAt": "2026-04-10T09:00:00Z", "colonyStrength": 2, "broodPattern": "solid",
		"temper": "calm", "miteCount": 12, "note": "调整前",
	}, 201, `"concerns":["mite_load","low_strength"]`); err != nil {
		return err
	}

	// 非法或不完整的覆盖必须被拒绝，不能退回默认门槛后冒充成功。
	if _, err := c.expect("PUT", northPath, map[string]any{
		"miteLoadThreshold": 61, "lowStrengthCeiling": 2,
	}, 400, `"code":"invalid_request"`, `"field":"miteLoadThreshold"`); err != nil {
		return err
	}
	if _, err := c.expect("PUT", northPath, map[string]any{
		"miteLoadThreshold": 15,
	}, 400, `"code":"invalid_request"`, `"field":"lowStrengthCeiling"`); err != nil {
		return err
	}
	if _, err := c.expect("PUT", "/districts/"+url.PathEscape("西荒片区")+"/concern-thresholds",
		map[string]any{"miteLoadThreshold": 15, "lowStrengthCeiling": 1},
		404, `"code":"not_found"`); err != nil {
		return err
	}
	// 被拒后门槛仍然是默认值，证明失败的设定没有产生副作用。
	if _, err := c.expect("GET", northPath, nil, 200, `"source":"default"`); err != nil {
		return err
	}

	// 合法覆盖整体生效，响应回显完整门槛与来源。
	if _, err := c.expect("PUT", northPath, map[string]any{
		"miteLoadThreshold": 15, "lowStrengthCeiling": 1,
	}, 200, `"district":"北岭片区"`, `"source":"override"`,
		`"miteLoadThreshold":15`, `"lowStrengthCeiling":1`, `"updatedAt"`); err != nil {
		return err
	}
	if _, err := c.expect("GET", northPath, nil, 200,
		`"source":"override"`, `"miteLoadThreshold":15`, `"lowStrengthCeiling":1`); err != nil {
		return err
	}

	// 调整后同样的边界巡检不再触发这两个关注项——新门槛只作用于新提交的巡检。
	if _, err := c.expect("POST", "/hives/hive-01/inspections", map[string]any{
		"observedAt": "2026-04-20T09:00:00Z", "colonyStrength": 2, "broodPattern": "solid",
		"temper": "calm", "miteCount": 12, "note": "调整后",
	}, 201, `"concerns":[]`); err != nil {
		return err
	}

	// 再用新门槛上的边界值确认覆盖确实按 15 / 1 生效。
	if _, err := c.expect("POST", "/hives/hive-01/inspections", map[string]any{
		"observedAt": "2026-04-21T09:00:00Z", "colonyStrength": 1, "broodPattern": "solid",
		"temper": "calm", "miteCount": 15, "note": "",
	}, 201, `"concerns":["mite_load","low_strength"]`); err != nil {
		return err
	}

	// 其他片区不受影响，仍然按默认门槛派生。
	if _, err := c.expect("POST", "/hives/hive-02/inspections", map[string]any{
		"observedAt": "2026-04-20T09:00:00Z", "colonyStrength": 2, "broodPattern": "solid",
		"temper": "calm", "miteCount": 12, "note": "",
	}, 201, `"concerns":["mite_load","low_strength"]`); err != nil {
		return err
	}

	// 已保存的巡检原样读回：旧巡检保留调整前派生的关注项，新巡检为空。
	listed, err := c.expect("GET", "/hives/hive-01/inspections", nil, 200, `"note":"调整前"`, `"note":"调整后"`)
	if err != nil {
		return err
	}
	var records []struct {
		Note     string   `json:"note"`
		Concerns []string `json:"concerns"`
	}
	if err := json.Unmarshal(listed.body, &records); err != nil {
		return fmt.Errorf("巡检列表不是合法 JSON：%w", err)
	}
	if len(records) != 3 {
		return fmt.Errorf("巡检列表期望 3 条，实际 %d：%s", len(records), listed.text())
	}
	if strings.Join(records[0].Concerns, ",") != "mite_load,low_strength" {
		return fmt.Errorf("调整前巡检的关注项被改动：%s", listed.text())
	}
	if len(records[1].Concerns) != 0 {
		return fmt.Errorf("调整后巡检不应有关注项：%s", listed.text())
	}

	// 状态汇总按最近一次巡检呈现：该巡检带关注项，蜂箱进 concernHives，
	// latestConcerns 与巡检读取结果必须是同一份固化内容。
	report, err := c.expect("GET", "/apiaries/apiary-north/status?asOf=2026-04-22T00:00:00Z", nil, 200,
		`"concernHives":["hive-01"]`,
		`"hiveId":"hive-01"`, `"latestConcerns":["mite_load","low_strength"]`)
	if err != nil {
		return err
	}
	if !strings.Contains(report.text(), `"concernHives":["hive-01"]`) {
		return fmt.Errorf("状态汇总关注项蜂箱与预期不符：%s", report.text())
	}
	return nil
}
