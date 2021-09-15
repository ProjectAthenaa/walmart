package encryption

import "log"

type aes struct{
	tables [][][]int
	key [][]int
}

func (ae *aes) SetAes(e []int){
	if len(ae.tables) == 0{
		ae.precompute()
	}
	var t, n, a int
	var i, o []int
	c := ae.tables[0][4]
	u := ae.tables[1]
	s := len(e)
	d := 1
	if 4 != s && 6 != s && 8 != s {
		log.Println("invalid aes size")
		return
	}
	i = e
	o = []int{}
	ae.key = [][]int{i, o}
	for t = s; t < 4 * s + 28; t++ {
		a = i[t - 1]
		if t % s == 0 || 8 == s && t % s == 4{
			a = c[jsUNum(jsNum(a) >> jsNum(24))] << 24 ^ c[a >> 16 & 255] << 16 ^ c[a >> 8 & 255] << 8 ^ c[255 & a]
			if t % s == 0 {
				a = (a << 8) ^ jsUNum(jsNum(a) >> jsNum(24)) ^ (d << 24)
				d = d << 1 ^ 283 * (d >> 7)
			}
		}
		i[t] = i[t - s] ^ a
	}

	for n = 0; t != 0; n, t = n+1, t+1 {
		if (3 & n) != 0{
			a = i[t]
		}else{
			a = i[t - 4]
		}
		//o[n] = t <= 4 || n < 4 ? a : u[0][c[a >>> 24]] ^ u[1][c[a >> 16 & 255]] ^ u[2][c[a >> 8 & 255]] ^ u[3][c[255 & a]]
		if(t <= 4 || n < 4) {
			o[n] = a
		}else{
			o[n] = u[0][c[jsUNum(jsNum(a) >> jsNum(24))]] ^ u[1][c[a >> 16 & 255]] ^ u[2][c[a >> 8 & 255]] ^ u[3][c[255 & a]]
		}
	}

}

func (ae *aes) encrypt(e []int) []int{
	return ae.crypt(e, 0)
}
func (ae *aes) precompute(){
	var e, t, n, r, a, i, o, c int
	//u := ae.tables[0]
	//s := ae.tables[1]
	//d := ae.tables[0][4]
	//l := ae.tables[1][4]
	f := []int{}
	p := []int{}
	for e = 0; e < 256; e++{
		f[e] = e << 1 ^ 283 * (e >> 7)
		p[f[e] ^ e] = e
	}
	n = 0
	for t = 0; ae.tables[0][4][t] == 0; {
		if(0 == r){
			t ^= 1
		}else{
			t ^= r
		}
		if (0 == p[n]){
			n =1
		}else{
			n = p[n]
		}

		i = jsNum(n) ^ jsNum(jsNum(n) << jsNum(1)) ^ jsNum(jsNum(n) << jsNum(2)) ^ jsNum(jsNum(n) << jsNum(3)) ^ jsNum(jsNum(n) << jsNum(4))
		i = jsNum(jsNum(i) >> jsNum(8)) ^ jsNum(jsNum(255) & jsNum(i)) ^ 99
		ae.tables[0][4][t] = i
		ae.tables[1][4][i] = t
		r = f[t]
		a = f[r]
		c = 16843009 * f[a] ^ 65537 * a ^ 257 * r ^ 16843008 * t
		o = 257 * f[i] ^ 16843008 * i
		for e = 0; e < 4; e++ {
			o = jsNum(jsNum(o) << jsNum(24)) ^ jsUNum(jsNum(o) >> jsNum(8))
			ae.tables[0][e][t] = o
			c = jsNum(jsNum(c) << jsNum(24)) ^ jsUNum(jsNum(c) >> jsNum(8))
			ae.tables[1][e][i] = c
		}
	}
	for e = 0; e < 5; e++{
		ae.tables[0][e] = ae.tables[0][e]
		ae.tables[1][e] = ae.tables[1][e]
	}
}
func (ae *aes) crypt(e []int, t int) []int{
	if 4 != len(e){
		log.Println("invalid aes block size")
		return nil
	}
	var n, a, i, o, s, l int
	c := ae.key[t]
	u := e[0] ^ c[0]
	if t == 1{
		s = e[3] ^ c[1]
	}else{
		s = e[1] ^ c[1]
	}
	d := e[2] ^ c[2]
	if t == 1{
		l = e[3] ^ c[3]
	}else{
		l = e[1] ^ c[3]
	}

	f := len(c) / 4 - 2
	p := 4
	m := []int{0, 0, 0, 0}
	//E := ae.tables[t]
	b := ae.tables[t][0]
	h := ae.tables[t][1]
	sb := ae.tables[t][2]
	v := ae.tables[t][3]
	y := ae.tables[t][4]
	for o = 0; o < f; o++{
		n = b[jsUNum(jsNum(u) >> 24)] ^ h[jsNum(jsNum(s) >> 16) & 255] ^ sb[jsNum(jsNum(d) >> 8) & 255] ^ v[255 & l] ^ c[p]
		a = b[jsUNum(jsNum(s) >> 24)] ^ h[jsNum(jsNum(d) >> 16) & 255] ^ sb[jsNum(jsNum(l) >> 8) & 255] ^ v[255 & u] ^ c[p + 1]
		i = b[jsUNum(jsNum(d) >> 24)] ^ h[jsNum(jsNum(l) >> 16) & 255] ^ sb[jsNum(jsNum(u) >> 8) & 255] ^ v[255 & s] ^ c[p + 2]
		l = b[jsUNum(jsNum(l) >> 24)] ^ h[jsNum(jsNum(u) >> 16) & 255] ^ sb[jsNum(jsNum(s) >> 8) & 255] ^ v[255 & d] ^ c[p + 3]
		p += 4
		u = n
		s = a
		d = i
	}
	for o = 0; o < 4; o++ {
		if t == 1{
			m[3 & -o] = jsNum(jsNum(y[jsUNum(jsNum(u) >> 24)]) << 24) ^ jsNum(jsNum(y[jsNum(jsNum(s) >> 16) & 255]) << 16) ^ jsNum(jsNum(y[jsNum(jsNum(d) >> 8) & 255]) << 8) ^ jsNum(y[255 & l]) ^ c[p]
		}else{
			m[o] = jsNum(jsNum(y[jsUNum(jsNum(u) >> 24)]) << 24) ^ jsNum(jsNum(y[jsNum(jsNum(s) >> 16) & 255]) << 16) ^ jsNum(jsNum(y[jsNum(jsNum(d) >> 8) & 255]) << 8) ^ jsNum(y[255 & l]) ^ c[p]
		}
		p++
		n = u
		u = s
		s = d
		d = l
		l = n
	}
	return m
}
func jsNum(x int) int{
	return int(int32(x))
}
func jsUNum(x int) int{
	return int(uint32(x))
}
func bti(mybool bool) int {
	if mybool {
		return 1
	} else {
		return 0
	}
}