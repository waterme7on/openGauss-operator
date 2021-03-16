package crd

type OpenGaussList struct {
	Config OpenGaussListConfiguration
}

func (l *OpenGaussList) Items() []OpenGaussConfiguration {
	return l.Config.OgList
}

func (l *OpenGaussList) Size() int {
	return len(l.Config.OgList)
}

type OpenGauss struct {
	Config OpenGaussConfiguration
}
