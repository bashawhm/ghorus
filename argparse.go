package main

import (
	"fmt"
	"strconv"
)

type PortList []int

func (pl *PortList) String() string {
	return fmt.Sprintf("%v", *pl)
}

func (pl *PortList) Set(elem string) error {
	num, err := strconv.Atoi(elem)
	*pl = append(*pl, num)
	return err
}
