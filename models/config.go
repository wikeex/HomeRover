package models

type ControllerConfig struct {
	Server 		Server		`json:"server"`
	Local		Local		`json:"local"`
	Joystick	Joystick	`json:"joystick"`
}

type Server struct {
	IP			string		`json:"ip"`
	Port		int			`json:"port"`
}

type Local struct {
	Port		int			`json:"port"`
}

type Joystick struct {
	ReadFreq	int			`json:"readFreq"`
}
