package main

import (
	"image/color"
	"time"

	"github.com/dropbox/godropbox/container/bitvector"
	"tinygo.org/x/drivers/flash"
	"tinygo.org/x/drivers/waveshare-epd/epd4in2"

	"machine"
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

	epdConfig epd4in2.Config
)

//var address = []int64{0x4000, 0x8000, 0xc000, 0x10000, 0x14000, 0x18000, 0x1c000, 0x20000}

func main() {

	for k := 0; k < 10; k++ {
		println(k)
		time.Sleep(time.Second)
	}

	blockDevice.Configure(&flash.DeviceConfig{
		Identifier: flash.DefaultDeviceIdentifier,
	})

	err := machine.SPI0.Configure(machine.SPIConfig{Frequency: 2000000})
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
	time.Sleep(5000 * time.Millisecond) //3000
	var num int64 = 1
	buf := make([]byte, 37)
	var addr int64 = 0
	for {
		display.ClearBuffer()
		display.ClearDisplay()
		display.WaitUntilIdle()
		time.Sleep(3000 * time.Millisecond) // needs min ~3 sec
		var i, j int16
		for i = 0; i < 300; i++ { //400 300 works 310 works for everyone except Dirac
			n := 0
			//addr := address[num] + int64(i*37)
			addr = num*0x4000 + int64(i*37)
			//blockDevice.ReadAt(buf, int64(addr))
			blockDevice.ReadAt(buf, addr)
			//fmt.Printf("About to write the %d\r\n row", i)
			bv := bitvector.NewBitVector(buf, 296)
			for j = 0; j < 296; j++ {
				if bv.Element(n) == 0 {
					display.SetPixel(i, j, black)
				}
				n++
			}
		}
		time.Sleep(1500 * time.Millisecond) // needs min ~1.5 sec
		display.Display()
		display.WaitUntilIdle()
		time.Sleep(10 * time.Second)
		num++
		if num == 9 { //9
			num = 1
		} else if num == 6 {
			num = 7
		}
	}
}
