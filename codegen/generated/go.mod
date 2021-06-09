module main

go 1.15

replace acme.com/example => ../acme

require acme.com/example v0.0.0-00010101000000-000000000000

replace firebase.com/functions => ../functions
