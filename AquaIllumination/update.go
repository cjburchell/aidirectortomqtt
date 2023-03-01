package AquaIllumination

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
)

type AiResponse struct {
	Status      bool            `json:"response_status"`
	Code        int             `json:"response_code"`
	Description string          `json:"response_desc"`
	Response    json.RawMessage `json:"response"`
}

func GetAIData(client http.Client, host string, path string) (*AiResponse, error) {
	url := fmt.Sprintf("http://%s/api/%s", host, path)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.WithStack(fmt.Errorf("%s repsponse not OK: %d", url, resp.StatusCode))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data AiResponse
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		return nil, err
	}

	if !data.Status {
		return nil, errors.WithStack(fmt.Errorf("%s unexpected reply: %s", url, data.Description))
	}

	return &data, nil

}

func GetAll(client http.Client, host string) (*Director, error) {
	var director Director
	data, err := GetAIData(client, host, "controller/version")
	if err != nil {
		return nil, err
	}

	var result []map[string]string
	json.Unmarshal(data.Response, &result)
	director.Version = result[0]["version"]
	director.Name = result[0]["package"]

	data, err = GetAIData(client, host, "tanks")
	if err != nil {
		return nil, err
	}
	var tanks []AiTank
	json.Unmarshal(data.Response, &tanks)
	director.Tank = make(map[int]AiTank)
	for _, tank := range tanks {
		director.Tank[tank.TankId] = tank
	}

	data, err = GetAIData(client, host, "devices/statistics")
	if err != nil {
		return nil, err
	}
	var stats []AiDeviceStats
	json.Unmarshal(data.Response, &stats)
	director.DeviceStats = make(map[int]AiDeviceStats)
	director.Devices = make(map[int]AiDevice)
	for _, deviceStats := range stats {
		director.DeviceStats[deviceStats.Id] = deviceStats

		deviceData, err := GetAIData(client, host, fmt.Sprintf("devices/%d/info", deviceStats.Id))
		if err != nil {
			return nil, err
		}

		var device AiDevice
		json.Unmarshal(deviceData.Response, &device)
		director.Devices[deviceStats.Id] = device
	}

	director.GroupColors = make(map[int]AiColors)
	director.GroupMode = make(map[int]bool)
	for _, tank := range tanks {
		for _, group := range tank.Groups {
			data, err = GetAIData(client, host, fmt.Sprintf("groups/%d/led/intensity/colors", group.GroupId))
			if err != nil {
				return nil, err
			}

			var colors AiColors
			json.Unmarshal(data.Response, &colors)
			director.GroupColors[group.GroupId] = colors

			data, err = GetAIData(client, host, fmt.Sprintf("groups/%d/mode", group.GroupId))
			if err != nil {
				return nil, err
			}

			var mode AiMode
			json.Unmarshal(data.Response, &mode)
			director.GroupMode[group.GroupId] = mode.Mode
		}
	}

	return &director, nil
}
