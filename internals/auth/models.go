package auth

// Users Table structure.
type Users struct {
	UserID     int    `json:"user_id" gorm:"primaryKey;autoIncrement;not null"`
	UserName   string `json:"user_name" gorm:"not null"`
	Password   string `json:"password" gorm:"not null"`
	MailID     string `json:"mail_id" gorm:"not null"`
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
