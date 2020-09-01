package ksyun

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-ksyun/logger"
	"strconv"
)

// redis security group List
func dataSourceRedisSecurityGroups() *schema.Resource {
	return &schema.Resource{
		// redis security group List List Query Function
		Read: dataSourceRedisSecurityGroupsRead,
		// Define input and output parameters
		Schema: map[string]*schema.Schema{
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"instances": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"created": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"updated": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceRedisSecurityGroupsRead(d *schema.ResourceData, meta interface{}) error {
	var (
		allInstances []interface{}
		az           map[string]string
		limit        = 100
		nextToken    string
		err          error
	)

	action := "DescribeSecurityGroups"
	conn := meta.(*KsyunClient).kcsv1conn
	readReq := make(map[string]interface{})
	if az, err = queryAz(conn); err != nil {
		return fmt.Errorf("error on reading instances, because there is no available area in the region")
	}
	for k := range az {
		readReq["AvailableZone"] = k
		for {
			readReq["Limit"] = fmt.Sprintf("%v", limit)
			if nextToken != "" {
				readReq["Offset"] = nextToken
			}
			logger.Debug(logger.ReqFormat, action, readReq)
			resp, err := conn.DescribeSecurityGroups(&readReq)
			if err != nil {
				return fmt.Errorf("error on reading redis security group list req(%v):%s", readReq, err)
			}
			logger.Debug(logger.RespFormat, action, readReq, *resp)
			result, ok := (*resp)["Data"]
			if !ok {
				break
			}
			item, ok := result.(map[string]interface{})
			if !ok {
				break
			}
			items, ok := item["list"].([]interface{})
			if !ok {
				break
			}
			if items == nil || len(items) < 1 {
				break
			}
			allInstances = append(allInstances, items...)
			if len(items) < limit {
				break
			}
			nextToken = strconv.Itoa(int(item["limit"].(float64)) + int(item["offset"].(float64)))
		}
	}

	values := GetSubSliceDByRep(allInstances, redisSecKeys)
	if err := dataSourceKscSave(d, "instances", []string{}, values); err != nil {
		return fmt.Errorf("error on save redis security group list, %s", err)
	}
	return nil
}
