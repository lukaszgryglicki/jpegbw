package main // import "github.com/go-gl/example/gl41core-cube"

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/png"
	"jpegbw"
	"log"
	"math/cmplx"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const windowWidth = 1200
const windowHeight = 1200

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

var vertexShader = `
#version 330
uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;
in vec3 vert;
in vec2 vertTexCoord;
out vec2 fragTexCoord;
void main() {
    fragTexCoord = vertTexCoord;
    gl_Position = projection * camera * model * vec4(vert, 1);
}
` + "\x00"

var fragmentShader = `
#version 330
uniform sampler2D tex;
in vec2 fragTexCoord;
out vec4 outputColor;
void main() {
    outputColor = texture(tex, fragTexCoord);
}
` + "\x00"

func makeScene() []float32 {
	// Config
	r0 := float32(-1.5)
	r1 := float32(1.5)
	ri := float32(0.025)
	i0 := float32(-1.5)
	i1 := float32(1.5)
	ii := float32(0.025)

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

	var (
		dataR [][]float32
		dataI [][]float32
		dataM [][]float32
	)
	x := 0
	r2 := r1 + ri
	i2 := i1 + ii
	for r := r0; r <= r2; r += ri {
		rowR := []float32{}
		rowI := []float32{}
		rowM := []float32{}
		y := 0
		for i := i0; i <= i2; i += ii {
			z, err := fc.FparF([]complex128{complex(float64(r), float64(i))})
			if err != nil {
				panic(err)
			}
			rowR = append(rowR, float32(real(z)))
			rowI = append(rowI, float32(imag(z)))
			rowM = append(rowM, float32(cmplx.Abs(z)))
			y++
		}
		dataR = append(dataR, rowR)
		dataI = append(dataI, rowI)
		dataM = append(dataM, rowM)
		x++
	}
	triangles := []float32{}
	x = 0
	s := float32(0.0)
	si := (float32(1.0) / (r1 - r0)) * ri
	ti := (float32(1.0) / (i1 - i0)) * ii
	for r := r0; r <= r1; r += ri {
		rt := r + ri
		xt := x + 1
		st := s + si
		y := 0
		t := float32(0.0)
		for i := i0; i <= i1; i += ii {
			it := i + ii
			yt := y + 1
			tt := t + ti
			// triangles CCW
			triangles = append(
				triangles,
				[]float32{
					r, it, dataR[x][yt], s, tt,
					r, i, dataR[x][y], s, t,
					rt, i, dataR[xt][y], st, t,
					r, it, dataR[x][yt], s, tt,
					rt, i, dataR[xt][y], st, t,
					rt, it, dataR[xt][yt], st, tt,
				}...,
			)
			y++
			t += ti
		}
		x++
		s += si
	}
	x = 0
	s = float32(0.0)
	for r := r0; r <= r1; r += ri {
		rt := r + ri
		xt := x + 1
		st := s + si
		y := 0
		t := float32(0.0)
		for i := i0; i <= i1; i += ii {
			it := i + ii
			yt := y + 1
			tt := t + ti
			// triangles CCW
			triangles = append(
				triangles,
				[]float32{
					r, it, dataI[x][yt], s, tt,
					r, i, dataI[x][y], s, t,
					rt, i, dataI[xt][y], st, t,
					r, it, dataI[x][yt], s, tt,
					rt, i, dataI[xt][y], st, t,
					rt, it, dataI[xt][yt], st, tt,
				}...,
			)
			y++
			t += ti
		}
		x++
		s += si
	}
	x = 0
	s = float32(0.0)
	for r := r0; r <= r1; r += ri {
		rt := r + ri
		xt := x + 1
		st := s + si
		y := 0
		t := float32(0.0)
		for i := i0; i <= i1; i += ii {
			it := i + ii
			yt := y + 1
			tt := t + ti
			// triangles CCW
			triangles = append(
				triangles,
				[]float32{
					r, it, dataM[x][yt], s, tt,
					r, i, dataM[x][y], s, t,
					rt, i, dataM[xt][y], st, t,
					r, it, dataM[x][yt], s, tt,
					rt, i, dataM[xt][y], st, t,
					rt, it, dataM[xt][yt], st, tt,
				}...,
			)
			y++
			t += ti
		}
		x++
		s += si
	}
	/*
	  l := len(triangles)
	  for i:= 0; i < l; i += 5 {
	    fmt.Printf(
	      "(%g, %g, %g) x (%g, %g)\n",
	      triangles[i], triangles[i+1], triangles[i+2], triangles[i+3], triangles[i+4],
	    )
	  }
	*/
	return triangles
}

