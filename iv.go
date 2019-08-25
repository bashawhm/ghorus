package main

import (
	"encoding/xml"
	"io/ioutil"
	"os"
	"sort"
)

type Bpm struct {
	XMLName xml.Name `xml:"bpm"`
	Bpm     float32  `xml:"bpm,attr"`
	Ticks   int      `xml:"ticks,attr"`
	Time    float32  `xml:"time,attr"`
}

type Bpms struct {
	XMLName xml.Name `xml:"bpms"`
	Track   string   `xml:"track,attr"`
	Bpms    []Bpm    `xml:"bpm"`
}

type Meta struct {
	XMLName xml.Name `xml:"meta"`
	Bpms    Bpms     `xml:"bpms"`
	Args    []string `xml:"args"`
}

type Note struct {
	XMLName xml.Name `xml:"note"`
	Ampl    float32  `xml:"ampl,attr"`
	Dur     float32  `xml:"dur,attr"`
	Pitch   float32  `xml:"pitch,attr"`
	Time    float32  `xml:"time,attr"`
	Vel     int      `xml:"vel,attr"`
}

type Text struct {
	XMLName  xml.Name `xml:"text"`
	Text     string   `xml:"text,attr"`
	Time     float32  `xml:"time,attr"`
	TextType string   `xml:"type,attr"`
}

type Stream struct {
	XMLName    xml.Name `xml:"stream"`
	StreamType string   `xml:"type,attr"`
	Notes      []Note   `xml:"note"`
	Texts      []Text   `xml:"text"`
}

type Streams struct {
	XMLName xml.Name `xml:"streams"`
	Streams []Stream `xml:"stream"`
}

type Iv struct {
	XMLName xml.Name `xml:"iv"`
	Meta    Meta     `xml:"meta"`
	Streams Streams  `xml:"streams"`
}

func LoadIv(file *os.File) Iv {
	var iv Iv

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	err = xml.Unmarshal(bytes, &iv)
	if err != nil {
		panic(err)
	}

	return iv
}

func (iv Iv) GetNoteStreams() []Stream {
	var streams []Stream
	for i := 0; i < len(iv.Streams.Streams); i++ {
		if iv.Streams.Streams[i].StreamType == "ns" {
			streams = append(streams, iv.Streams.Streams[i])
		}
	}
	return streams
}

type byFirstTime []Stream

func (ns byFirstTime) Len() int {
	return len(ns)
}

func (ns byFirstTime) Swap(i, j int) {
	ns[i], ns[j] = ns[j], ns[i]
}

func (ns byFirstTime) Less(i, j int) bool {
	return ns[i].Notes[0].Time < ns[j].Notes[0].Time
}

func FirstNoteStreams(ns []Stream) []Stream {
	nsq := ns
	sort.Sort(byFirstTime(nsq))
	return nsq
}

type byLastTime []Stream

func (ns byLastTime) Len() int {
	return len(ns)
}

func (ns byLastTime) Swap(i, j int) {
	ns[i], ns[j] = ns[j], ns[i]
}

func (ns byLastTime) Less(i, j int) bool {
	return ns[i].Notes[len(ns[i].Notes)-1].Time+ns[i].Notes[len(ns[i].Notes)-1].Dur > ns[j].Notes[len(ns[j].Notes)-1].Time+ns[j].Notes[len(ns[j].Notes)-1].Dur
}

func LastNoteStreams(ns []Stream) []Stream {
	nsq := ns
	sort.Sort(byLastTime(nsq))
	return nsq
}
