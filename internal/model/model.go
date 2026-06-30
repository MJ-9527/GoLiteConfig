package model

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type PublishConfigRequest struct {
	App     string            `json:"app"`
	Env     string            `json:"env"`
	Config  map[string]string `json:"config"`
	Publish string            `json:"publish"`
	Comment string            `json:"comment"`
}

type PublishConfigResponse struct {
	App      string `json:"app"`
	Env      string `json:"env"`
	Version  string `json:"version"`
	Revision int64  `json:"revision"`
}

type GetConfigRequest struct {
	App      string            `json:"app"`
	Env      string            `json:"env"`
	Version  string            `json:"version"`
	Revision int64             `json:"revision"`
	Configs  map[string]string `json:"configs"`
}

type ConfigMeta struct {
	Version   string `json:"version"`
	Revision  int64  `json:"revision"`
	Publisher string `json:"publisher"`
	Comment   string `json:"comment"`
	CreatedAt int64  `json:"created_at"`
}
