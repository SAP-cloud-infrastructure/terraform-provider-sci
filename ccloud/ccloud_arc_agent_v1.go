package ccloud

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sapcc/gophercloud-sapcc/v2/arc/v1/agents"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
)

func arcCCloudArcAgentV1ReadAgent(ctx context.Context, d *schema.ResourceData, arcClient *gophercloud.ServiceClient, agent *agents.Agent, region string) {
	if len(agent.Facts) == 0 {
		facts, err := agents.GetFacts(ctx, arcClient, agent.AgentID).Extract()
		if err != nil {
			log.Printf("Unable to retrieve facts for ccloud_arc_agent_v1: %s", err)
		}
		agent.Facts = facts
		log.Printf("[DEBUG] Retrieved ccloud_arc_agent_v1 facts %s: %+v", agent.AgentID, agent.Facts)
	}

	d.Set("display_name", agent.DisplayName)
	d.Set("agent_id", agent.AgentID)
	d.Set("project", agent.Project)
	d.Set("organization", agent.Organization)
	d.Set("all_tags", agent.Tags)
	d.Set("created_at", agent.CreatedAt.Format(time.RFC3339))
	d.Set("updated_at", agent.UpdatedAt.Format(time.RFC3339))
	d.Set("updated_with", agent.UpdatedWith)
	d.Set("updated_by", agent.UpdatedBy)

	d.Set("facts", expandToMapStringString(agent.Facts))
	factsAgents := agent.Facts["agents"]
	if v, ok := factsAgents.(map[string]interface{}); ok {
		d.Set("facts_agents", expandToMapStringString(v))
	} else {
		d.Set("facts_agents", map[string]string{})
	}

	d.Set("region", region)
}

func arcCCloudArcAgentV1WaitForAgent(ctx context.Context, arcClient *gophercloud.ServiceClient, agentID, filter string, timeout time.Duration) (*agents.Agent, error) {
	var agent interface{}
	var msg string
	var err error

	// This condition is required, otherwise zero timeout will always raise:
	// "timeout while waiting for state to become 'active'"
	if timeout > 0 {
		// Retryable case, when timeout is set
		waitForAgent := &retry.StateChangeConf{
			Target:         []string{"active"},
			Refresh:        arcCCloudArcAgentV1GetAgent(ctx, arcClient, agentID, filter, timeout),
			Timeout:        timeout,
			Delay:          1 * time.Second,
			MinTimeout:     1 * time.Second,
			NotFoundChecks: 1000, // workaround for default 20 retries, when the resource is nil
		}
		agent, err = waitForAgent.WaitForStateContext(ctx)
	} else {
		// When timeout is not set, just get the agent
		agent, msg, err = arcCCloudArcAgentV1GetAgent(ctx, arcClient, agentID, filter, timeout)()
	}

	if len(msg) > 0 && msg != "active" {
		return nil, fmt.Errorf(msg)
	}

	if err != nil {
		return nil, err
	}

	return agent.(*agents.Agent), nil
}

func arcCCloudArcAgentV1GetAgent(ctx context.Context, arcClient *gophercloud.ServiceClient, agentID, filter string, timeout time.Duration) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		var agent *agents.Agent
		var err error

		if len(agentID) == 0 && len(filter) == 0 {
			return nil, "", fmt.Errorf("At least one of agent_id or filter parameters is expected in ccloud_arc_agent_v1")
		}

		if len(agentID) > 0 {
			agent, err = agents.Get(ctx, arcClient, agentID).Extract()
			if err != nil {
				if gophercloud.ResponseCodeIs(err, http.StatusNotFound) && timeout > 0 {
					// Retryable case, when timeout is set
					return nil, fmt.Sprintf("Unable to retrieve %s ccloud_arc_agent_v1: %v", agentID, err), nil
				}
				return nil, "", fmt.Errorf("Unable to retrieve %s ccloud_arc_agent_v1: %v", agentID, err)
			}
		} else {
			listOpts := agents.ListOpts{Filter: filter}

			log.Printf("[DEBUG] ccloud_arc_agent_v1 list options: %#v", listOpts)

			allPages, err := agents.List(arcClient, listOpts).AllPages(ctx)
			if err != nil {
				return nil, "", fmt.Errorf("Unable to list ccloud_arc_agent_v1: %s", err)
			}

			allAgents, err := agents.ExtractAgents(allPages)
			if err != nil {
				return nil, "", fmt.Errorf("Unable to retrieve ccloud_arc_agent_v1: %s", err)
			}

			if len(allAgents) == 0 {
				// Retryable case, when timeout is set
				return nil, "No ccloud_arc_agent_v1 found", nil
			}

			if len(allAgents) > 1 {
				return nil, "", fmt.Errorf("More than one ccloud_arc_agent_v1 found (%d)", len(allAgents))
			}

			agent = &allAgents[0]
		}

		log.Printf("[DEBUG] Retrieved ccloud_arc_agent_v1 %s: %+v", agent.AgentID, *agent)

		return agent, "active", nil
	}
}

func updateArcAgentTagsV1(ctx context.Context, arcClient *gophercloud.ServiceClient, agentID string, oldTagsRaw, newTagsRaw interface{}) error {
	var tagsToDelete []string
	oldTags, _ := oldTagsRaw.(map[string]interface{})
	newTags, _ := newTagsRaw.(map[string]interface{})

	// Determine if any tag keys were removed from the configuration.
	// Then request those keys to be deleted.
	for oldKey := range oldTags {
		var found bool
		for newKey := range newTags {
			if oldKey == newKey {
				found = true
			}
		}

		if !found {
			tagsToDelete = append(tagsToDelete, oldKey)
		}
	}

	for _, key := range tagsToDelete {
		err := agents.DeleteTag(ctx, arcClient, agentID, key).ExtractErr()
		if err != nil {
			return fmt.Errorf("Error deleting %s tag from %s ccloud_arc_agent_v1: %v", key, agentID, err)
		}
	}

	// Update existing tags and add any new tags.
	tagsOpts := make(agents.Tags)
	for k, v := range newTags {
		tagsOpts[k] = v.(string)
	}

	err := agents.CreateTags(ctx, arcClient, agentID, tagsOpts).ExtractErr()
	if err != nil {
		return fmt.Errorf("Error updating tags for %s ccloud_arc_agent_v1: %v", agentID, err)
	}

	return nil
}

func arcAgentV1ParseTimeout(raw interface{}) (time.Duration, error) {
	if list, ok := raw.([]interface{}); ok {
		for _, t := range list {
			if timeout, ok := t.(map[string]interface{}); ok {
				if v, ok := timeout["read"]; ok {
					return time.ParseDuration(v.(string))
				}
			}
		}
	}

	return time.Duration(0), nil
}

func serverV2StateRefreshFunc(ctx context.Context, client *gophercloud.ServiceClient, instanceID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		s, err := servers.Get(ctx, client, instanceID).Extract()
		if err != nil {
			if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
				return s, "DELETED", nil
			}
			return nil, "", err
		}

		return s, s.Status, nil
	}
}
