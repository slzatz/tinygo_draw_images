package main

import (
	"fmt"
	"image/color"
	"io"
	"time"

	"github.com/dropbox/godropbox/container/bitvector"
	"tinygo.org/x/drivers/flash"
	"tinygo.org/x/drivers/waveshare-epd/epd4in2"
	"tinygo.org/x/tinyfs/littlefs"

	"machine"
	//"tinygo.org/x/tinydraw"
)

var (
	display epd4in2.Device
	black   = color.RGBA{1, 1, 1, 255}

	blockDevice = flash.NewQSPI(
		machine.QSPI_CS,
		machine.QSPI_SCK,
		machine.QSPI_DATA0,
		machine.QSPI_DATA1,
		machine.QSPI_DATA2,
		machine.QSPI_DATA3,
	)

	fs        = littlefs.New(blockDevice)
	epdConfig epd4in2.Config
)

func main() {

	for k := 0; k < 10; k++ {
		println(k)
		time.Sleep(time.Second)
	}

	flashConfig := &flash.DeviceConfig{Identifier: flash.DefaultDeviceIdentifier}
	if err := blockDevice.Configure(flashConfig); err != nil {
		for {
			time.Sleep(5 * time.Second)
			println("Config was not valid: "+err.Error(), "\r")
		}
	}

	// Configure littlefs with parameters for caches and wear levelling
	fs.Configure(&littlefs.Config{
		CacheSize:     512,
		LookaheadSize: 512,
		BlockCycles:   100,
	})

	if err := fs.Mount(); err != nil {
		println("Could not mount LittleFS filesystem: " + err.Error() + "\r\n")
	} else {
		println("Successfully mounted LittleFS filesystem.\r\n")
	}

	err := machine.SPI0.Configure(machine.SPIConfig{Frequency: 2000000}) //115200 worked
	if err != nil {
		println(err)
	}
	/*
		pins for epd display
		dc Data/Command: high for data, low for command
		cs Chip Select: low active
	*/
	busy := machine.D6
	rst := machine.D5
	dc := machine.D4
	cs := machine.D9
	//var config epd4in2.Config // smaller numbers for arduino nano 33 IoT and mkr wifi 1010
	epdConfig.Width = 400        // 200
	epdConfig.Height = 300       // 150
	epdConfig.LogicalWidth = 400 // 200
	epdConfig.Rotation = 0

	display = epd4in2.New(machine.SPI0, cs, dc, rst, busy)
	display.Configure(epdConfig)
	/*
		time.Sleep(3000 * time.Millisecond)
		display.ClearDisplay()
		display.ClearDisplay()
		display.ClearBuffer()
		time.Sleep(3000 * time.Millisecond) // needs min ~3 sec
	*/

	dir, err := fs.Open("/")
	if err != nil {
		println("Could not open littlefs directory")
		return
	}
	defer dir.Close()
	infos, err := dir.Readdir(0)
	if err != nil {
		println("Could not read littlefs directory")
		return
	}

	time.Sleep(3000 * time.Millisecond)
	num := 0
	for {
		fname := infos[num].Name()
		display.ClearDisplay()
		display.ClearBuffer()
		time.Sleep(3000 * time.Millisecond) // needs min ~3 sec
		f, err := fs.Open(fname)
		//f, err := fs.Open(info.Name())
		if err != nil {
			println("Could not open: " + err.Error())
			return
		}
		//defer f.Close()
		buf := make([]byte, 37)
		var i, j int16
		for i = 0; i < 400; i++ {
			n := 0
			_, err := f.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}
				println("Error reading file: " + err.Error())
			}
			fmt.Printf("About to write the %d\r\n row", i)
			bv := bitvector.NewBitVector(buf, 296)
			for j = 0; j < 296; j++ {
				if bv.Element(n) == 0 {
					display.SetPixel(i, j, black)
				}
				n++
			}
		}
		f.Close()
		time.Sleep(1500 * time.Millisecond) // needs min ~1.5 sec
		display.Display()
		time.Sleep(100 * time.Second)
		num++
		if num == len(infos) {
			num = 0
		}
	}
}
