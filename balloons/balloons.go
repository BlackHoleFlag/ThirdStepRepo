package main										//all thanks to :
													//Jack Mott - https://www.twitch.tv/jackmott42
													//Stefan Gustavson, 2003-2005 Contact: stefan.gustavson@liu.se

import (
	"os"
	"fmt"
	"time"
	"github.com/yawhoo/noise"
	"image/png"
	"github.com/veandco/go-sdl2/sdl"
)

const winWidth, winHeight int = 800, 600			//using this constant for future window

type texture struct {
	pos
	pixels []byte
	w,h, pitch int									//pitch = width * size of pixel
	scale float32
}

type rgba struct{
	r, g, b byte
}

type pos struct {
	x,y float32
}

func clear(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
}

func setPixel(x, y int, c rgba, pixels []byte) {
	index := (y*winWidth +x)*4
	if index < len(pixels)-4 && index >=0 {
		pixels[index] = c.r
		pixels[index+1] = c.g
		pixels[index+2] = c.b 
	}
}

func (tex *texture) drawScaled (scaleX, scaleY float32, pixels []byte){			//used for scaling balloons to mimic dept
	newWidth := int(float32(tex.w) * scaleX)
	newHeight := int(float32(tex.h) * scaleY)
	texW4 := tex.w *4			//<--pitch
	for y := 0; y < newHeight; y++ {
		fy := float32(y)/ float32(newHeight) * float32(tex.h -1)
		fyi := int(fy)
		screenY := int(fy * scaleY) + int(tex.y)
		screenIndex := screenY * winWidth * 4 + int(tex.x) * 4
		for x := 0; x < newWidth; x++ {
			fx := float32(x)/ float32(newWidth) * float32( tex.w -1)
			screenX :=int(fx * scaleX) + int(tex.x)
			if screenX >= 0 && screenX < winWidth && screenY >= 0 && screenY < winHeight {
			fxi4 := int(fx) *4

			pixels[screenIndex] = tex.pixels[fyi * texW4 + fxi4]	//red
			screenIndex++	
			pixels[screenIndex] = tex.pixels[fyi * texW4 + fxi4+1]	//green
			screenIndex++
			pixels[screenIndex] = tex.pixels[fyi * texW4 + fxi4+2]	//blue
			screenIndex++ 											//skip alfa
			screenIndex++
			}
		}
	}
}

func flerp(a,b,pct float32) float32 {
	return a+(b-a)* pct
}

func blerp(c00, c10, c01, c11, tx, ty float32) float32{
	return flerp(flerp(c00, c10, tx), flerp(c01, c11, tx), ty)
}

func (tex *texture) drawBilenearScaled(scaleX, scaleY float32, pixels []byte){			//used for scaling balloons to mimic dept
	newWidth := int(float32(tex.w) * scaleX)
	newHeight := int(float32(tex.h) * scaleY)
	texW4 := tex.w *4			//<--pitch
	for y := 0; y < newHeight; y++ {
		fy := float32(y)/ float32(newHeight) * float32(tex.h -1)
		fyi := int(fy)
		screenY := int(fy * scaleY) + int(tex.y)
		screenIndex := screenY * winWidth * 4 + int(tex.x) * 4
		ty := fy - float32(fyi)
		for x := 0; x < newWidth; x++ {
			fx := float32(x)/ float32(newWidth) * float32( tex.w -1)
			screenX :=int(fx * scaleX) + int(tex.x)
			if screenX >= 0 && screenX < winWidth && screenY >= 0 && screenY < winHeight {
			fxi := int(fx) 
				//indexes
			c00i := fyi * texW4 + fxi * 4
			c10i := fyi * texW4 + (fxi + 1) * 4
			c01i := (fyi + 1) * texW4 + fxi * 4
			c11i := (fyi + 1) * texW4 + (fxi + 1) * 4

			tx := fx - float32(fxi)					//percentage between top 2 pixels
				//colors
			for i := 0; i < 4; i++ {
				c00 := float32(tex.pixels[c00i + i])
				c10 := float32(tex.pixels[c10i + i])
				c01 := float32(tex.pixels[c01i + i])
				c11 := float32(tex.pixels[c11i + i])

				pixels[screenIndex] = byte(blerp(c00,c10,c01,c11,tx,ty))
				screenIndex++
			}
			
			}
		}
	}
}

