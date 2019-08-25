package main

import "encoding/xml"

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

func (iv Iv) GetNoteStreams() []Stream {
	var streams []Stream
	for i := 0; i < len(iv.Streams.Streams); i++ {
		if iv.Streams.Streams[i].StreamType == "ns" {
			streams = append(streams, iv.Streams.Streams[i])
		}
	}
	return streams
}
