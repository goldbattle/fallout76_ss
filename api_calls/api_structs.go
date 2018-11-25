package api_calls

type ExternalServerStatusPlatform struct {
	Message  string                 `json:"message"`
	Code     int                    `json:"code"`
	Response map[string]interface{} `json:"response"`
}

type ExternalServerStatus struct {
	Platform ExternalServerStatusPlatform `json:"platform"`
}
