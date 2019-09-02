package dockercompose

type STUNServer struct {
	*BaseComputer
}

func newSTUNServer(setup *Setup, name string, networks []*Network) *STUNServer {
	return &STUNServer{BaseComputer: newBaseComputer(setup, name, "gortc/gortcd", networks)}
}
