package encryption

import (
	"fmt"
	"math"
	"strconv"
)

type PIEStruct struct{
	L int
	E int
	K string
	Key_id string
	Phase int
}

func ProtectPANandCVV(e, t string, r int, PIE PIEStruct) []string{
	a := n.Distill(e)
	i := n.Distill(t)
	if len(a) < 13 || len(a) > 19 || len(i) > 4 || 1 == len(i) || 2 == len(i){
		return nil
	}
	c := a[:PIE.L] + a[len(a) - PIE.E:]
	if 1 == r{
		u := n.Luhn(a)
		s := substr(a, PIE.L + 1, len(a) - PIE.E)
		d := encrypt(s + i, c, PIE.K, 10)
		l := a[:PIE.L] + "0" + d[:len(d) - len(i)] + a[len(a) - PIE.E:]
		f := n.Reformat(n.FixLuhn(l, PIE.L, u), e)
		p := n.Reformat(d[len(d) - len(i):], t)
		return []string{f, p, n.Integrity(PIE.K, f, p)}
	}
	if 0 != n.Luhn(a){
		return nil
	}
	s := a[PIE.L + 1:(PIE.L + 1)+len(a) - PIE.E]
	var m int
	E := 23 - PIE.L - PIE.E
	b := s + i
	h := int(math.Floor( float64(float64(E) * math.Log(62) - 34 * math.Log(2)) / math.Log(10))) - len(b) - 1
	js := "11111111111111111111111111111"[:h] + strconv.Itoa(2 * len(i))
	d := "1" + encrypt(js + b, c, PIE.K, 10)
	parsed, err := strconv.ParseInt(PIE.Key_id, 16, 64)
	if err != nil{
		panic(err)
	}
	v := int(parsed)
	//y := new Array(len(d));
	var y []int
	for m = 0; m < len(d); m++{
		y[m], err = strconv.Atoi(fmt.Sprintf("%c",d[m]))
	}
	var g = convertRadix(y, len(d), 10, E, 62);
	g = bnMultiply(g, 62, 131072)
	g = bnMultiply(g, 62, 65536)
	g = bnAdd(g, 62, v)
	if(1 == PIE.Phase){
		bnAdd(g, 62, 4294967296);
	}
	var O = "";
	for m = 0; m < E; m++{
		O += string(n.base62[g[m]])
	}
	f := a[:PIE.L] + O[:E - 4] + a[len(a) - PIE.E:]
	p := O[E - 4:]
	return []string{f, p, n.Integrity(PIE.K, f, p)}
}