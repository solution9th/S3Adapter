package config

// Config config struct
type Config struct {
	Server Server
	MySQL  MySQL
}

// Server server config
type Server struct {
	HTTPPort  string
	PprofPort string
	EndPoint  string
	Region    string
	LogPath   string
	IsDebug   bool
}

// MySQL mysql config
type MySQL struct {
	Host   string
	Port   int
	User   string
	Passwd string
	DBName string
}
