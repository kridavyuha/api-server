package profile

import "github.com/kridavyuha/api-server/internals/leagues"

type Profile struct {
	UserID     int    `json:"user_id"`
	UserName   string `json:"user_name"`
	MailId     string `json:"mail_id"`
	ProfilePic string `json:"profile_pic"`
	Credits    int    `json:"credits"`
	Rating     int    `json:"rating"`
}

type CompleteProfile struct {
	Profile Profile          `json:"profile"`
	Leagues []leagues.League `json:"leagues"`
}
