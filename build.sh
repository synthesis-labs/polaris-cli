GOOS=linux GOARCH=386 go build -o bin/polaris-linux32 src/main.go
GOOS=linux GOARCH=amd64 go build -o bin/polaris-linux64 src/main.go
GOOS=darwin GOARCH=386 go build -o bin/polaris-osx32 src/main.go
GOOS=darwin GOARCH=amd64 go build -o bin/polaris-osx64 src/main.go
GOOS=windows GOARCH=386 go build -o bin/polaris-windows32 src/main.go
GOOS=windows GOARCH=amd64 go build -o bin/polaris-windows64 src/main.go
