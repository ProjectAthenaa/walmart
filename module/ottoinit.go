package module

import "github.com/robertkrimen/otto"

//card, cvv, l, k, e
func initializeOtto(ottovm *otto.Otto){
	ottovm.Run(`function ProtectPANandCVV(card, cvv, PIELVAL, PIEKVAL, PIEEVAL){
    var a = distill(card),
        i = distill(cvv);
    if (a.length < 13 || a.length > 19 || i.length > 4 || 1 == i.length || 2 == i.length) return null;
    var c = a.substr(0, PIELVAL) + a.substring(a.length - PIEEVAL);
    var u = luhn(a),
        s = a.substring(PIELVAL + 1, a.length - PIEEVAL),
        d = encrypt(s + i, c, PIEKVAL, 10), // need to decode encrypt
        l = a.substr(0, PIELVAL) + '0' + d.substr(0, d.length - i.length) + a.substring(a.length - PIEEVAL),
        f = reformat(fixluhn(l, PIELVAL, u), card),
        p = reformat(d.substring(d.length - i.length), cvv);
    return [f, p, integrity(PIEKVAL, f, p)];
};

function distill(e){
    for (var t = '', r = 0; r < e.length; ++r) n.base10.indexOf(e.charAt(r)) >= 0 && (t += e.substr(r, 1));
    return t;
};

function luhn(e){
    for (var t = e.length - 1, n = 0; t >= 0; ) (n += parseInt(e.substr(t, 1), 10)), (t -= 2);
    for (t = e.length - 2; t >= 0; ) {
        var r = 2 * parseInt(e.substr(t, 1), 10);
        (n += r < 10 ? r : r - 9), (t -= 2);
    }
    return n % 10;
};

function fixluhn(e, t, r){
    var a = luhn(e);
    return (
        a < r ? (a += 10 - r) : (a -= r),
            0 != a
                ? ((a = (e.length - t) % 2 != 0 ? 10 - a : a % 2 == 0 ? 5 - a / 2 : (9 - a) / 2 + 5),
                e.substr(0, t) + a + e.substr(t + 1))
                : e
    );
};

function reformat(e, t){
    for (var r = '', a = 0, i = 0; i < t.length; ++i)
        a < e.length && n.base10.indexOf(t.charAt(i)) >= 0 ? ((r += e.substr(a, 1)), ++a) : (r += t.substr(i, 1));
    return r;
};

function integrity(e, t, n){
    var o =
            String.fromCharCode(0) +
            String.fromCharCode(t.length) +
            t +
            String.fromCharCode(0) +
            String.fromCharCode(n.length) +
            n,
        c = HexToWords(e);
    c[3] ^= 1;
    var u = new aes(c),
        s = compute(u, o);
    return WordToHex(s[0]) + WordToHex(s[1]);
};

var n = {
    base10: '0123456789',
    base62: '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz',
};

function aes(e){
    this._tables[0][0][0] || this._precompute();
    var t,
        n,
        a,
        i,
        o,
        c = this._tables[0][4],
        u = this._tables[1],
        s = e.length,
        d = 1;
    if (4 !== s && 6 !== s && 8 !== s) throw new r.exception.invalid('invalid aes key size');
    for (this._key = [(i = e.slice(0)), (o = [])], t = s; t < 4 * s + 28; t++)
        (a = i[t - 1]),
        (t % s == 0 || (8 === s && t % s == 4)) &&
        ((a = (c[a >>> 24] << 24) ^ (c[(a >> 16) & 255] << 16) ^ (c[(a >> 8) & 255] << 8) ^ c[255 & a]),
        t % s == 0 && ((a = (a << 8) ^ (a >>> 24) ^ (d << 24)), (d = (d << 1) ^ (283 * (d >> 7))))),
            (i[t] = i[t - s] ^ a);
    for (n = 0; t; n++, t--)
        (a = i[3 & n ? t : t - 4]),
            (o[n] =
                t <= 4 || n < 4
                    ? a
                    : u[0][c[a >>> 24]] ^ u[1][c[(a >> 16) & 255]] ^ u[2][c[(a >> 8) & 255]] ^ u[3][c[255 & a]]);
};

aes.prototype = {
    encrypt: function (e) {
        return this._crypt(e, 0);
    },
    decrypt: function (e) {
        return this._crypt(e, 1);
    },
    _tables: [
        [[], [], [], [], []],
        [[], [], [], [], []],
    ],
    _precompute: function () {
        var e,
            t,
            n,
            r,
            a,
            i,
            o,
            c,
            u = this._tables[0],
            s = this._tables[1],
            d = u[4],
            l = s[4],
            f = [],
            p = [];
        for (e = 0; e < 256; e++) p[(f[e] = (e << 1) ^ (283 * (e >> 7))) ^ e] = e;
        for (t = n = 0; !d[t]; t ^= 0 == r ? 1 : r, n = 0 == p[n] ? 1 : p[n])
            for (
                i = ((i = n ^ (n << 1) ^ (n << 2) ^ (n << 3) ^ (n << 4)) >> 8) ^ (255 & i) ^ 99,
                    d[t] = i,
                    l[i] = t,
                    c = (16843009 * f[(a = f[(r = f[t])])]) ^ (65537 * a) ^ (257 * r) ^ (16843008 * t),
                    o = (257 * f[i]) ^ (16843008 * i),
                    e = 0;
                e < 4;
                e++
            )
                (u[e][t] = o = (o << 24) ^ (o >>> 8)), (s[e][i] = c = (c << 24) ^ (c >>> 8));
        for (e = 0; e < 5; e++) (u[e] = u[e].slice(0)), (s[e] = s[e].slice(0));
    },
    _crypt: function (e, t) {
        if (4 !== e.length) throw new r.exception.invalid('invalid aes block size');
        var n,
            a,
            i,
            o,
            c = this._key[t],
            u = e[0] ^ c[0],
            s = e[t ? 3 : 1] ^ c[1],
            d = e[2] ^ c[2],
            l = e[t ? 1 : 3] ^ c[3],
            f = c.length / 4 - 2,
            p = 4,
            m = [0, 0, 0, 0],
            E = this._tables[t],
            b = E[0],
            h = E[1],
            _ = E[2],
            v = E[3],
            y = E[4];
        for (o = 0; o < f; o++)
            (n = b[u >>> 24] ^ h[(s >> 16) & 255] ^ _[(d >> 8) & 255] ^ v[255 & l] ^ c[p]),
                (a = b[s >>> 24] ^ h[(d >> 16) & 255] ^ _[(l >> 8) & 255] ^ v[255 & u] ^ c[p + 1]),
                (i = b[d >>> 24] ^ h[(l >> 16) & 255] ^ _[(u >> 8) & 255] ^ v[255 & s] ^ c[p + 2]),
                (l = b[l >>> 24] ^ h[(u >> 16) & 255] ^ _[(s >> 8) & 255] ^ v[255 & d] ^ c[p + 3]),
                (p += 4),
                (u = n),
                (s = a),
                (d = i);
        for (o = 0; o < 4; o++)
            (m[t ? 3 & -o : o] =
                (y[u >>> 24] << 24) ^ (y[(s >> 16) & 255] << 16) ^ (y[(d >> 8) & 255] << 8) ^ y[255 & l] ^ c[p++]),
                (n = u),
                (u = s),
                (s = d),
                (d = l),
                (l = n);
        return m;
    },
};

function HexToKey(e){
    return new aes(HexToWords(e));
};
function HexToWords(e){
    var t = new Array(4);
    if (32 != e.length) return null;
    for (var n = 0; n < 4; n++) t[n] = parseInt(e.substr(8 * n, 8), 16);
    return t;
};

var Hex = '0123456789abcdef';

function WordToHex(e){
    for (var t = 32, n = ''; t > 0; ) (t -= 4), (n += Hex.substr((e >>> t) & 15, 1));
    return n;
};

function MSBnotZero(e){
    return 2147483647 != (2147483647 | e);
};

function leftShift(e){
    (e[0] = ((2147483647 & e[0]) << 1) | (e[1] >>> 31)),
        (e[1] = ((2147483647 & e[1]) << 1) | (e[2] >>> 31)),
        (e[2] = ((2147483647 & e[2]) << 1) | (e[3] >>> 31)),
        (e[3] = (2147483647 & e[3]) << 1);
};

var var_Rb = 135;

function compute(e, t){
    var n = [0, 0, 0, 0],
        r = e.encrypt(n),
        a = r[0];
    leftShift(r), MSBnotZero(a) && (r[3] ^= var_Rb);
    for (var o = 0; o < t.length; )
        (n[(o >> 2) & 3] ^= (255 & t.charCodeAt(o)) << (8 * (3 - (3 & o)))),
        0 == (15 & ++o) && o < t.length && (n = e.encrypt(n));
    return (
        (0 != o && 0 == (15 & o)) ||
        ((a = r[0]), leftShift(r), MSBnotZero(a) && (r[3] ^= var_Rb), (n[(o >> 2) & 3] ^= 128 << (8 * (3 - (3 & o))))),
            (n[0] ^= r[0]),
            (n[1] ^= r[1]),
            (n[2] ^= r[2]),
            (n[3] ^= r[3]),
            e.encrypt(n)
    );
};

var alphabet = [
    '0',
    '1',
    '2',
    '3',
    '4',
    '5',
    '6',
    '7',
    '8',
    '9',
    'A',
    'B',
    'C',
    'D',
    'E',
    'F',
    'G',
    'H',
    'I',
    'J',
    'K',
    'L',
    'M',
    'N',
    'O',
    'P',
    'Q',
    'R',
    'S',
    'T',
    'U',
    'V',
    'W',
    'X',
    'Y',
    'Z',
];
function precompF(e, t, n, r){
    var a = new Array(4),
        i = n.length;
    return (
        (a[0] = 16908544 | ((r >> 16) & 255)),
            (a[1] = (((r >> 8) & 255) << 24) | ((255 & r) << 16) | 2560 | (255 & Math.floor(t / 2))),
            (a[2] = t),
            (a[3] = i),
            e.encrypt(a)
    );
};

function precompb(e, t){
    for (var n = Math.ceil(t / 2), r = 0, a = 1; n > 0; ) --n, (a *= e) >= 256 && ((a /= 256), ++r);
    return a > 1 && ++r, r;
};

function bnMultiply(e, t, n){
    var r,
        a = 0;
    for (r = e.length - 1; r >= 0; --r) {
        var i = e[r] * n + a;
        (e[r] = i % t), (a = (i - e[r]) / t);
    }
};

function bnAdd(e, t, n){
    for (var r = e.length - 1, a = n; r >= 0 && a > 0; ) {
        var i = e[r] + a;
        (e[r] = i % t), (a = (i - e[r]) / t), --r;
    }
};

function convertRadix(e, t, n, r, a){
    var i,
        c = new Array(r);
    for (i = 0; i < r; ++i) c[i] = 0;
    for (var u = 0; u < t; ++u) bnMultiply(c, a, n), bnAdd(c, a, e[u]);
    return c;
};

function cbcmacq(e, t, n, r){
    for (var a = new Array(4), i = 0; i < 4; ++i) a[i] = e[i];
    for (var o = 0; 4 * o < n; ) {
        for (i = 0; i < 4; ++i)
            a[i] =
                a[i] ^ ((t[4 * (o + i)] << 24) | (t[4 * (o + i) + 1] << 16) | (t[4 * (o + i) + 2] << 8) | t[4 * (o + i) + 3]);
        (a = r.encrypt(a)), (o += 4);
    }
    return a;
};

function F(e, t, n, r, a, i, c, u, s){
    var d = Math.ceil(s / 4) + 1,
        l = (n.length + s + 1) & 15;
    l > 0 && (l = 16 - l);
    var f,
        p = new Array(n.length + l + s + 1);
    for (f = 0; f < n.length; f++) p[f] = n.charCodeAt(f);
    for (; f < l + n.length; f++) p[f] = 0;
    p[p.length - s - 1] = t;
    for (var m = convertRadix(r, a, u, s, 256), E = 0; E < s; E++) p[p.length - s + E] = m[E];
    var b,
        h = cbcmacq(c, p, p.length, e),
        _ = h,
        v = new Array(2 * d);
    for (f = 0; f < d; ++f)
        f > 0 &&
        0 == (3 & f) &&
        ((b = (f >> 2) & 255),
            (b |= (b << 8) | (b << 16) | (b << 24)),
            (_ = e.encrypt([
                h[0] ^ b, // fix
                h[1] ^ b,
                h[2] ^ b,
                h[3] ^ b,
            ]))),
            (v[2 * f] = _[3 & f] >>> 16),
            (v[2 * f + 1] = 65535 & _[3 & f]);
    return convertRadix(v, 2 * d, 65536, i, u);
};

function DigitToVal(e, t, n){
    var r = new Array(t);
    if (256 == n) {
        for (var a = 0; a < t; a++) r[a] = e.charCodeAt(a);
        return r;
    }
    for (var i = 0; i < t; i++) {
        var o = parseInt(e.charAt(i), n);
        if (NaN == o || !(o < n)) return '';
        r[i] = o;
    }
    return r;
};

function ValToDigit(e, t){
    var n,
        r = '';
    if (256 == t) for (n = 0; n < e.length; n++) r += String.fromCharCode(e[n]);
    else for (n = 0; n < e.length; n++) r += alphabet[e[n]];
    return r;
};

function encryptWithCipher(e, t, n, r){
    var a = e.length,
        i = Math.floor(a / 2),
        c = precompF(n, a, t, r),
        u = precompb(r, a),
        s = DigitToVal(e, i, r),
        d = DigitToVal(e.substr(i), a - i, r);
    if ('' == s || '' == d) return '';
    for (var l = 0; l < 5; l++) {
        var f,
            p = F(n, 2 * l, t, d, d.length, s.length, c, r, u);
        f = 0;
        for (var m = s.length - 1; m >= 0; --m) {
            (E = s[m] + p[m] + f) < r ? ((s[m] = E), (f = 0)) : ((s[m] = E - r), (f = 1));
        }
        p = F(n, 2 * l + 1, t, s, s.length, d.length, c, r, u);
        f = 0;
        for (m = d.length - 1; m >= 0; --m) {
            var E;
            (E = d[m] + p[m] + f) < r ? ((d[m] = E), (f = 0)) : ((d[m] = E - r), (f = 1));
        }
    }
    return ValToDigit(s, r) + ValToDigit(d, r);
};
function encrypt(e, t, n, r){
    var i = HexToKey(n);
    return null == i ? '' : encryptWithCipher(e, t, i, r);
};`)
}
