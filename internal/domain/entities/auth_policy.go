package entities

const (
	AuthTypeBasic  string = "basic"
	AuthTypeBearer string = "bearer"
	AuthTypeAPIKey string = "api"
	AuthTypeNone   string = "none"
)

type AuthPolicy struct {
	Type    string                 `json:"type"`
	Enabled bool                   `json:"enabled"`
	Config  map[string]interface{} `json:"config"`
}

func (ap *AuthPolicy) RequiresAuth() bool {
	if ap.Enabled == false {
		return false
	}

	if ap.Type == AuthTypeNone {
		return false
	}
	return true
}

func (ap *AuthPolicy) GetType() string {
	return ap.Type
}

func (ap *AuthPolicy) Validate() error {

	return nil
}
