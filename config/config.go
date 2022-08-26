package config

type Config struct {
	TdServer string
	MdServer string
	BrokerID string
	User     string
	Password string
	AppID    string
	AuthCode string
	DB       struct {
		Type string
		Uri  string
	}
	Taos string
}
