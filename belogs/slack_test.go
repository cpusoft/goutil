package belogs

// func TestSLACKWriter_WriteMsg(t *testing.T) {
// 	sc := `
// {
//   "webhookurl":"",
//   "level":7
// }
// `
// 	l := newSLACKWriter()
// 	err := l.Init(sc)
// 	if err != nil {
// 		Debug(err)
// 	}
//
// 	err = l.WriteMsg(&LogMsg{
// 		Level: 7,
// 		Msg: `{ "abs"`,
// 		When: time.Now(),
// 		FilePath: "main.go",
// 		LineNumber: 100,
// 		enableFullFilePath: true,
// 		enableFuncCallDepth: true,
// 	})
//
// 	if err != nil {
// 		Debug(err)
// 	}
//
// }
