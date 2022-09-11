package main
import (
    "fmt"
    "io"
    "os"
    "encoding/binary"
    "bytes"
)

type Pmstruct struct {
    Framelen uint16 // :2
    Pm10_s, Pm25_s, Pm100_s uint16 // 2:4, 4:6, 6:8
    Pm10_e, Pm25_e, Pm100_e uint16
    P03, P05, P10, P25, P50, P100 uint16
    Unused uint16
    Checksum uint16
}
func main() {
    f, err := os.Open("/dev/cu.usbserial-A904QDZJ")
    //f, err := os.Open("/dev/ttyAMA0")
    if err != nil {
        panic("did not open")
    }
    fmt.Println("Finding sensor stream...")
    for {
        b := make([]byte, 1)
        _, err := f.Read(b)
        if err != nil {
            panic(err)
        }
        if b[0] == '\x42' {
            discard := make([]byte, 31)
            io.ReadAtLeast(f, discard, 31)
            break
        }
    }
    for {
        dat := make([]byte, 32)
        _, err := io.ReadAtLeast(f, dat, 32)
        if err != nil {
            panic(err)
        }

        checksum := uint16(0)
        for i := 0; i < 30; i++ {
            checksum += uint16(dat[i])
        }

        pmdata := Pmstruct{}
        err = binary.Read(bytes.NewBuffer(dat[2:]), binary.BigEndian, &pmdata)
        if err != nil {
            panic(err)
        }
        if checksum != pmdata.Checksum { //read_uint16(dat[30:]) {
            fmt.Printf("Checksum mismatch! %d, %d\n", checksum, pmdata.Checksum)
        } else {
            fmt.Printf("PM2.5: %d\t\tAQI2.5: %d\n", pmdata.Pm25_s, ComputeAQI(pmdata.Pm25_s))
        }

    }

}

func read_uint16(data []byte) (uint16) {
    return binary.BigEndian.Uint16(data)
}

func ComputeAQI(pmbyte uint16) (aqi uint16) {
    var pmmin, pmmax, aqimin, aqimax float64
    pm25 := float64(pmbyte)
    switch {
    case pm25 <= 12:
        pmmin = 0
        pmmax = 12
        aqimin = 0
        aqimax = 50
    case pm25 > 12 && pm25 <= 35.4:
        pmmin = 12
        pmmax = 35.4
        aqimin = 50
        aqimax = 100
    case pm25 > 35.4 && pm25 <= 55.4:
        pmmin = 35.4
        pmmax = 55.4
        aqimin = 100
        aqimax = 150
    case pm25 > 55.4 && pm25 <= 150.4:
        pmmin = 55.4
        pmmax = 150.4
        aqimin = 150
        aqimax = 200
    case pm25 > 150.4 && pm25 <= 250.4:
        pmmin = 150.4
        pmmax = 250.4
        aqimin = 200
        aqimax = 300
    case pm25 > 250.4 && pm25 <= 350.4:
        pmmin = 250.4
        pmmax = 350.4
        aqimin = 300
        aqimax = 400
    case pm25 > 350.4 && pm25 <= 500.4:
        pmmin = 350.4
        pmmax = 500.4
        aqimin = 400
        aqimax = 500
    case pm25 > 500.4:
        pmmin = 500.4
        pmmax = 650.4
        aqimin = 500
        aqimax = 600
    default:
        return 999
    }
    aqi = uint16((float64(pm25) - pmmin)*(aqimax - aqimin)/(pmmax - pmmin)+aqimin)
    return
}
