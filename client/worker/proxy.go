package worker

type Proxy struct {
}

var ProxyIntance *Proxy

func init() {
	ProxyIntance = &Proxy{}
}

func (a *Proxy) Start() {

}

func (p *Proxy) Write() {

}

func (p *Proxy) Read() {

}