func(tex *texture) draw(pixels []byte){
	for y := 0; y < tex.h; y++ {
		for x := 0; x < tex.w; x++ {
			screenY := y + int(tex.y)
			screenX := x + int(tex.x)
			if screenX >= 0 && screenX < winWidth && screenY >=0 && screenY < winHeight {
				texIndex := y *tex.pitch + x * 4					//reads data from texture
				screenIndex := screenY * winWidth * 4 + screenX * 4 //copy data on a screen

				pixels[screenIndex] = tex.pixels[texIndex]
				pixels[screenIndex+1] = tex.pixels[texIndex+1]
				pixels[screenIndex+2] = tex.pixels[texIndex+2]
				pixels[screenIndex+3] = tex.pixels[texIndex+3]
			}
		}
	}
}
func(tex *texture) drawAlpha(pixels []byte){							//more expensive than "draw", but work with pics with alfa value
	for y := 0; y < tex.h; y++ {
		screenY := y + int(tex.y)
		for x := 0; x < tex.w; x++ {
			
			screenX := x + int(tex.x)
			if screenX >= 0 && screenX < winWidth && screenY >=0 && screenY < winHeight {
				texIndex := y *tex.pitch + x * 4					//reads data from texture
				screenIndex := screenY * winWidth * 4 + screenX * 4 //copy data on a screen

				srcR := int(tex.pixels[texIndex])
				srcG := int(tex.pixels[texIndex+1])
				srcB := int(tex.pixels[texIndex+2])
				srcA := int(tex.pixels[texIndex+3])

				dstR := int(pixels[screenIndex])
				dstG := int(pixels[screenIndex+1])
				dstB := int(pixels[screenIndex+2])
				
				rstrR := (srcR * 255 + dstR *(255 - srcA))/255   //taking out alfa and scale it back down to 255
				rstrG := (srcG * 255 + dstG *(255 - srcA))/255
				rstrB := (srcB * 255 + dstB *(255 - srcA))/255

				pixels[screenIndex] = byte(rstrR)
				pixels[screenIndex+1] = byte(rstrG)
				pixels[screenIndex+2] = byte(rstrB)
				//pixels[screenIndex+3] = tex.pixels[texIndex+3]
			}
		}
	}
}


func loadBalloons()[]texture{

	balloonStrs := []string{"balloons/balloon_red.png", "balloons/balloon_green.png", "balloons/balloon_blue.png"}
	balloonTextures := make([]texture, len(balloonStrs))

	for i, bstr := range balloonStrs {	
	infile, err := os.Open(bstr)
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	img, err := png.Decode(infile)
	if err != nil {
		panic(err)
	}
	w := img.Bounds().Max.X
	h := img.Bounds().Max.Y

	balloonPixels := make([]byte, w*h*4)			//all ballon pixels would be extracted here
	bIndex := 0
	for y := 0; y < h; y++ {						
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x,y).RGBA()		//extract rgba colors from img (16byte val because have alfa)
			balloonPixels[bIndex] = byte(r/256)		
			bIndex++
			balloonPixels[bIndex] = byte(g/256)
			bIndex++
			balloonPixels[bIndex] = byte(b/256)
			bIndex++
			balloonPixels[bIndex] = byte(a/256)
			bIndex++
		}	
	}
	balloonTextures[i] = texture{pos{float32(i*60),float32(i*60)}, balloonPixels, w, h, w *4, float32(1+i)}
	}
	return balloonTextures
}

func lerp(b1, b2 byte, pct float32)byte {						//gives us linear interpolation between 2 bytes
	return byte(float32(b1) + pct *(float32(b2)- float32(b1))) 	//pct - percent
}

func colorLerp(c1, c2 rgba, pct float32) rgba {				//linear interpolation of each component (red,green,blue)
	return rgba{lerp(c1.r, c2.r, pct), lerp(c1.g, c2.g, pct), lerp(c1.b, c2.b, pct)}

}

