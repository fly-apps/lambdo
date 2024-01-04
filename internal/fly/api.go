package fly

type Api struct {
	Token string
}

// NewApi returns an instance of the Fly API
func NewApi(token string) *Api {
	return &Api{
		Token: token,
	}
}
