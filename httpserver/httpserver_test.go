package httpserver

/*
func TestListenAndServe(t *testing.T) {

	router, err := rest.MakeRouter(
		rest.Get("/hello", hello),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		httpserver.ListenAndServe(":8080", &router)
	}()
	go func() {
		httpserver.ListenAndServeTLS(":8081", `E:\Go\common-util\conf\server.crt`, `E:\Go\common-util\conf\server.key`, &router)
	}()
	select {}
}

//
func hello(w rest.ResponseWriter, req *rest.Request) {
	w.WriteJson("hello world")
}
*/
