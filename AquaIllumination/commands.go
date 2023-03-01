package AquaIllumination

import (
	"bytes"
	"encoding/json"
	"fmt"
	logger "github.com/cjburchell/uatu-go"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"strconv"
)

func SetAIData(client http.Client, host string, path string, payload interface{}) error {
	url := fmt.Sprintf("http://%s/api/%s", host, path)

	jsonBody, _ := json.Marshal(payload)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.WithStack(fmt.Errorf("%s repsponse not OK: %d", url, resp.StatusCode))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var data AiResponse
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		return err
	}

	if !data.Status {
		return errors.WithStack(fmt.Errorf("%s unexpected reply: %s", url, data.Description))
	}

	return nil
}

func SetManualMode(groupId int, toggle bool, client http.Client, log logger.ILog, host string) {
	var mode AiMode
	mode.Mode = !toggle
	err := SetAIData(client, host, fmt.Sprintf("groups/%d/mode", groupId), mode)
	if err != nil {
		log.Error(err, "unable to set mode for group %d", groupId)
	}
}

func ToggleLight(groupId int, colorId string, toggle bool, client http.Client, log logger.ILog, host string) {
	var colors AiColors
	colors.Colors = make([]AiColor, 1)
	colors.Colors[0].Color = colorId
	if toggle {
		colors.Colors[0].Intensity = 100
	} else {
		colors.Colors[0].Intensity = 0
	}

	err := SetAIData(client, host, fmt.Sprintf("groups/%d/led/intensity/colors", groupId), colors)
	if err != nil {
		log.Error(err, "unable to set mode for group %d", groupId)
	}
}

func SetIntensity(groupId int, colorId string, payload string, client http.Client, log logger.ILog, host string) {
	var colors AiColors
	colors.Colors = make([]AiColor, 1)
	colors.Colors[0].Color = colorId
	intensity, _ := strconv.Atoi(payload)
	colors.Colors[0].Intensity = intensity
	err := SetAIData(client, host, fmt.Sprintf("groups/%d/led/intensity/colors", groupId), colors)
	if err != nil {
		log.Error(err, "unable to set mode for group %d", groupId)
	}
}
