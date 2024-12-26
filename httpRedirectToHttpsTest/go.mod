module github.com/bronze1man/httpRedirectToHttps/httpRedirectToHttpsTest

go 1.19

require (
	github.com/bronze1man/httpRedirectToHttps v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.33.0
)

replace github.com/bronze1man/httpRedirectToHttps => ../

require golang.org/x/text v0.21.0 // indirect
