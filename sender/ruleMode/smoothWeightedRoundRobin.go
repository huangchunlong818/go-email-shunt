package ruleMode

// Server represents a server with a weight and a current weight.
type Server struct {
	Id            int64
	Name          string
	Weight        int
	CurrentWeight int
}

// SmoothWeightedRoundRobin contains the list of servers and the total weight.
type SmoothWeightedRoundRobin struct {
	Servers     []*Server
	TotalWeight int
}

// NewSmoothWeightedRoundRobin creates a new SmoothWeightedRoundRobin with the given servers.
func NewSmoothWeightedRoundRobin(servers []*Server) *SmoothWeightedRoundRobin {
	swrr := &SmoothWeightedRoundRobin{Servers: servers}
	for _, server := range servers {
		swrr.TotalWeight += server.Weight
	}
	return swrr
}

// Next returns the next server according to the smooth weighted round robin algorithm.
func (swrr *SmoothWeightedRoundRobin) Next() *Server {
	var best *Server
	for _, server := range swrr.Servers {
		server.CurrentWeight += server.Weight
		if best == nil || server.CurrentWeight > best.CurrentWeight {
			best = server
		}
	}
	best.CurrentWeight -= swrr.TotalWeight
	return best
}
