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
type ListVersionsResponse struct {
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

type WatchConfigResponse struct {
	App      string            `json:"app"`
	Env      string            `json:"env"`
	Version  string            `json:"version"`
	Revision int64             `json:"revision"`
	Configs  map[string]string `json:"configs"`
}

type DeleteVersionsRequest struct {
	App      string   `json:"app"`
	Env      string   `json:"env"`
	Version  string   `json:"version"`
	Versions []string `json:"versions"`
}

type DeleteVersionsResponse struct {
	App      string   `json:"app"`
	Env      string   `json:"env"`
	Current  string   `json:"current"`
	Deleted  []string `json:"deleted"`
	Skipped  []string `json:"skipped,omitempty"`
	Revision int64    `json:"revision"`
}

type ConfigDiffResponse struct {
	App         string              `json:"app"`
	Env         string              `json:"env"`
	FromVersion string              `json:"from_version"`
	ToVersion   string              `json:"to_version"`
	Added       map[string]string   `json:"added"`
	Removed     map[string]string   `json:"removed"`
	Modified    map[string]DiffItem `json:"modified"`
}

type DiffItem struct {
	From string `json:"from"`
	To   string `json:"to"`
}
