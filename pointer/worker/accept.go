package worker

type Accept struct {
}

var AcceptIntance *Accept

func init() {
	AcceptIntance = &Accept{}
}

func (a *Accept) Start() {

}

func (p *Accept) Write() {

}

func (p *Accept) Read() {

}