func getGradient(c1,c2 rgba) []rgba{						//making gradient from 2 colors
	result:= make([]rgba, 256)
	for i := range result {
		pct := float32(i) / float32(255)
		result[i] = colorLerp(c1, c2, pct)
	}
	return result
}

func getDualGradient(c1, c2, c3, c4 rgba) []rgba{					//second gradient
	result:= make([]rgba, 256)
	for i := range result {
		pct := float32(i) / float32(255)
		if pct < 0.5 {
			result[i] = colorLerp(c1,c2, pct * float32(2))
		} else {
			result[i] = colorLerp(c3, c4, pct * float32(1.5)- float32(0.5))
		}
		
	}
	return result
}

func clamp(min, max, v int) int {							//this func make sure values in between range of what we looking for
	if v < min {
		v = min
	}else if v > max {
		v = max
	}
	return v
}

func rescaleAndDraw(noise []float32,min, max float32, gradient []rgba, w, h int) []byte {		//rescale noise --> turn in to byte --> set pixel array with that byte
	result := make([]byte, w * h * 4)
	scale := 255.0 / (max-min)								//to expand our values  from 0 to 255 using min-max
	offset := min * scale

	for i := range noise {									//going thru slice of noise
		noise[i] = noise[i] * scale - offset				//modifying and putting back
		c := gradient[clamp(0,255,int(noise[i]))]			//passing min/max and 
		p := i * 4
		result[p] = c.r									//set those byte in to pixels slice
		result[p + 1] = c.g
		result[p+ 2] = c.b
	}
	return result
}

func main(){

	err := sdl.Init(sdl.INIT_EVERYTHING)		//initializing sdl, wont work without it
	if err != nil {
		fmt.Println(err)
		return
	}

defer sdl.Quit() 								//closing properly after everything is done

window, err := sdl.CreateWindow("My name is what? my name is who? my name is window", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,		// creating window with sdl and our const for window
int32(winWidth), int32(winHeight), sdl.WINDOW_SHOWN)
if err != nil {
	fmt.Println(err)
	return
}
defer window.Destroy()							//all "defers" initializing from bottom to top

renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)  	// creating renderer
if err != nil {
	fmt.Println(err)							//always checking for mistakes
	return
}
defer renderer.Destroy()						//and cleaning up

tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(winWidth), int32(winHeight))//using renderer to create textures
if err != nil {
	fmt.Println(err)
	return
}
defer tex.Destroy()

cloudNoise, min, max := noise.MakeNoise(noise.FBM, 0.009, 0.4, 3, 3, winWidth, winHeight)			//noise for background
cloudGradient := getGradient(rgba{0,0,240}, rgba{250,250,250})										//grad for bg
cloudPixels := rescaleAndDraw(cloudNoise, min, max, cloudGradient, winWidth, winHeight) 
cloudTexture := texture{pos{0,0}, cloudPixels, winWidth, winHeight, winWidth *4, 1}

pixels := make([]byte, winWidth * winHeight * 4)  //we going to use "pixels"(slice of byte) to draw on a screen
balloonTextures := loadBalloons()						//balloon textures
dir := 1												//dirrection for balloon

for {
	frameStart := time.Now()
	for event := sdl.PollEvent(); event !=nil; event = sdl.PollEvent() {
		switch event.(type) {
		case *sdl.QuitEvent:
			return
		}
	}
	cloudTexture.draw(pixels)

	for _, tex := range balloonTextures {
		tex.drawBilenearScaled(tex.scale, tex.scale, pixels)
	}
	balloonTextures[1].x += float32(1*dir)
	if balloonTextures[1].x > 400 || balloonTextures[1].x < 0 {
		dir = dir * -1
	}
	
	tex.Update(nil, pixels, winWidth * 4)
	renderer.Copy (tex, nil, nil)
	renderer.Present()
	elapsedTime := float32(time.Since(frameStart).Seconds() * 1000)
	fmt.Println("ms per frame: ", elapsedTime)
	if elapsedTime < 5 {
		sdl.Delay(5 - uint32(elapsedTime))

	}
	sdl.Delay(16)
}
}