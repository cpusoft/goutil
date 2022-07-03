package asn1node

import (
	"errors"
	"io"

	"github.com/cpusoft/goutil/belogs"
)

type Header struct {
	Class      int
	Tag        int
	IsCompound bool
}

func EqualHeaders(a, b Header) bool {
	if a.Class != b.Class {
		return false
	}
	if a.Tag != b.Tag {
		return false
	}
	if a.IsCompound != b.IsCompound {
		return false
	}
	return true
}

func EncodeHeader(data []byte, h *Header) ([]byte, error) {

	if h.Tag < 0 {
		belogs.Error("EncodeHeader(): fail, tag is negative %d", h.Tag)
		return data, errors.New("h.Tag is error")
	}

	b := (byte(h.Class) << 6)

	if h.IsCompound {
		b |= 0x20
	}

	tag := h.Tag

	// Low-tag-number form

	if tag < 0x1F {
		b |= byte(tag)
		data = append(data, b)
		return data, nil
	}

	// High-tag-number form

	var k int
	if tag := h.Tag; tag >= 0x1F {
		for tag > 0 {
			k++
			tag >>= 7
		}
	}

	bs := make([]byte, 1+k)
	b |= 0x1F
	bs[0] = b

	// last octet
	bs[k] = byte(tag & 0x7F)
	tag >>= 7
	k--

	for k > 0 {
		bs[k] = byte(tag&0x7F) | 0x80
		tag >>= 7
		k--
	}

	data = append(data, bs...)

	return data, nil
}

func DecodeHeader(data []byte, h *Header) (rest []byte, err error) {

	if len(data) == 0 {
		return data, io.EOF
	}

	b := data[0]
	data = data[1:]

	class := int(b >> 6)
	isCompound := ((b & 0x20) == 0x20)
	tag := int(b & 0x1F)

	// Low-tag-number form

	if tag != 0x1F {
		*h = Header{
			Class:      class,
			Tag:        tag,
			IsCompound: isCompound,
		}
		return data, nil
	}

	// High-tag-number form

	tag = 0
	for i := 0; ; i++ {

		if i > 5 {
			return data, errors.New("decode tag: hasn't final octet")
		}

		if len(data) == 0 {
			return data, errors.New("Insufficient data length")
		}
		b := data[0]
		data = data[1:]

		tag = (tag << 7) | int(b&0x7F)
		if (b & 0x80) == 0 {
			break
		}
	}

	*h = Header{
		Class:      class,
		Tag:        tag,
		IsCompound: isCompound,
	}

	return data, nil
}
