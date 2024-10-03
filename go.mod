module github.com/Mmx233/BackoffCli

go 1.23.1

replace github.com/Mmx233/BackoffCli/backoff => ./backoff

require (
	github.com/Microsoft/go-winio v0.6.2
	github.com/Mmx233/BackoffCli/backoff v0.0.0-20241003114231-8b0d8e75ed14
	github.com/alecthomas/kingpin/v2 v2.4.0
	github.com/antonfisher/nested-logrus-formatter v1.3.1
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/alecthomas/units v0.0.0-20240927000941-0f3dac36c52b // indirect
	github.com/xhit/go-str2duration/v2 v2.1.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
)
