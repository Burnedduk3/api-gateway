package entities

const (
	AuthTypeAPIKey string = "api"
	AuthTypeNone   string = "none"
)

type AuthPolicy struct {
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
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
