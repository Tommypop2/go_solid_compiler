SET GOOS=js
SET GOARCH=wasm

go build -o ./build/compiler.wasm .

SET GOOS=windows
SET GOARCH=amd64
