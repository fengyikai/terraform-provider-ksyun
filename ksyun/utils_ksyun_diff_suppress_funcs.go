package ksyun

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-ksyun/logger"
	"strings"
)

func purchaseTimeDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	if v, ok := d.GetOk("charge_type"); ok && (v.(string) == "Monthly" || v.(string) == "PrePaidByMonth") {
		return false
	}
	return true
}

func chargeSchemaDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	mappings := map[string]string{
		"PostPaidByPeak":     "Peak",
		"PostPaidByDay":      "Daily",
		"PostPaidByTransfer": "TrafficMonthly",
		"PrePaidByMonth":     "Monthly",
		"Peak":               "PostPaidByPeak",
		"Daily":              "PostPaidByDay",
		"TrafficMonthly":     "PostPaidByTransfer",
		"Monthly":            "PrePaidByMonth",
	}
	if old == new {
		return true
	}
	if v, ok := mappings[old]; ok && v == new {
		return true
	}
	return false
}

func kecImportDiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	//由于一些字段暂时无法支持从查询中返回 所以现在设立做特殊处理拦截变更 用来适配导入的场景 后续支持后在对导入场景做优化
	if !d.IsNewResource() {
		if !d.Get("has_init_info").(bool) {
			if k == "local_volume_snapshot_id" {
				return true
			}
			if k == "user_data" {
				return true
			}
		}
	}
	if (k == "keep_image_login" || k == "key_id") && !d.IsNewResource() && !d.HasChange("image_id") {
		return true
	}

	return false
}

func kcsParameterDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	if old != "" && new == "" {
		return true
	}
	if k == "parameters.notify-keyspace-events" && old == "" && new == "" {
		return true
	}
	return false
}

func rdsParameterDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	if k == "parameters.#" {
		logger.Debug(logger.RespFormat, "DemoTest", d.ConnInfo())
		logger.Debug(logger.RespFormat, "DemoTest", d.Get("parameters"))
		return false
	}
	return true
}

func kcsSecurityGroupDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	if d.Get("security_group_ids") != nil && old == "" && new != "" {
		if sgs, ok := d.Get("security_group_ids").(*schema.Set); ok {
			if (*sgs).Contains(new) {
				err := d.Set("security_group_id", new)
				if err == nil {
					return true
				}
			}
		}
	}
	return false
}

func networkAclEntryDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	if d.Get("protocol") != "icmp" && (k == "icmp_type" || k == "icmp_code") {
		return true
	}
	if d.Get("protocol") != "tcp" && d.Get("protocol") != "udp" && (k == "port_range_from" || k == "port_range_to") {
		return true
	}
	return false
}

func securityGroupEntryDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	if d.Get("protocol") != "icmp" && (k == "icmp_type" || k == "icmp_code") {
		return true
	}
	if d.Get("protocol") != "tcp" && d.Get("protocol") != "udp" && (k == "port_range_from" || k == "port_range_to") {
		return true
	}
	return false
}

func loadBalancerDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	if d.Get("type") != "internal" && (k == "subnet_id" || k == "private_ip_address") {
		return true
	}
	return false
}

func lbListenerDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	if d.Get("listener_protocol") != "HTTPS" && (k == "certificate_id" || k == "tls_cipher_policy" || k == "enable_http2") {
		return true
	}
	if d.Get("listener_protocol") != "HTTP" && k == "redirect_listener_id" {
		return true
	}
	if d.Get("listener_protocol") != "HTTPS" && d.Get("listener_protocol") != "HTTP" && k == "http_protocol" {
		return true
	}

	return false
}

func lbHealthCheckDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	if d.Get("listener_protocol") != "" && d.Get("listener_protocol") != "HTTP" && d.Get("listener_protocol") != "HTTPS" && (k == "url_path" || k == "host_name" || k == "is_default_host_name") {
		return true
	}
	if d.Get("host_name") != "" && k == "is_default_host_name" {
		return true
	}
	if k == "host_name" && d.Get("is_default_host_name").(bool) {
		return true
	}
	return false
}

func lbRuleDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	if d.Get("listener_sync") != "off" && (k == "method" || strings.HasPrefix(k, "session.") || strings.HasPrefix(k, "health_check.")) {
		return true
	}
	if k == "session.0.cookie_name" && d.Get("session.0.cookie_type") != "RewriteCookie" {
		return true
	}
	if k == "health_check.0.host_name" && d.Get("health_check.0.is_default_host_name").(bool) {
		return true
	}
	return false
}
