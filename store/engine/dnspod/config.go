package dnspod

type Config struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	LoginToken string `json:"login_token"`
}
