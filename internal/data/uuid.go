// Copyright 2016 Hajime Hoshi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package data

import (
	"encoding/hex"
	"fmt"
	"regexp"

	"github.com/vmihailenco/msgpack"
)

var (
	uuidRe = regexp.MustCompile(`^([0-9a-f]{8})-([0-9a-f]{4})-([0-9a-f]{4})-([0-9a-f]{4})-([0-9a-f]{12})$`)
)

type UUID [16]uint8

func UUIDFromString(str string) (UUID, error) {
	var id UUID
	if err := id.UnmarshalText([]uint8(str)); err != nil {
		return UUID{}, err
	}
	return id, nil
}

func (u *UUID) String() string {
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex.EncodeToString(u[0:4]),
		hex.EncodeToString(u[4:6]),
		hex.EncodeToString(u[6:8]),
		hex.EncodeToString(u[8:10]),
		hex.EncodeToString(u[10:16]))
}

func (u *UUID) isZero() bool {
	for _, v := range u {
		if v != 0 {
			return false
		}
	}
	return true
}

func (u *UUID) MarshalText() ([]uint8, error) {
	return []uint8(u.String()), nil
}

func (u *UUID) UnmarshalText(text []uint8) error {
	m := uuidRe.FindStringSubmatch(string(text))
	if m == nil {
		return fmt.Errorf("data: invalid UUID format: %s", text)
	}
	if _, err := hex.Decode(u[0:4], []uint8(m[1])); err != nil {
		return err
	}
	if _, err := hex.Decode(u[4:6], []uint8(m[2])); err != nil {
		return err
	}
	if _, err := hex.Decode(u[6:8], []uint8(m[3])); err != nil {
		return err
	}
	if _, err := hex.Decode(u[8:10], []uint8(m[4])); err != nil {
		return err
	}
	if _, err := hex.Decode(u[10:16], []uint8(m[5])); err != nil {
		return err
	}
	if u.isZero() {
		return nil
	}
	if u[6]>>4 != 4 {
		return fmt.Errorf("data: UUID version must be 4: %s", text)
	}
	x := u[8] >> 4
	if x < 0x8 && 0xb < x {
		return fmt.Errorf("data: the two most significant bits of the clock_seq_hi_and_reserved part must be 0 and 1 respectively: %s", text)
	}
	return nil
}

func (u *UUID) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeBytes(u[:])
}

func (u *UUID) DecodeMsgpack(dec *msgpack.Decoder) error {
	b, err := dec.DecodeBytes()
	if err != nil {
		return err
	}
	copy(u[:], b)
	return nil
}
