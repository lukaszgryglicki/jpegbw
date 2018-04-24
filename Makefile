GO_BIN_FILES=cmd/jpegbw/jpegbw.go cmd/gengo/gengo.go
GO_LIB_FILES=fpar.go
GO_BIN_CMDS=jpegbw/cmd/jpegbw jpegbw/cmd/gengo
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
BINARIES=jpegbw gengo
STRIP=strip
C_LIBS=libjpegbw.so libbyname.so
C_ENV=
C_LINK=-lm -ldl
C_FILES=jpegbw.h jpegbw.c byname.h byname.c
C_FLAGS=-Wall -ansi -pedantic -shared -O3 -ffast-math -Wstrict-prototypes -Wmissing-prototypes -Wshadow -Wconversion
GCC=gcc

all: ${C_LIBS} check ${BINARIES}

gengo: cmd/gengo/gengo.go
	${GO_ENV} ${GO_BUILD} -o gengo cmd/gengo/gengo.go

jpegbw: cmd/jpegbw/jpegbw.go ${C_LIBS} ${GO_LIB_FILES}
	${GO_ENV} ${GO_BUILD} -o jpegbw cmd/jpegbw/jpegbw.go

libjpegbw.so: jpegbw.c jpegbw.h
	${C_ENV} ${GCC} ${C_FLAGS} -o libjpegbw.so jpegbw.c ${C_LINK}

libbyname.so: byname.c byname.h
	${C_ENV} ${GCC} ${C_FLAGS} -o libbyname.so byname.c ${C_LINK}

fmt: ${GO_BIN_FILES} ${GO_LIB_FILES}
	./for_each_go_file.sh "${GO_FMT}"
	./for_each_pgo_file.sh "${GO_FMT}"

lint: ${GO_BIN_FILES} ${GO_LIB_FILES}
	./for_each_go_file.sh "${GO_LINT}"

vet: ${GO_BIN_FILES} ${GO_LIB_FILES}
	./for_each_go_file.sh "${GO_VET}"

imports: ${GO_BIN_FILES} ${GO_LIB_FILES}
	./for_each_go_file.sh "${GO_IMPORTS}"

const: ${GO_BIN_FILES} ${GO_LIB_FILES}
	${GO_CONST} ./...

usedexports: ${GO_BIN_FILES} ${GO_LIB_FILES}
	${GO_USEDEXPORTS} ./...

errcheck: ${GO_BIN_FILES} ${C_LIBS}
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
