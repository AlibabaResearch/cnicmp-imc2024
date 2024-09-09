package utils

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"time"
)

// const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const letterBytes = "abcdef0123456789"

var mrand *rand.Rand

func init() {
	mrand = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func RandStringBytesRmndr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[mrand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func RandIPAddr() string {
	var ip string
	for i := 0; i < 4; i++ {
		ip += strconv.Itoa(mrand.Intn(256))
		if i < 3 {
			ip += "."
		}
	}
	return ip
}

// RandMACAddr generates a random unicast and locally administered MAC address.
func RandMACAddr() (net.HardwareAddr, error) {
	buf := make([]byte, 6)
	if _, err := mrand.Read(buf); err != nil {
		return nil, fmt.Errorf("Unable to retrieve 6 rnd bytes: %s", err)
	}

	// Set locally administered addresses bit and reset multicast bit
	buf[0] = (buf[0] | 0x02) & 0xfe

	return buf, nil
}
