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

func ComputeAQI(pm25 uint16) (aqi uint16) {
    var pmmin, pmmax, aqimin, aqimax float64
    switch {
    case pm25 >= 0 && pm25 < 51:
        pmmin = 0
        pmmax = 15.5
        aqimin = 0
        aqimax = 51
    case pm25 >= 51 && pm25 < 101:
        pmmin = 15.5
        pmmax = 40.5
        aqimin = 51
        aqimax = 101
    case pm25 >= 101 && pm25 < 151:
        pmmin = 40.5
        pmmax = 65.5
        aqimin = 101
        aqimax = 151
    case pm25 >= 151 && pm25 < 201:
        pmmin = 65.5
        pmmax = 150.5
        aqimin = 151
        aqimax = 201
    default:
        return 300
    }
    aqi = uint16((float64(pm25) - pmmin)*(aqimax - aqimin)/(pmmax - pmmin)+aqimin)
    return
}
