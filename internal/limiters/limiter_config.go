package limiters

type LimiterType int

const (
	FixedWindow   LimiterType = iota // 0
	SlidingWindow                    // 1
)

type LimiterConfig struct {
	requestLimit int64 // number of requests per minute allowed
	limiterType  LimiterType
	storage      string
}

func NewLimiterConfig(limit int64, limitType LimiterType) *LimiterConfig {
	return &LimiterConfig{
		requestLimit: limit,
		limiterType:  limitType,
	}
}
