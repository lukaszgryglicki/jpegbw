GO_BIN_FILES=cmd/jpegbw/jpegbw.go
GO_BIN_CMDS=jpegbw/cmd/jpegbw
GO_ENV=CGO_ENABLED=1
GO_BUILD=go build -ldflags '-s -w'
#GO_BUILD=go build -ldflags '-s -w' -race
GO_INSTALL=go install -ldflags '-s'
GO_FMT=gofmt -s -w
GO_LINT=golint -set_exit_status
GO_VET=go vet
GO_CONST=goconst
GO_IMPORTS=goimports -w
GO_USEDEXPORTS=usedexports
GO_ERRCHECK=errcheck -asserts -ignore '[FS]?[Pp]rint*'
BINARIES=jpegbw
STRIP=strip
C_LIBS=jpegbw.so
C_ENV=
C_LINK=-lm
C_FILES=jpegbw.h jpegbw.c
GCC=gcc

all: check ${BINARIES} ${C_LIBS}

jpegbw: cmd/jpegbw/jpegbw.go
	${GO_ENV} ${GO_BUILD} -o jpegbw cmd/jpegbw/jpegbw.go

jpegbw.so: ${C_FILES}
	${C_ENV} ${GCC} -shared -O3 -ffast-math -o jpegbw.so jpegbw.c ${C_LINK}

fmt: ${GO_BIN_FILES}
	./for_each_go_file.sh "${GO_FMT}"

lint: ${GO_BIN_FILES}
	./for_each_go_file.sh "${GO_LINT}"

vet: ${GO_BIN_FILES}
	./for_each_go_file.sh "${GO_VET}"

imports: ${GO_BIN_FILES}
	./for_each_go_file.sh "${GO_IMPORTS}"

const: ${GO_BIN_FILES}
	${GO_CONST} ./...

usedexports: ${GO_BIN_FILES}
	${GO_USEDEXPORTS} ./...

errcheck: ${GO_BIN_FILES}
	${GO_ERRCHECK} ./...

check: fmt lint imports vet const usedexports errcheck

install: ${BINARIES} ${C_LIBS}
	${GO_INSTALL} ${GO_BIN_CMDS}
	cp ${C_LIBS} /usr/local/lib

strip: ${BINARIES} ${C_LIBS}
	${STRIP} ${BINARIES}
	${STRIP} ${C_LIBS}

clean:
	rm -f ${BINARIES} ${C_LIBS}
