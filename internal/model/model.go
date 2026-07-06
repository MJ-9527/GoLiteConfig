package model

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type PublishConfigRequest struct {
	App       string            `json:"app"`
	Env       string            `json:"env"`
	Configs   map[string]string `json:"configs"`
	Publisher string            `json:"publisher"`
	Comment   string            `json:"comment"`
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

type GetConfigResponse struct {
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
type ListConfigsRequest struct {
	App      string       `json:"app"`
	Env      string       `json:"env"`
	Current  string       `json:"current"`
	Versions []ConfigMeta `json:"versions"`
}

type RollbackRequest struct {
	App           string `json:"app"`
	Env           string `json:"env"`
	TargetVersion string `json:"target_version"`
	Publisher     string `json:"publisher"`
	Comment       string `json:"comment"`
}

type RollbackResponse struct {
	App           string `json:"app"`
	Env           string `json:"env"`
	FromVersion   string `json:"from_version"`
	TargetVersion string `json:"target_version"`
	NewVersion    string `json:"new_version"`
	Revision      int64  `json:"revision"`
}

type WatchConfigsResponse struct {
	App      string            `json:"app"`
	Env      string            `json:"env"`
	Version  string            `json:"version"`
	Revision int64             `json:"revision"`
	Configs  map[string]string `json:"configs"`
}
