package oneforall

type Result struct {
	Subdomains []Subdomain `json:"subdomains"`
}

type Subdomain struct {
	ID         int     `gorm:"column:id;primaryKey"`
	Alive      int     `gorm:"column:alive"`
	Request    int     `gorm:"column:request"`
	Resolve    int     `gorm:"column:resolve"`
	URL        string  `gorm:"column:url"`
	Subdomain  string  `gorm:"column:subdomain"`
	Port       int     `gorm:"column:port"`
	Level      int     `gorm:"column:level"`
	CNAME      string  `gorm:"column:cname"`
	IP         string  `gorm:"column:ip"`
	CDN        int     `gorm:"column:cdn"`
	Status     int     `gorm:"column:status"`
	Reason     string  `gorm:"column:reason"`
	Title      string  `gorm:"column:title"`
	Banner     string  `gorm:"column:banner"`
	Header     string  `gorm:"column:header"`
	History    string  `gorm:"column:history"`
	Response   string  `gorm:"column:response"`
	IPTimes    string  `gorm:"column:ip_times"`
	CNAMETimes string  `gorm:"column:cname_times"`
	TTL        string  `gorm:"column:ttl"`
	CIDR       string  `gorm:"column:cidr"`
	ASN        string  `gorm:"column:asn"`
	Org        string  `gorm:"column:org"`
	Addr       string  `gorm:"column:addr"`
	ISP        string  `gorm:"column:isp"`
	Resolver   string  `gorm:"column:resolver"`
	Module     string  `gorm:"column:module"`
	Source     string  `gorm:"column:source"`
	Elapse     float64 `gorm:"column:elapse"`
	Find       int     `gorm:"column:find"`
}
