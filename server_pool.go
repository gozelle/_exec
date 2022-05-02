package _exec

func NewServerPool() *ServerPool {
	return &ServerPool{}
}

type ServerPool struct {
	servers map[string]*Server
}

func (p *ServerPool) New(server *Server) *Server {
	if p.servers == nil {
		p.servers = map[string]*Server{}
	}
	v, ok := p.servers[server.Key]
	if ok {
		return v
	}
	p.servers[server.Key] = server
	return server
}

func (p *ServerPool) Close() {
	for _, server := range p.servers {
		server.Close()
	}
}
