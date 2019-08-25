package main

import (
	"fmt"
	"strconv"
	"strings"
)

type PortList []int

func (pl *PortList) String() string {
	return fmt.Sprintf("%v", *pl)
}

func (pl *PortList) Set(elem string) error {
	splits := strings.Split(elem, ",")
	for i := 0; i < len(splits); i++ {
		num, _ := strconv.Atoi(splits[i])
		*pl = append(*pl, num)
	}
	return nil
}
