BUILD_DIR=build

.PHONY: all clean cmd api test

all: cmd api test

cmd:
	CXX=${CXX} CC=${CC} go build -o ${BUILD_DIR}/snowboy-detect example/detect.go
	CXX=${CXX} CC=${CC} go build -o ${BUILD_DIR}/snowboy-listen example/listen.go

api:
	CXX=${CXX} CC=${CC} go build -o ${BUILD_DIR}/snowboy-api example/api.go

test:
	cp $$GOPATH/src/github.com/Kitt-AI/snowboy/resources/* ${BUILD_DIR}
	CXX=${CXX} CC=${CC} go test -cover -race

clean:
	rm -rf ${BUILD_DIR}/*