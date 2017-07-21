package main

import (
    "log"
    "os"
    "encoding/binary"

    "github.com/spaolacci/murmur3"
)

const (
    LongShift = 30
    LongBase = uint32(1) << LongShift
    LongMash = uint32(LongBase - 1)
    LongDecimalShift = 9
    LongDecimalBase = uint32(1000000000)
)

func Sum128toString(data string) string {

    var key string = "Not implemented yet"
    var v []uint64

    h1, h2 := murmur3.Sum128([]byte(data))

    b1 := make([]byte, 8)
    binary.LittleEndian.PutUint64(b1, h1)

    b2 := make([]byte, 8)
    binary.LittleEndian.PutUint64(b2, h2)

    b := append(b1, b2...)

    var pstartbyte []byte = b
    var pendbyte []byte = b[15:]
    var numsignificantbytes int
    var ndigits int
    var idigit int = 0
    var is_signed = pendbyte[0] >= 0x80

    {
        var p []byte = pendbyte
        var i int
        var insignficant byte
        if is_signed {
            insignficant = 0xff
        } else{
            insignficant = 0x00
        }

        for i := 0; i < 16; i++  {
            if p[i] != insignficant {
                break
            }
        }
        numsignificantbytes = 16 - i

        if (is_signed && numsignificantbytes < 16) {
            numsignificantbytes++
        }
    }
    ndigits = (numsignificantbytes * 8 + 30 - 1) / 30
    v = make([]uint64, ndigits)

    {
        var i int;
        var carry uint64 = 1
        var accum uint64 = 0
        var accumbits uint = 0
        var p []byte = pstartbyte

        for i = 0; i < numsignificantbytes; i++ {
            var thisbyte uint64 = uint64(p[i])

            if (is_signed) {
                thisbyte = (0xff ^ thisbyte) + carry
                carry = thisbyte >> 8
                thisbyte &= 0xff
            }

            accum |= uint64(thisbyte << accumbits)

            accumbits += 8;
            if (accumbits >= 30) {
                v[idigit] = uint64(uint32(accum) & LongMash)
                idigit++
                accum >>= 30
                accumbits -= 30
            }
        }
        if accumbits != 0 {
            v[idigit] = uint64(accum)
            idigit++
        }
    }

    {
        var size, strlen, size_a, i, j int
        var rem, tenpow uint32
        var pout []uint64
        var scratch []uint64 = make([]uint64, ndigits)
        var p []byte
        var negative int = 0
        var addL int = 1

        pout = scratch

        size = 0
        size_a = ndigits
        
        for i = size_a-1; i >= 0; i-- {

            var hi uint32 = uint32(v[i])

            for j := 0; j < size; j++ {
                var z uint64 = uint64(pout[j]) << LongShift | uint64(hi)
                hi = uint32(z / uint64(LongDecimalBase))
                pout[j] = uint64(uint32(z) - (hi * LongDecimalBase))
            }

            for hi > 0 {
                pout[size] = uint64(hi % LongDecimalBase)
                size++
                hi /= LongDecimalBase
            }
        }
        
        if size == 0 {
            pout[size] = 0
            size++
        }

        strlen = addL + negative + 1 + (size - 1) * LongDecimalShift
        tenpow = 10
        rem = uint32(pout[size-1])
        for rem >= tenpow {
            tenpow *= 10
            strlen++
        }

        p = make([]byte, strlen)
        pos := strlen-1
        p[pos] = 0x00
        pos--

        for i=0; i < size - 1; i++ {
            rem = uint32(pout[i])
            for j = 0; j < LongDecimalShift; j++ {
                p[pos] = 0x30 + byte(rem % 10)
                pos--
                rem /= 10
            }
        }

        rem = uint32(pout[i])
        for rem != 0 {
            p[pos] = 0x30 + byte(rem % 10)
            pos--
            rem /= 10
        }

        key = string(p)
    }
    return key
}

func main(){
    if len(os.Args) == 2 {
        key := Sum128toString(os.Args[1])
        log.Printf("key = %v\n", key)
    }
    return
}