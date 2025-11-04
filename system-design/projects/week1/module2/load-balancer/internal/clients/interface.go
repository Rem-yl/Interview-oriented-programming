package clients

type HttpClient interface {
	Get(url string) ([]byte, error)
	Post(url string, body []byte) ([]byte, error)
}
