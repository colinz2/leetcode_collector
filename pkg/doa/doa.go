package doa

func MustOK(err error) {
	if err != nil {
		panic(err)
	}
}

func Assert(ok bool) {
	if !ok {
		panic("")
	}
}
