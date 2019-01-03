package teafile_test

type myThing struct {
	ts int64 `teafiles:"time"`
}

// func TestNewFile(t *testing.T) {
// 	f := teafiles.File{}

// 	if err := f.Write(myThing{}); err != nil {
// 		t.Error(err)
// 	}
// }
