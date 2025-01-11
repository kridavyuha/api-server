package auth

// Users Table structure.
type Users struct {
	UserID     int    `json:"user_id" gorm:"primaryKey;autoIncrement"`
	UserName   string `json:"user_name"`
	Password   string `json:"password"`
	MailID     string `json:"mail_id"`
	ProfilePic string `json:"profile_pic"`
}

type LoginRequestBody struct {
	Password string `json:"password"`
	UserName string `json:"user_name"`
}

type SignUpRequestBody struct {
	UserName string `json:"user_name"`
	MailID   string `json:"mail_id"`
	Password string `json:"password"`
}
