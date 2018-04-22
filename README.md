# jpegbw

Converts JPEG(s) to B/W.

It actually supports JPEGs, PNGs, GIFs.

# usage

Q=90 R=0.2125 G=0.7154 B=0.0721 LO=5 HI=5 GA=1.41 ./jpegbw in.jpg

# expression parser

- You can use functions parser for example: `F="x1*x2+x3"`.
- Any math operations are allowed like `+, -, /, *, ^` etc.
- You can group expreccions using `( )`, for example `F="(x1+x2)*x3"`.
- `x1` will be replaced with greyscale value of current pixel, range is 0-1.
- `x2` will be replaced with current pixel's `x` position, range is 0-1.
- `x3` will be replaced with current pixel's `y` position, range is 0-1.
- You can also call functions from external C libraries.

# external functions

- To use external C function you must provide path to a dynamic library (`.so` on linux, `.dylib` on mac, `.dll` on windows etc).
- Library path example on mac: `LIB="/usr/lib/libm.dylib"`. Usually the `math lib` is what you need, linux: `LIB="/usr/lib/libm.so"`.
- Example usage: `time LIB="/lib/aarch64-linux-gnu/libm-2.24.so" F="sin(x1*3.14159)^2" jpegbw in.jpg`.

# multithreading

- Use `N=4` to specify to run using 4 threads, if no N is defined it will use Go runtime to get number of cores available.
