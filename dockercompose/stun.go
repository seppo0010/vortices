package dockercompose

type STUNServer struct {
	*BaseComputer
}

func newSTUNServer(name string, networks []*Network) *STUNServer {
	return &STUNServer{BaseComputer: newBaseComputer(name, "gortc/gortcd", networks)}
}
