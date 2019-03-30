GO_BIN_FILES=cmd/jpegbw/jpegbw.go cmd/gengo/gengo.go cmd/cmap/cmap.go cmd/f/f.go cmd/jpeg/jpeg.go cmd/hist/hist.go
GO_LIB_FILES=fpar.go hist.go
GO_BIN_CMDS=github.com/lukaszgryglicki/jpegbw/cmd/jpegbw github.com/lukaszgryglicki/jpegbw/cmd/gengo github.com/lukaszgryglicki/jpegbw/cmd/cmap github.com/lukaszgryglicki/jpegbw/cmd/f github.com/lukaszgryglicki/jpegbw/cmd/jpeg github.com/lukaszgryglicki/jpegbw/cmd/hist
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
BINARIES=jpegbw gengo cmap f plot jpeg hist
STRIP=strip
C_LIBS=libjpegbw.so libbyname.so libtet.so
C_ENV=
C_LINK=-lm -ldl
C_FILES=jpegbw.h jpegbw.c byname.h byname.c tet.h tet.cpp util.h util.c
C_FLAGS=-Wall -ansi -pedantic -fPIC -shared -O3 -ffast-math -Wstrict-prototypes -Wmissing-prototypes -Wshadow -Wconversion
C_TEST=-Wall -ansi -pedantic -O3 -ffast-math -Wstrict-prototypes -Wmissing-prototypes -Wshadow -Wconversion
GCC=gcc

all: ${C_LIBS} check ${BINARIES}

gengo: cmd/gengo/gengo.go
	${GO_ENV} ${GO_BUILD} -o gengo cmd/gengo/gengo.go

hist: cmd/hist/hist.go ${GO_LIB_FILES}
	${GO_ENV} ${GO_BUILD} -o hist cmd/hist/hist.go

cmap: cmd/cmap/cmap.go ${C_LIBS} ${GO_LIB_FILES}
	${GO_ENV} ${GO_BUILD} -o cmap cmd/cmap/cmap.go

plot: cmd/plot/plot.go ${C_LIBS} ${GO_LIB_FILES}
	${GO_ENV} ${GO_BUILD} -o plot cmd/plot/plot.go

jpegbw: cmd/jpegbw/jpegbw.go ${C_LIBS} ${GO_LIB_FILES}
	${GO_ENV} ${GO_BUILD} -o jpegbw cmd/jpegbw/jpegbw.go

jpeg: cmd/jpeg/jpeg.go ${C_LIBS} ${GO_LIB_FILES}
	${GO_ENV} ${GO_BUILD} -o jpeg cmd/jpeg/jpeg.go

f: cmd/f/f.go ${C_LIBS} ${GO_LIB_FILES}
	${GO_ENV} ${GO_BUILD} -o f cmd/f/f.go

libjpegbw.so: jpegbw.c jpegbw.h util.h util.c
	${C_ENV} ${GCC} ${C_FLAGS} -o libjpegbw.so jpegbw.c util.c ${C_LINK}

libbyname.so: byname.c byname.h util.h util.c
	${C_ENV} ${GCC} ${C_FLAGS} -o libbyname.so byname.c util.c ${C_LINK}

libtet.so: tet.c tet.h util.h util.c
	${C_ENV} ${GCC} ${C_FLAGS} -o libtet.so tet.c util.c ${C_LINK}

fmt: ${GO_BIN_FILES} ${GO_LIB_FILES}
	./for_each_go_file.sh "${GO_FMT}"
	./for_each_pgo_file.sh "${GO_FMT}"

lint: ${GO_BIN_FILES} ${GO_LIB_FILES}
	./for_each_go_file.sh "${GO_LINT}"
	./for_each_pgo_file.sh "${GO_LINT}"

vet: ${GO_BIN_FILES} ${GO_LIB_FILES}
	./for_each_go_file.sh "${GO_VET}"

imports: ${GO_BIN_FILES} ${GO_LIB_FILES}
	./for_each_go_file.sh "${GO_IMPORTS}"
	./for_each_pgo_file.sh "${GO_IMPORTS}"

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
	rm -f ${BINARIES} ${C_LIBS} *.hist *.hint
