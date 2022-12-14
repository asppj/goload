package conf

type CommonConf struct {
	Mode    string `json:"mode" default:"dev" option:"dev,prod"`
	AppName string `json:"appName" default:"commonApp" valid:"required"`
}

type LocalConf struct {
	L       []Logger          `json:"l" desc:"日志" default:"0,1,2,3" option:"default"`
	Log     []Logger          `json:"log_Map2" desc:"日志" default:"0,1,2,3" option:"default"`
	Log2    []*Logger         `json:"log_Map3" desc:"日志" default:"0,1,2,3" option:"default"`
	WhiteIP []string          `json:"white_IP" desc:"白名单" default:"127.0.0.1,10.0.0.1,198.0.0.1" option:"0,1,2,3"`
	LogMap  map[string]Logger `json:"logMap" desc:"日志" default:"default,app,server" option:"default"`
	LogMap2 map[int]Logger    `json:"logMap2" desc:"日志" default:"1,2,3" option:"default"`
	cfgFile string            `default:"cfgFile" option:"" valid:"required"   desc:"配置文件地址"` // 不支持这种不可导出字段
	*CommonConf
	AppName string `json:"appName" default:"demoApp" desc:"app名字" valid:"option(testDemo|devDemo)"`
	Redis   *Redis `json:"redis" desc:"redis配置"`
}

type Redis struct {
	Host   string `json:"host" default:"127.0.0.1"`
	Port   int    `json:"port" default:"5678"`
	DB     int8   `json:"DB" default:"5"`
	Enable bool   `json:"enable" default:"true"`
}

type Logger struct {
	Name   string   `json:"name,omitempty" default:"appLog"`
	Level  string   `json:"level" default:"debug"`
	Output []string `json:"output" default:"stdio,file://" option:"stdio,file://**,es://**,vector://**" desc:"日志输出路径"`
}
