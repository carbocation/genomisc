package bulkprocess

type ExtractDicomOptions struct {
	IncludeOverlay bool
	WindowScaling  string
}

func OptIncludeOverlay() func(o *ExtractDicomOptions) {
	return func(o *ExtractDicomOptions) {
		o.IncludeOverlay = true
	}
}

func OptWindowScalingOfficial() func(o *ExtractDicomOptions) {
	return func(o *ExtractDicomOptions) {
		o.WindowScaling = "official"
	}
}
func OptWindowScalingPythonic() func(o *ExtractDicomOptions) {
	return func(o *ExtractDicomOptions) {
		o.WindowScaling = "pythonic"
	}
}
func OptWindowScalingRaw() func(o *ExtractDicomOptions) {
	return func(o *ExtractDicomOptions) {
		o.WindowScaling = "raw"
	}
}
