module github.com/system-pclub/GCatch/GCatch

go 1.20

require golang.org/x/tools v0.9.2

replace golang.org/x/tools v0.9.2 => ./tools

require github.com/aclements/go-z3 v0.0.0-20220809013456-4675d5f90ca5

require (
	golang.org/x/mod v0.10.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
)
