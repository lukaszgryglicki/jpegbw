# jpegbw

Converts JPEG(s) to B/W.

It actually supports JPEGs, PNGs, GIFs.

# usage

Q=90 R=0.2125 G=0.7154 B=0.0721 LO=5 HI=5 GA=1.41 ./jpegbw in.jpg

# expression parser

- You can use functions parser for example: `F="x1*x2+x3^x4"`.
- Any math operations are allowed like `+, -, /, *, ^` etc.
- Complex number are `_` separated, for example: `3_1` means 3+i, `_1` means 0+i, `_0` means 0+0i, `1_` or `1` means just 1+0i. `_` means 0+0i.
- You can group expreccions using `( )`, for example `F="(x1+x2)*x3"`.
- Functions can take 1, 2, 3 or 4 arguments.
- `x1` will be replaced with greyscale value of current pixel, range is 0-1.
- `x2` will be replaced with current pixel's `x` and `y` position `x+yi` , range is 0-1.
- `x3` will be replaced with current pixel's red and green colors `r+gi`, range is 0-1.
- `x4` will be replaced with current pixel's blue and alpha colors `b+ai`, range is 0-1.
- `x5` will be replaced with number indicating processing file number (scaled), and previous pixel's value `pn+prev*i` range is 0-1.
- You can also call functions from external C libraries.

# external functions

- To use external C function you must provide path to a dynamic library (`.so` on linux, `.dylib` on mac, `.dll` on windows etc).
- Library path example on mac: `LIB="/usr/lib/libm.dylib"`. Usually the `math lib` is what you need, linux: `LIB="/usr/lib/libm.so"`.
- Example usage: `time LIB="/lib/aarch64-linux-gnu/libm-2.24.so" F="sin(x1*3.14159)^2" jpegbw in.jpg`.
- You can use max up to 4-args functions, example: `R=0.25 G=0.6 B=0.15 LO=3 HI=3 LIB="/usr/lib/libm.dylib" F="fma(x2,x3,x1)" ./jpegbw in.png`.
- Other: `time R=0.25 G=0.6 B=0.15 LO=6 HI=6 LIB="/usr/lib/libm.dylib" F="((fma(x2,x3,x1)+fma(1-x2,x3,x1)+fma(x2,1-x3,x1)+fma(1-x2,1-x3,x1))/4)^2" ./jpegbw in.png`.
- Using local C library `libjepgbw.so`: `LIB="./libjepgbw.so" F="func(x1)" ./jpegbw in.png`.
- After `make install` just: `LIB="libjepgbw.so" F="func(x1)" jpegbw in.png`.
- Toon function: `LIB="libjepgbw.so" F="toon(x1,5)" jpegbw in.png`.
- Vingette function: `LIB="libjepgbw.so" F="vingette(x1, x2, x3)" jpegbw in.png`.
- Alpha function: `LIB="libjepgbw.so" F="alpha(x1, x2, x3, 1.4)" jpegbw in.png`.

# multithreading

- Use `N=4` to specify to run using 4 threads, if no N is defined it will use Go runtime to get number of cores available.

# combine 3 grayscale images into RGB image

- See `combine*.sh` scripts.

# other

- Use `O=".jpg:.png"` to overwite file name config. This will save JPG as PNG.

# build

- `go get github.com/andybons/gogif`
- `make && make install`.
- If you don't have tools required for `make check` do `sudo ./deps.sh`.
- If you still have any issues with additional check, compile binaries directly: `make jpegbw libjpegbw.so libbyname.so`.
- You can build debug binaries by using conditional compilation (`gengo` + `gen.sh` tools).
- Those tools are specially written to allow no additional overhead on non-debug binaries.

# install

- First build and then `sudo make install`.
- Package: `go get github.com/lukaszgryglicki/jpegbw`.

# development
- Edit `*.pgo` files instead of `*.go` files.
- Once done run `./gen.sh`.
- `*.go` files are generated from `*.pgo` files.

# cmap

Program to generate complex functions contour charts:

- Run `cmap` to see help.
- Example: `LIB="libjpegbw.so" X=1600 Y=1600 K=2 R0=-1 R1=4 I0=-4 I1=4 ./cmap complex_log.png "clog(x1)"`
```
(1600 x 1600) Real: [-1.000000,4.000000] Imag: [-4.000000,4.000000] Threads: 8
Values range: (-5.960527040805936-3.1390910953307114i) - (1.7328679513998633+3.139091095330712i), modulo range: 0.002795 - 6.116321, lines range: -5.960527 - 1.732868
Processed in: 22.294005s, MPPS: 0.110, 0
Real values from minimum to max are: red --> cyan/teal
Imag values from minimum to max are: blue --> yellow
Modulo values from minimum to max are: green --> pink
Re = 0 red almost white
Im = 0 blue almost white
Mod = 0 green almost white
Complex plane Re = 0, Im = 0 and modulo unit circle: white
Time: 22.294628s
```

# tetration

- There is a tetration library `libtet.so`.
- You can test it via: `clear; LIB="./libtet.so" ./cmap out.jpg 'tettest(1,_1,0.5,_1)'`
- You can use functions from the tetration library (see `tet.c`, `tet.h`):
  - `tet(z)` complex natural tetration of z (base e).
  - `ate(z)` complex natural abel-tetration (aka super logarithm) of z (base e).
  - `hexp(z, h)` - partial iterate of exp(z) function. For example h=1 --> exp(x), 2=2 --> exp(exp(x)), h=0 -> x, but h=0.5 half iterate exponential.
  - h in `hexp(z, h)` is complex, so you can make i-iterate of exp(z) via: `hexp(z, _1)`.
- `tettest` function calls C `exit` internally, it is supposed to be called once with 4 args to call various combinations of all functions mentioned above.
- All those functions can be used to generate contour charts of tetration, super log, half iterate exp etc.

# 'f' program

- Can be used to compute up to 4 args complex function
- You must provide 2 to 5 args: function def and 1-4 arguments
- Example: `LIB="./libtet.so" ./f 'csin(x1)*ccos(x2)*cpow(x3, x4)' 1 -2 _3 -_4`
