package queue

func WithPrefetch(count int) func(*ConsumeOptions) {
	return func(o *ConsumeOptions) {
		o.PrefetchCount = &count
	}
}

func WithPriority(priority int) func(*ConsumeOptions) {
	return func(o *ConsumeOptions) {
		o.Priority = &priority
	}
}
