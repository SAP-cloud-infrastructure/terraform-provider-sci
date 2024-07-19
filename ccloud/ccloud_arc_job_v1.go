package ccloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sapcc/gophercloud-sapcc/v2/arc/v1/jobs"

	"github.com/gophercloud/gophercloud/v2"
)

type chefZeroPayload struct {
	RunList    []string                 `json:"run_list"`
	RecipeURL  string                   `json:"recipe_url"`
	Attributes map[string]interface{}   `json:"attributes,omitempty"`
	Debug      bool                     `json:"debug,omitempty"`
	Nodes      []map[string]interface{} `json:"nodes,omitempty"`
	NodeName   string                   `json:"name,omitempty"`
}

type chefEnableOptions struct {
	OmnitruckURL string `json:"omnitruck_url,omitempty"`
	ChefVersion  string `json:"chef_version"`
}

type tarballPayload struct {
	URL         string            `json:"url"`
	Path        string            `json:"path"`
	Arguments   []string          `json:"arguments,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
}

func arcCCloudArcJobV1BuildPayload(v []interface{}) (string, string, error) {
	var payload string

	for _, a := range v {
		if a != nil {
			action := a.(map[string]interface{})

			if v, ok := action["script"]; ok && len(v.(string)) > 0 {
				return "script", v.(string), nil
			}

			if v, ok := action["tarball"]; ok && len(v.([]interface{})) > 0 {
				v, err := arcCCloudArcJobV1ParseTarball(v.([]interface{}))
				return "tarball", v, err
			}

			if v, ok := action["enable"]; ok && len(v.([]interface{})) > 0 {
				v, err := arcCCloudArcJobV1ParseChefEnable(v.([]interface{}))
				return "enable", v, err
			}

			if v, ok := action["zero"]; ok && len(v.([]interface{})) > 0 {
				v, err := arcCCloudArcJobV1ParseChefZero(v.([]interface{}))
				return "zero", v, err
			}
		}
	}

	return "", payload, nil
}

func arcCCloudArcJobV1ParseTarball(v []interface{}) (string, error) {
	var payload string

	for _, t := range v {
		if t != nil {
			var tarball tarballPayload
			tmp := t.(map[string]interface{})

			if val, ok := tmp["url"]; ok {
				tarball.URL = val.(string)
			}
			if val, ok := tmp["path"]; ok {
				tarball.Path = val.(string)
			}
			if val, ok := tmp["arguments"]; ok {
				tarball.Arguments = expandToStringSlice(val.([]interface{}))
			}
			if val, ok := tmp["environment"]; ok {
				tarball.Environment = expandToMapStringString(val.(map[string]interface{}))
			}

			bytes, err := json.Marshal(tarball)
			if err != nil {
				return "", err
			}
			payload = string(bytes[:])
		}
	}

	return payload, nil
}

func arcCCloudArcJobV1ParseChefEnable(v []interface{}) (string, error) {
	var payload string

	for _, c := range v {
		if c != nil {
			var chefEnable chefEnableOptions
			chef := c.(map[string]interface{})

			if val, ok := chef["omnitruck_url"]; ok {
				chefEnable.OmnitruckURL = val.(string)
			}
			if val, ok := chef["chef_version"]; ok {
				chefEnable.ChefVersion = val.(string)
			}

			bytes, err := json.Marshal(chefEnable)
			if err != nil {
				return "", err
			}
			payload = string(bytes[:])
		}
	}

	return payload, nil
}

func arcCCloudArcJobV1ParseChefZero(v []interface{}) (string, error) {
	var payload string

	for _, c := range v {
		if c != nil {
			var chefZero struct {
				chefZeroPayload
				chefEnableOptions
			}
			chef := c.(map[string]interface{})

			if val, ok := chef["run_list"]; ok {
				chefZero.RunList = expandToStringSlice(val.([]interface{}))
			}
			if val, ok := chef["recipe_url"]; ok {
				chefZero.RecipeURL = val.(string)
			}
			if val, ok := chef["attributes"]; ok {
				err := json.Unmarshal([]byte(val.(string)), &chefZero.Attributes)
				if err != nil {
					return "", err
				}
			}
			if val, ok := chef["debug"]; ok {
				chefZero.Debug = val.(bool)
			}
			if val, ok := chef["nodes"]; ok {
				err := json.Unmarshal([]byte(val.(string)), &chefZero.Nodes)
				if err != nil {
					return "", err
				}
			}
			if val, ok := chef["node_name"]; ok {
				chefZero.NodeName = val.(string)
			}
			if val, ok := chef["omnitruck_url"]; ok {
				chefZero.OmnitruckURL = val.(string)
			}
			if val, ok := chef["chef_version"]; ok {
				chefZero.ChefVersion = val.(string)
			}

			bytes, err := json.Marshal(chefZero)
			if err != nil {
				return "", err
			}
			payload = string(bytes[:])
		}
	}

	return payload, nil
}

func arcCCloudArcJobV1FlattenExecute(job *jobs.Job) ([]map[string]interface{}, error) {
	if !strSliceContains([]string{"tarball", "script"}, job.Action) {
		return []map[string]interface{}{}, nil
	}

	if job.Action == "script" {
		return []map[string]interface{}{{
			"script":  job.Payload,
			"tarball": []map[string]interface{}{},
		}}, nil
	}

	var tarball tarballPayload

	err := json.Unmarshal([]byte(job.Payload), &tarball)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal execute %s payload: %v", job.Action, err)
	}

	return []map[string]interface{}{{
		"script": "",
		"tarball": []map[string]interface{}{{
			"url":         tarball.URL,
			"path":        tarball.Path,
			"arguments":   tarball.Arguments,
			"environment": tarball.Environment,
		}},
	}}, nil
}

func arcCCloudArcJobV1FlattenChef(job *jobs.Job) ([]map[string]interface{}, error) {
	if !strSliceContains([]string{"zero", "enable"}, job.Action) {
		return []map[string]interface{}{}, nil
	}

	var chef struct {
		chefZeroPayload
		chefEnableOptions
	}

	err := json.Unmarshal([]byte(job.Payload), &chef)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal chef %s payload: %v", job.Action, err)
	}

	if job.Action == "enable" {
		return []map[string]interface{}{{
			"enable": []map[string]interface{}{{
				"omnitruck_url": chef.OmnitruckURL,
				"chef_version":  chef.ChefVersion,
			}},
			"zero": []map[string]interface{}{},
		}}, nil
	}

	attributes, err := json.Marshal(chef.Attributes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chef attributes: %v", err)
	}
	nodes, err := json.Marshal(chef.Nodes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chef nodes: %v", err)
	}

	return []map[string]interface{}{{
		"enable": []map[string]interface{}{},
		"zero": []map[string]interface{}{{
			"run_list":      chef.RunList,
			"recipe_url":    chef.RecipeURL,
			"attributes":    string(attributes[:]),
			"debug":         chef.Debug,
			"nodes":         string(nodes[:]),
			"node_name":     chef.NodeName,
			"omnitruck_url": chef.OmnitruckURL,
			"chef_version":  chef.ChefVersion,
		}},
	}}, nil
}

func arcCCloudArcJobV1Filter(ctx context.Context, d *schema.ResourceData, arcClient *gophercloud.ServiceClient, resourceName string) ([]jobs.Job, error) {
	agentID := d.Get("agent_id").(string)
	timeout := d.Get("timeout").(int)
	agent := d.Get("agent").(string)
	action := d.Get("action").(string)
	status := d.Get("status").(string)

	listOpts := jobs.ListOpts{AgentID: agentID}

	log.Printf("[DEBUG] %s list options: %#v", resourceName, listOpts)

	allPages, err := jobs.List(arcClient, listOpts).AllPages(ctx)
	if err != nil {
		return nil, fmt.Errorf("Unable to list %s: %v", resourceName, err)
	}

	allJobs, err := jobs.ExtractJobs(allPages)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve %s: %v", resourceName, err)
	}

	var jobs []jobs.Job
	for _, job := range allJobs {
		found := true
		if found && timeout > 0 && job.Timeout != timeout {
			found = false
		}
		if found && len(agent) > 0 && job.Agent != agent {
			found = false
		}
		if found && len(action) > 0 && job.Action != action {
			found = false
		}
		if found && len(status) > 0 && job.Status != status {
			found = false
		}

		if found {
			jobs = append(jobs, job)
		}
	}

	return jobs, nil
}

func flattenArcJobUserV1(user jobs.User) []interface{} {
	return []interface{}{map[string]interface{}{
		"id":          user.ID,
		"name":        user.Name,
		"domain_id":   user.DomainID,
		"domain_name": user.DomainName,
		"roles":       user.Roles,
	}}
}

func waitForArcJobV1(ctx context.Context, arcClient *gophercloud.ServiceClient, id string, target []string, pending []string, timeout time.Duration) error {
	log.Printf("[DEBUG] Waiting for %s job to become %v.", id, target)

	stateConf := &retry.StateChangeConf{
		Target:     target,
		Pending:    pending,
		Refresh:    arcJobV1GetStatus(ctx, arcClient, id),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func arcJobV1GetStatus(ctx context.Context, arcClient *gophercloud.ServiceClient, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		job, err := jobs.Get(ctx, arcClient, id).Extract()
		if err != nil {
			return nil, "", fmt.Errorf("Unable to retrieve %s ccloud_arc_job_v1: %v", id, err)
		}

		return job, job.Status, nil
	}
}

func arcJobV1GetLog(ctx context.Context, arcClient *gophercloud.ServiceClient, id string) []byte {
	var err error
	l := []byte("Log not available")

	res := jobs.GetLog(ctx, arcClient, id)
	if res.Err != nil {
		log.Printf("[DEBUG] Error retrieving logs for %s ccloud_arc_job_v1: %s", id, res.Err)
		return l
	}

	logData, err := res.ExtractContent()
	if err != nil {
		log.Printf("[DEBUG] Error extracting logs for %s ccloud_arc_job_v1: %v", id, err)
		return l
	}

	return logData
}
