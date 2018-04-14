// Copyright 2018 Hajime Hoshi
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
	"fmt"

	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack"
)

type UUID uuid.UUID

func NewUUID() UUID {
	return UUID(uuid.New())
}

func (u UUID) String() string {
	return uuid.UUID(u).String()
}

func (u UUID) MarshalText() ([]byte, error) {
	return uuid.UUID(u).MarshalText()
}

func (u *UUID) UnmarshalText(data []byte) error {
	return (*uuid.UUID)(u).UnmarshalText(data)
}

func (u UUID) MarshalBinary() ([]byte, error) {
	return uuid.UUID(u).MarshalBinary()
}

func (u *UUID) UnmarshalBinary(data []byte) error {
	return (*uuid.UUID)(u).UnmarshalBinary(data)
}

func (u UUID) EncodeMsgpack(enc *msgpack.Encoder) error {
	b, err := u.MarshalBinary()
	if err != nil {
		return err
	}
	return enc.EncodeBytes(b)
}

func (u *UUID) DecodeMsgpack(dec *msgpack.Decoder) error {
	b, err := dec.DecodeBytes()
	if err != nil {
		return err
	}
	switch len(b) {
	case 36:
		return u.UnmarshalText(b)
	case 16:
		return u.UnmarshalBinary(b)
	default:
		return fmt.Errorf("data: binary length must be 36 or 16; got %d", len(b))
	}
}
