package main

import (
	"fmt"
	"jpegbw"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl" // OR: github.com/go-gl/gl/v2.1/gl
	"github.com/go-gl/glfw/v3.2/glfw"
)

const (
	width  = 1000
	height = 1000

	vertexShaderSource = `
		#version 410
		in vec3 vp;
		void main() {
			gl_Position = vec4(vp, 1.0);
		}
	` + "\x00"

	fragmentShaderSource = `
		#version 410
		out vec4 frag_colour;
		void main() {
			frag_colour = vec4(1, 1, 1, 1.0);
		}
	` + "\x00"
)

func makeScene() []float32 {
	var fc jpegbw.FparCtx
	// LIB, NF
	lib := os.Getenv("LIB")
	if lib != "" {
		nf := 128
		nfs := os.Getenv("NF")
		if nfs != "" {
			v, err := strconv.Atoi(nfs)
			if err != nil {
				panic(err)
			}
			if v < 1 || v > 0xffff {
				panic(fmt.Errorf("NF must be from 1-65535 range"))
			}
			nf = v
		}
		ok := fc.Init(lib, uint(nf))
		if !ok {
			panic(fmt.Errorf("LIB init failed for: %s", lib))
		}
		defer func() { fc.Tidy() }()
	}
	err := fc.FparFunction(os.Args[1])
	if err != nil {
		panic(err)
	}
	err = fc.FparOK(1)
	if err != nil {
		panic(err)
	}

	// Config
	r0 := float32(-1.0)
	r1 := float32(1.0)
	ri := float32(0.05)
	i0 := float32(-1.0)
	i1 := float32(1.0)
	ii := float32(0.05)
	var (
		dataR [][]float32
		dataI [][]float32
	)
	x := 0
	r2 := r1 + ri
	i2 := i1 + ii
	for r := r0; r <= r2; r += ri {
		rowR := []float32{}
		rowI := []float32{}
		y := 0
		for i := i0; i <= i2; i += ii {
			z, err := fc.FparF([]complex128{complex(float64(r), float64(i))})
			if err != nil {
				panic(err)
			}
			rowR = append(rowR, float32(real(z)))
			rowI = append(rowI, float32(imag(z)))
			y++
		}
		dataR = append(dataR, rowR)
		dataI = append(dataI, rowI)
		x++
	}
	triangles := []float32{}
	x = 0
	for r := r0; r <= r1; r += ri {
		rt := r + ri
		xt := x + 1
		y := 0
		for i := i0; i <= i1; i += ii {
			it := i + ii
			yt := y + 1
			// triangles CCW
			triangles = append(
				triangles,
				[]float32{
					r, it, dataR[x][yt],
					r, i, dataR[x][y],
					rt, i, dataR[xt][y],
					r, it, dataR[x][yt],
					rt, i, dataR[xt][y],
					rt, it, dataR[xt][yt],
					r, it, dataI[x][yt],
					r, i, dataI[x][y],
					rt, i, dataI[xt][y],
					r, it, dataI[x][yt],
					rt, i, dataI[xt][y],
					rt, it, dataI[xt][yt],
				}...,
			)
			y++
		}
		x++
	}
	return triangles
}

func main() {
	runtime.LockOSThread()

	window := initGlfw()
	defer glfw.Terminate()
	program := initOpenGL()

	scene := makeScene()
	sceneLen := int32(len(scene) / 3)
	vao := makeVao(scene)
	for !window.ShouldClose() {
		draw(vao, sceneLen, window, program)
	}
}

func draw(vao uint32, nT int32, window *glfw.Window, program uint32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)

	gl.BindVertexArray(vao)
	gl.DrawArrays(gl.TRIANGLES, 0, nT)

	glfw.PollEvents()
	window.SwapBuffers()
}

// initGlfw initializes glfw and returns a Window to use.
func initGlfw() *glfw.Window {
	if len(os.Args) < 2 {
		panic(fmt.Errorf("please provide function definition"))
	}
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, os.Args[1], nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	return window
}

// initOpenGL initializes OpenGL and returns an intiialized program.
func initOpenGL() uint32 {
	if err := gl.Init(); err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	prog := gl.CreateProgram()
	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
	gl.LinkProgram(prog)
	return prog
}

// makeVao initializes and returns a vertex array from the points provided.
func makeVao(points []float32) uint32 {
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

	return vao
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}
