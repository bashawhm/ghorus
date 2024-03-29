package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	CMD_KA   = iota
	CMD_PING = iota
	CMD_QUIT = iota
	CMD_PLAY = iota
	CMD_CAPS = iota
)

type Packet struct {
	cmd  int
	data [8]int
}

func (p Packet) serialize() []byte {
	cmd := make([]byte, 4)
	binary.BigEndian.PutUint32(cmd, uint32(p.cmd))
	var data [8][]byte
	for i := 0; i < 8; i++ {
		data[i] = make([]byte, 4)
		binary.BigEndian.PutUint32(data[i], uint32(p.data[i]))
	}
	var final []byte
	final = append(final, cmd...)
	for i := 0; i < 8; i++ {
		final = append(final, data[i]...)
	}
	return final
}

func getClients(conn *net.UDPConn, pl PortList) []*net.UDPAddr {
	var clients []*net.UDPAddr
	var bAddr []*net.UDPAddr

	if len(pl) == 0 {
		port := "13676"
		addr, _ := net.ResolveUDPAddr("udp4", "255.255.255.255:"+port)
		bAddr = append(bAddr, addr)
	} else {
		for i := 0; i < len(pl); i++ {
			addr, _ := net.ResolveUDPAddr("udp4", "255.255.255.255:"+strconv.Itoa(pl[i]))
			bAddr = append(bAddr, addr)
		}
	}

	pkt := Packet{cmd: CMD_PING}
	for i := 0; i < len(bAddr); i++ {
		conn.WriteToUDP(pkt.serialize(), bAddr[i])
	}

	done := make(chan bool, 1)
	go func() {
		for {
			select {
			case <-done:
				break
			default:
				msg := make([]byte, 36)
				conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
				_, remote, err := conn.ReadFromUDP(msg)
				if err != nil {
					continue
				}
				clients = append(clients, remote)
			}
		}
	}()

	time.Sleep(time.Second)
	done <- true
	return clients
}

func sendToAll(broadcaster *net.UDPConn, clients []*net.UDPAddr, pkt Packet) {
	for i := 0; i < len(clients); i++ {
		broadcaster.WriteToUDP(pkt.serialize(), clients[i])
	}
}

func float32ToBytes(float float32) []byte {
	bits := math.Float32bits(float)
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint32(bytes, bits)
	return bytes
}

func intFromBytes(bytes []byte) int {
	num := binary.BigEndian.Uint32(bytes)
	return int(num)
}

func floatRaw(f float32) int {
	return intFromBytes(float32ToBytes(f))
}

func sendStreamToClient(broadcaster *net.UDPConn, client *net.UDPAddr, noteStream Stream, wg *sync.WaitGroup) {
	baseTime := time.Now()
	for i := 0; i < len(noteStream.Notes); i++ {
		dur := int(noteStream.Notes[i].Dur)
		durFranctional := int(int(noteStream.Notes[i].Dur*1000000) % 1000000)
		pitch := int(440.0 * math.Pow(2.0, float64(noteStream.Notes[i].Pitch-69)/12.0))
		ampl := floatRaw(noteStream.Notes[i].Ampl)

		for (time.Now().UnixNano() - baseTime.UnixNano()) < int64(noteStream.Notes[i].Time*1000000000) {
			time.Sleep(time.Duration(noteStream.Notes[i].Time*1000000000)*time.Nanosecond - time.Duration(time.Now().UnixNano()-baseTime.UnixNano()))
		}

		pkt := Packet{cmd: CMD_PLAY, data: [8]int{dur, durFranctional, pitch, ampl}}
		broadcaster.WriteToUDP(pkt.serialize(), client)

		time.Sleep(time.Duration(int64(noteStream.Notes[i].Dur*1000000000) - (time.Now().UnixNano() - baseTime.UnixNano()) - int64(noteStream.Notes[i].Time*1000000000)))
	}
	wg.Done()
}

func main() {
	var pl PortList
	flag.Var(&pl, "port", "The port to look for clients on")
	quit := flag.Bool("quit", false, "Sends quit packet to all clients")
	flag.Parse()

	if *quit {
		addr, _ := net.ResolveUDPAddr("udp4", "0.0.0.0:0")
		broadcaster, err := net.ListenUDP("udp4", addr)
		if err != nil {
			panic(err)
		}

		clients := getClients(broadcaster, pl)
		pkt := Packet{cmd: CMD_QUIT}
		sendToAll(broadcaster, clients, pkt)

	}

	if len(flag.Args()) <= 0 {
		fmt.Println("Usage: ./ghorus <iv_file>")
		return
	}

	xmlFile, err := os.Open(flag.Args()[0])
	if err != nil {
		panic(err)
	}
	defer xmlFile.Close()

	iv := LoadIv(xmlFile)

	noteStreams := iv.GetNoteStreams()
	fmt.Printf("Note Streams: %d\n", len(noteStreams))

	addr, _ := net.ResolveUDPAddr("udp4", "0.0.0.0:0")
	broadcaster, err := net.ListenUDP("udp4", addr)
	if err != nil {
		panic(err)
	}

	clients := getClients(broadcaster, pl)
	fmt.Printf("Clients: %d\n", len(clients))

	lastToPlay := LastNoteStreams(noteStreams)
	playTime := lastToPlay[0].Notes[len(lastToPlay[0].Notes)-1].Time + lastToPlay[0].Notes[len(lastToPlay[0].Notes)-1].Dur
	fmt.Printf("Playtime: %f\n", playTime)

	nsq := FirstNoteStreams(noteStreams)
	var wg sync.WaitGroup
	for i := 0; i < len(clients) && i < len(nsq); i++ {
		wg.Add(1)
		go sendStreamToClient(broadcaster, clients[i], nsq[i], &wg)
	}

	for i := 0; i < int(playTime); i++ {
		fmt.Printf("\x1b[G\x1b[K[%d]", i)
		time.Sleep(1 * time.Second)
	}

	wg.Wait()
	fmt.Println("Done!")
}
