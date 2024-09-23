module github.com/Mmx233/BackoffCli

go 1.23.1

replace github.com/Mmx233/BackoffCli/backoff => ./backoff

require (
	github.com/Microsoft/go-winio v0.6.2
	github.com/Mmx233/BackoffCli/backoff v0.0.0-20240923083431-23a2ff3353db
	github.com/alecthomas/kingpin/v2 v2.4.0
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/alecthomas/units v0.0.0-20240626203959-61d1e3462e30 // indirect
	github.com/xhit/go-str2duration/v2 v2.1.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
)
