package profile

type Profile struct {
	UserID     int    `json:"user_id"`
	UserName   string `json:"user_name"`
	MailId     string `json:"mail_id"`
	ProfilePic string `json:"profile_pic"`
}
