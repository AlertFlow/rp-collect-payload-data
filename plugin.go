package main

import (
	"encoding/json"
	"time"

	"gitlab.justlab.xyz/alertflow-public/runner/pkg/executions"
	"gitlab.justlab.xyz/alertflow-public/runner/pkg/models"
	"gitlab.justlab.xyz/alertflow-public/runner/pkg/payloads"

	log "github.com/sirupsen/logrus"
)

type CollectPayloadDataPlugin struct{}

func (p *CollectPayloadDataPlugin) Init() models.Plugin {
	return models.Plugin{
		Name:    "Collect Payload Data",
		Type:    "action",
		Version: "1.0.1",
		Creator: "JustNZ",
	}
}

func (p *CollectPayloadDataPlugin) Details() models.ActionDetails {
	params := []models.Param{
		{
			Key:         "PayloadID",
			Type:        "text",
			Default:     "00000000-0000-0000-0000-00000000",
			Required:    true,
			Description: "The Payload ID to collect data from",
		},
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		log.Error(err)
	}

	return models.ActionDetails{
		Name:        "Collect Payload Data",
		Description: "Collects Payload data from AlertFlow",
		Icon:        "solar:letter-opened-broken",
		Type:        "collect_payload_data",
		Category:    "Data",
		Function:    p.Execute,
		IsHidden:    true,
		Params:      json.RawMessage(paramsJSON),
	}
}

func (p *CollectPayloadDataPlugin) Execute(execution models.Execution, flow models.Flows, payload models.Payload, steps []models.ExecutionSteps, step models.ExecutionSteps, action models.Actions) (data map[string]interface{}, finished bool, canceled bool, no_pattern_match bool, failed bool) {
	payloadID := ""

	if action.Params == nil {
		payloadID = execution.PayloadID
	} else {
		for _, param := range action.Params {
			if param.Key == "PayloadID" {
				payloadID = param.Value
			}
		}
	}

	err := executions.UpdateStep(execution.ID.String(), models.ExecutionSteps{
		ID:             step.ID,
		ActionID:       action.ID.String(),
		ActionMessages: []string{"Collecting payload data from AlertFlow"},
		Pending:        false,
		Running:        true,
		StartedAt:      time.Now(),
	})
	if err != nil {
		log.Error("Error updating step: ", err)
	}

	payload, err = payloads.GetData(payloadID)
	if err != nil {
		err := executions.UpdateStep(execution.ID.String(), models.ExecutionSteps{
			ID:             step.ID,
			ActionMessages: []string{"Failed to get Payload Data"},
			Error:          true,
			Finished:       true,
			Running:        false,
			FinishedAt:     time.Now(),
		})
		if err != nil {
			log.Error("Error updating step: ", err)
		}

		return nil, false, false, false, true
	}

	err = executions.UpdateStep(execution.ID.String(), models.ExecutionSteps{
		ID:             step.ID,
		ActionMessages: []string{"Payload Data collected"},
		Running:        false,
		Finished:       true,
		FinishedAt:     time.Now(),
	})
	if err != nil {
		log.Error("Error updating step: ", err)
	}

	return map[string]interface{}{"payload": payload}, true, false, false, false
}

var Plugin CollectPayloadDataPlugin
