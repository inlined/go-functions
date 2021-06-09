module firebase.com/deployer

go 1.15

require (
	acme.com/example v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/tools v0.1.2
)

replace acme.com/example => ../acme

replace firebase.com/functions => ../functions
