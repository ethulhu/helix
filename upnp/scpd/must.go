package scpd

func Must(d Document, err error) Document {
	if err != nil {
		panic(err)
	}
	return d
}
