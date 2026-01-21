package limiters

type RateLimiter interface {
	Allow(client_id string) (bool, error)
}