func main() {
	if len(os.Args) < 2 {
		panic(fmt.Errorf("please provide function definition"))
	}
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(windowWidth, windowHeight, os.Args[1], nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	// Initialize Glow
	if err := gl.Init(); err != nil {
		panic(err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)

	// Configure the vertex and fragment shaders
	program, err := newProgram(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}

	gl.UseProgram(program)

	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(windowWidth)/windowHeight, 0.05, 20.0)
	projectionUniform := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

	camera := mgl32.LookAtV(mgl32.Vec3{3, 3, 5}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	cameraUniform := gl.GetUniformLocation(program, gl.Str("camera\x00"))
	gl.UniformMatrix4fv(cameraUniform, 1, false, &camera[0])

	model := mgl32.Ident4()
	modelUniform := gl.GetUniformLocation(program, gl.Str("model\x00"))
	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

	textureUniform := gl.GetUniformLocation(program, gl.Str("tex\x00"))

	gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))

	// Load the texture
	textures, err := newTextures([3]string{"r.png", "i.png", "m.png"})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("textures %+v\n", textures)

	sceneAll := makeScene()
	sceneAllLen := len(sceneAll)
	sceneLen := sceneAllLen / 3
	var scene [3][]float32
	scene[0] = sceneAll[:sceneLen]
	scene[1] = sceneAll[sceneLen : 2*sceneLen]
	scene[2] = sceneAll[2*sceneLen:]
	fmt.Printf("sceneAllLen: %d, sceneLen %d\n", sceneAllLen, sceneLen)

	// Configure the vertex data
	var (
		vao [3]uint32
		vbo [3]uint32
	)
	for v := 0; v < 3; v++ {
		gl.GenVertexArrays(1, &vao[v])
		gl.BindVertexArray(vao[v])

		gl.GenBuffers(1, &vbo[v])
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo[v])
		gl.BufferData(gl.ARRAY_BUFFER, sceneLen*4, gl.Ptr(scene[v]), gl.STATIC_DRAW)

		vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vert\x00")))
		gl.EnableVertexAttribArray(vertAttrib)
		gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(0))

		texCoordAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vertTexCoord\x00")))
		gl.EnableVertexAttribArray(texCoordAttrib)
		gl.VertexAttribPointer(texCoordAttrib, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
	}
	fmt.Printf("vao: %+v, vbo: %+v\n", vao, vbo)

	// Configure global settings
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LEQUAL)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.ClearColor(1.0, 1.0, 1.0, 1.0)

	angle := 0.0
	previousTime := glfw.GetTime()

	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		//gl.Clear(gl.COLOR_BUFFER_BIT)

		// Update
		time := glfw.GetTime()
		elapsed := time - previousTime
		previousTime = time

		angle += elapsed
		model = mgl32.HomogRotate3D(float32(-90.0), mgl32.Vec3{1, 0, 0})
		model = mgl32.HomogRotate3D(float32(angle), mgl32.Vec3{0, 1, 0}).Mul4(model)

		// Render
		gl.UseProgram(program)
		gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

		for v := 0; v < 3; v++ {
			gl.Uniform1i(textureUniform, int32(v))
			gl.BindVertexArray(vao[v])
			switch v {
			case 0:
				gl.ActiveTexture(gl.TEXTURE0)
			case 1:
				gl.ActiveTexture(gl.TEXTURE1)
			default:
				gl.ActiveTexture(gl.TEXTURE2)
			}
			gl.BindTexture(gl.TEXTURE_2D, textures[v])

			gl.DrawArrays(gl.TRIANGLES, 0, int32(sceneLen))
		}

		// Maintenance
		window.SwapBuffers()
		glfw.PollEvents()
	}
}

func newProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
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

func newTextures(file [3]string) ([3]uint32, error) {
	textures := [3]uint32{0, 0, 0}
	for v := 0; v < 3; v++ {
		imgFile, err := os.Open(file[v])
		if err != nil {
			return textures, fmt.Errorf("texture %q not found on disk: %v", file, err)
		}
		defer func() { _ = imgFile.Close() }()
		img, _, err := image.Decode(imgFile)
		if err != nil {
			return textures, err
		}

		rgba := image.NewRGBA(img.Bounds())
		if rgba.Stride != rgba.Rect.Size().X*4 {
			return textures, fmt.Errorf("unsupported stride")
		}
		draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

		gl.GenTextures(1, &textures[v])
		switch v {
		case 0:
			gl.ActiveTexture(gl.TEXTURE0)
		case 1:
			gl.ActiveTexture(gl.TEXTURE1)
		default:
			gl.ActiveTexture(gl.TEXTURE2)
		}
		gl.BindTexture(gl.TEXTURE_2D, textures[v])
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		gl.TexImage2D(
			gl.TEXTURE_2D,
			0,
			gl.RGBA,
			int32(rgba.Rect.Size().X),
			int32(rgba.Rect.Size().Y),
			0,
			gl.RGBA,
			gl.UNSIGNED_BYTE,
			gl.Ptr(rgba.Pix),
		)
		fmt.Printf("%s texture size size %v\n", file[v], rgba.Rect.Size())
	}
	return textures, nil
}
