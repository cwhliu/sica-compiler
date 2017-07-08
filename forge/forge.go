package forge

type Forge struct {
	parser
}

func (f *Forge) Parse() error {
	if _, err := f.parse(); err != nil {
		return err
	}

	return nil
}
