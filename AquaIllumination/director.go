package AquaIllumination

type AiTank struct {
	TankId   int    `json:"tank_id"`
	TankName string `json:"tank_name"`
	Groups   []struct {
		GroupId     int `json:"group_id"`
		GroupNumber int `json:"group_number"`
	} `json:"groups"`
}

type AiDeviceStats struct {
	Id          int `json:"id"`
	CommPercent int `json:"comm_percent"`
}

type AiDevice struct {
	DeviceId    int    `json:"device_id"`
	DeviceType  string `json:"device_type"`
	Model       int    `json:"model"`
	DeviceName  string `json:"device_name"`
	GroupId     int    `json:"group_id"`
	Temperature []int  `json:"temperature"`
	FanSpeed    string `json:"fan_speed"`
}

type AiColor struct {
	Color     string `json:"color"`
	Intensity int    `json:"intensity"`
}

type AiColors struct {
	Colors []AiColor `json:"colors"`
}

type AiMode struct {
	Mode bool `json:"mode"`
}

type Director struct {
	Version     string
	Name        string
	Tank        map[int]AiTank
	Devices     map[int]AiDevice
	DeviceStats map[int]AiDeviceStats
	GroupColors map[int]AiColors
	GroupMode   map[int]bool
}
