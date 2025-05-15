package oneforall

type Result struct {
	Subdomains []Subdomain `json:"subdomains"`
}

type Subdomain struct {
	ID         int     `gorm:"column:id;primaryKey" json:"id"`
	Alive      int     `gorm:"column:alive" json:"alive"`
	Request    int     `gorm:"column:request" json:"request"`
	Resolve    int     `gorm:"column:resolve" json:"resolve"`
	URL        string  `gorm:"column:url" json:"url"`
	Subdomain  string  `gorm:"column:subdomain" json:"subdomain"`
	Port       int     `gorm:"column:port" json:"port"`
	Level      int     `gorm:"column:level" json:"level"`
	CNAME      string  `gorm:"column:cname" json:"cname"`
	IP         string  `gorm:"column:ip" json:"ip"`
	CDN        int     `gorm:"column:cdn" json:"cdn"`
	Status     int     `gorm:"column:status" json:"status"`
	Reason     string  `gorm:"column:reason" json:"reason"`
	Title      string  `gorm:"column:title" json:"title"`
	Banner     string  `gorm:"column:banner" json:"banner"`
	Header     string  `gorm:"column:header" json:"header"`
	History    string  `gorm:"column:history" json:"history"`
	Response   string  `gorm:"column:response" json:"response"`
	IPTimes    string  `gorm:"column:ip_times" json:"ip_times"`
	CNAMETimes string  `gorm:"column:cname_times" json:"cname_times"`
	TTL        string  `gorm:"column:ttl" json:"ttl"`
	CIDR       string  `gorm:"column:cidr" json:"cidr"`
	ASN        string  `gorm:"column:asn" json:"asn"`
	Org        string  `gorm:"column:org" json:"org"`
	Addr       string  `gorm:"column:addr" json:"addr"`
	ISP        string  `gorm:"column:isp" json:"isp"`
	Resolver   string  `gorm:"column:resolver" json:"resolver"`
	Module     string  `gorm:"column:module" json:"module"`
	Source     string  `gorm:"column:source" json:"source"`
	Elapse     float64 `gorm:"column:elapse" json:"elapse"`
	Find       int     `gorm:"column:find" json:"find"`
}
