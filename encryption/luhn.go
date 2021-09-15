package encryption

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

type luhnObj struct{
	base10 string
	base62 string
}

var n = luhnObj{
	base10 : "0123456789",
	base62 : "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
}

func (obj *luhnObj) Luhn(e string) int{
	t := len(e) - 1
	var inc int
	for inc = 0; t >= 0; {
		addint, err := strconv.Atoi(fmt.Sprintf("%c",e[t]))
		if err != nil{
			log.Println("couldnt parse int")
		}
		inc += addint
		t -= 2
	}
	for t = len(e) - 2; t >= 0;  {
		r, err  := strconv.Atoi(fmt.Sprintf("%c",e[t]))
		if err != nil{
			log.Println("couldnt parse int")
		}
		r *= 2
		if r < 10 {
			inc += r
		}else{
			inc += r -9
		}
		t -= 2
	}
	return inc % 10
}

func (obj *luhnObj) FixLuhn(e string, t, r int) string{
	a := obj.Luhn(e)
	if a < r{
		a += 10 - r
	}else {
		a -= r
	}

	if 0 != a{
		if (len(e) - t) % 2 != 0 {
			a = 10 - a
		}else {
			if a % 2 == 0{
				a = 5 - a / 2
			}else {
				a = (9 - a) / 2 + 5
			}
		}
		return e[:t] + fmt.Sprint(a) + e[t+1:]
	}else {
		return e
	}
}

func (obj *luhnObj) Distill(e string) string{
	var t strings.Builder
	for r := 0; r < len(e); r++{
		if strings.Index(n.base10, fmt.Sprintf("%c",e[r])) >= 0 {
			t.WriteString(fmt.Sprintf("%c",e[r]))
		}
	}
	return t.String()
}

func (obj *luhnObj) Reformat(e, t string) string{
	var r strings.Builder
	a := 0
	for i := 0; i < len(t); i++{
		if a < len(e) && strings.Index(n.base10, fmt.Sprintf("%c",t[i])) >= 0 {
			r.WriteString(fmt.Sprintf("%c",e[a]))
			a++
		}else{
			r.WriteString(fmt.Sprintf("%c",t[i]))
		}
	}
	return r.String()
}

func (obj *luhnObj) Integrity(e, t, n string) string{
	o := stringFromCharCode(0) + stringFromCharCode(len(t)) + t + stringFromCharCode(0) + stringFromCharCode(len(n)) + n
	c := HexToWords(e)
	c[3] ^= 1
	u := aes{}
	u.SetAes(c)
	s := compute(u, o)
	return WordToHex(s[0]) + WordToHex(s[1])
}