package backend

// GCInterface represents backend cache gc interface
type GCInterface interface {
	Init(options *FileConfig) (bool, error)
	Start()
	Stop()
}
