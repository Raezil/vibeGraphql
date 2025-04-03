package vibeGraphql

// ResolverFunc defines the function signature for all resolvers.
type ResolverFunc func(source interface{}, args map[string]interface{}) (interface{}, error)

// Global resolver registries.
var QueryResolvers = make(map[string]ResolverFunc)
var MutationResolvers = make(map[string]ResolverFunc)
var SubscriptionResolvers = make(map[string]ResolverFunc)

// Register functions.
func RegisterQueryResolver(field string, resolver ResolverFunc) {
	QueryResolvers[field] = resolver
}

func RegisterMutationResolver(field string, resolver ResolverFunc) {
	MutationResolvers[field] = resolver
}

func RegisterSubscriptionResolver(field string, resolver ResolverFunc) {
	SubscriptionResolvers[field] = resolver
}
