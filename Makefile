all:
	cd src && GOOS=windows go build -o ../build/hash-index-windows.exe .
	cd src && GOOS=linux go build -o ../build/hash-index-linux .
