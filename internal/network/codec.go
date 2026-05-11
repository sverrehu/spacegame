package network

import (
	"bytes"
	"encoding/binary"

	"github.com/sverrehu/spacegame/internal/model"
	"github.com/sverrehu/spacegame/internal/utils"
)

func encodeHitBy(buf *bytes.Buffer, t model.WeaponType, killed bool) error {
	err := encodeByte(buf, uint8(t))
	if err != nil {
		return err
	}
	return encodeBool(buf, killed)
}

func decodeHitBy(buf *bytes.Buffer) (model.WeaponType, bool, error) {
	t, err := decodeByte(buf)
	if err != nil {
		return 0, false, err
	}
	killed, err := decodeBool(buf)
	if err != nil {
		return 0, false, err
	}
	return model.WeaponType(t), killed, nil
}

func encodeInGameMessage(buf *bytes.Buffer, t InGameMessageType, text string) error {
	err := encodeByte(buf, uint8(t))
	if err != nil {
		return err
	}
	return encodeString(buf, text)
}

func decodeInGameMessage(buf *bytes.Buffer) (InGameMessageType, string, error) {
	t, err := decodeByte(buf)
	if err != nil {
		return 0, "", err
	}
	text, err := decodeString(buf)
	if err != nil {
		return 0, "", err
	}
	return InGameMessageType(t), text, nil
}

func encodeUpdates(buf *bytes.Buffer, updates []model.AnyObjectUpdate) error {
	err := encodeInt32(buf, int32(len(updates)))
	if err != nil {
		return err
	}
	for _, update := range updates {
		err := encodeUpdate(buf, &update)
		if err != nil {
			return err
		}
	}
	return nil
}

func encodeUpdate(buf *bytes.Buffer, update *model.AnyObjectUpdate) error {
	err := encodeGameObjectAdditionals(buf, update)
	if err != nil {
		return err
	}
	switch update.Type {
	case model.ObjectStar:
		err = encodeStarAdditionals(buf, update)
	case model.ObjectShip:
		err = encodeShipAdditionals(buf, update)
	case model.ObjectPhaser:
		err = encodePhaserAdditionals(buf, update)
	case model.ObjectBomb:
		err = encodeBombAdditionals(buf, update)
	case model.ObjectBombPack:
		err = encodeBombPackAdditionals(buf, update)
	case model.ObjectExplosion:
		err = encodeExplosionAdditionals(buf, update)
	default:
		panic("unknown update type")
	}
	return err
}

func encodeGameObjectAdditionals(buf *bytes.Buffer, u *model.AnyObjectUpdate) error {
	var fullFlag uint8 = 0
	if u.Full {
		fullFlag = 1 << 7
	}
	err := encodeByte(buf, fullFlag|uint8(u.Type))
	if err != nil {
		return err
	}
	err = encodeInt32(buf, u.Id)
	if err != nil {
		return err
	}
	err = encodeFloat64(buf, u.Position.X)
	if err != nil {
		return err
	}
	err = encodeFloat64(buf, u.Position.Y)
	if err != nil {
		return err
	}
	if u.Full {
		err = encodeColor(buf, u.Color)
		if err != nil {
			return err
		}
	}
	return encodeBool(buf, u.Delete)
}

func encodeStarAdditionals(buf *bytes.Buffer, u *model.AnyObjectUpdate) error {
	return nil
}

func encodeShipAdditionals(buf *bytes.Buffer, u *model.AnyObjectUpdate) error {
	err := encodeFloat64(buf, u.Direction)
	if err != nil {
		return err
	}
	err = encodeBool(buf, u.IsAlive)
	if err != nil {
		return err
	}
	err = encodeInt16(buf, u.Score)
	if err != nil {
		return err
	}
	err = encodeInt16(buf, u.AntiScore)
	if err != nil {
		return err
	}
	if u.Full {
		err = encodeColor(buf, u.Color)
		if err != nil {
			return err
		}
		err = encodeString(buf, u.Name)
		if err != nil {
			return err
		}
	}
	err = encodeByte(buf, u.Damage)
	if err != nil {
		return err
	}
	err = encodeByte(buf, u.PhaserHeat)
	if err != nil {
		return err
	}
	err = encodeInt16(buf, u.BombsLeft)
	if err != nil {
		return err
	}
	return nil
}

func encodePhaserAdditionals(buf *bytes.Buffer, u *model.AnyObjectUpdate) error {
	return nil
}

func encodeBombAdditionals(buf *bytes.Buffer, u *model.AnyObjectUpdate) error {
	err := encodeBool(buf, u.Flip)
	if err != nil {
		return err
	}
	return nil
}

func encodeBombPackAdditionals(buf *bytes.Buffer, u *model.AnyObjectUpdate) error {
	return nil
}

func encodeExplosionAdditionals(buf *bytes.Buffer, u *model.AnyObjectUpdate) error {
	err := encodeFloat64(buf, u.InnerRadius)
	if err != nil {
		return err
	}
	return encodeFloat64(buf, u.OuterRadius)
}

func decodeUpdates(buf *bytes.Buffer) ([]model.AnyObjectUpdate, error) {
	l, err := decodeInt32(buf)
	if err != nil {
		return nil, err
	}
	updates := make([]model.AnyObjectUpdate, 0, l)
	for i := 0; i < int(l); i++ {
		update, err := decodeUpdate(buf)
		if err != nil {
			return nil, err
		}
		updates = append(updates, *update)
	}
	return updates, nil
}

func decodeUpdate(buf *bytes.Buffer) (*model.AnyObjectUpdate, error) {
	u := &model.AnyObjectUpdate{}
	err := decodeGameObjectAdditionals(buf, u)
	if err != nil {
		return nil, err
	}
	switch u.Type {
	case model.ObjectStar:
		err = decodeStarAdditionals(buf, u)
	case model.ObjectShip:
		err = decodeShipAdditionals(buf, u)
	case model.ObjectPhaser:
		err = decodePhaserAdditionals(buf, u)
	case model.ObjectBomb:
		err = decodeBombAdditionals(buf, u)
	case model.ObjectBombPack:
		err = decodeBombPackAdditionals(buf, u)
	case model.ObjectExplosion:
		err = decodeExplosionAdditionals(buf, u)
	default:
		panic("unknown update type")
	}
	return u, nil
}

func decodeGameObjectAdditionals(buf *bytes.Buffer, u *model.AnyObjectUpdate) error {
	t, err := decodeByte(buf)
	if err != nil {
		return err
	}
	fullFlag := t & (1 << 7)
	u.Full = fullFlag != 0
	u.Type = model.ObjectType(t & 127)
	id, err := decodeInt32(buf)
	if err != nil {
		return err
	}
	u.Id = id
	x, err := decodeFloat64(buf)
	if err != nil {
		return err
	}
	u.Position.X = x
	y, err := decodeFloat64(buf)
	if err != nil {
		return err
	}
	if u.Full {
		col, err := decodeColor(buf)
		if err != nil {
			return err
		}
		u.Color = col
	}
	u.Position.Y = y
	del, err := decodeBool(buf)
	if err != nil {
		return err
	}
	u.Delete = del
	return nil
}

func decodeStarAdditionals(buf *bytes.Buffer, u *model.AnyObjectUpdate) error {
	return nil
}

func decodeShipAdditionals(buf *bytes.Buffer, u *model.AnyObjectUpdate) error {
	dir, err := decodeFloat64(buf)
	if err != nil {
		return err
	}
	isAlive, err := decodeBool(buf)
	if err != nil {
		return err
	}
	score, err := decodeInt16(buf)
	if err != nil {
		return err
	}
	antiScore, err := decodeInt16(buf)
	if err != nil {
		return err
	}
	var col utils.Color
	var name string
	if u.Full {
		col, err = decodeColor(buf)
		if err != nil {
			return err
		}
		name, err = decodeString(buf)
		if err != nil {
			return err
		}
	}
	damage, err := decodeByte(buf)
	if err != nil {
		return err
	}
	phaserHeat, err := decodeByte(buf)
	if err != nil {
		return err
	}
	bombsLeft, err := decodeInt16(buf)
	if err != nil {
		return err
	}
	u.Direction = dir
	u.Color = col
	u.Name = name
	u.IsAlive = isAlive
	u.Score = score
	u.AntiScore = antiScore
	u.Damage = damage
	u.PhaserHeat = phaserHeat
	u.BombsLeft = bombsLeft
	return nil
}

func decodePhaserAdditionals(buf *bytes.Buffer, u *model.AnyObjectUpdate) error {
	return nil
}

func decodeBombAdditionals(buf *bytes.Buffer, u *model.AnyObjectUpdate) error {
	f, err := decodeBool(buf)
	if err != nil {
		return err
	}
	u.Flip = f
	return nil
}

func decodeBombPackAdditionals(buf *bytes.Buffer, u *model.AnyObjectUpdate) error {
	return nil
}

func decodeExplosionAdditionals(buf *bytes.Buffer, u *model.AnyObjectUpdate) error {
	ir, err := decodeFloat64(buf)
	if err != nil {
		return err
	}
	or, err := decodeFloat64(buf)
	if err != nil {
		return err
	}
	u.InnerRadius = ir
	u.OuterRadius = or
	return nil
}

func encodeString(buf *bytes.Buffer, s string) error {
	err := encodeInt32(buf, int32(len(s)))
	if err != nil {
		return err
	}
	buf.WriteString(s)
	return nil
}

func decodeString(buf *bytes.Buffer) (string, error) {
	l, err := decodeInt32(buf)
	if err != nil {
		return "", err
	}
	return string(buf.Next(int(l))), nil
}

func encodeFloat64(buf *bytes.Buffer, v float64) error {
	//return binary.Write(buf, binary.BigEndian, v)
	// encoded as float32 to save space
	return encodeFloat32(buf, float32(v))
}

func decodeFloat64(buf *bytes.Buffer) (float64, error) {
	// encoded as float32 to save space
	v, err := decodeFloat32(buf)
	if err != nil {
		return 0, err
	}
	return float64(v), nil
}

func encodeFloat32(buf *bytes.Buffer, v float32) error {
	return binary.Write(buf, binary.BigEndian, v)
}

func decodeFloat32(buf *bytes.Buffer) (float32, error) {
	var v float32
	err := binary.Read(buf, binary.BigEndian, &v)
	return v, err
}

func encodeInt32(buf *bytes.Buffer, v int32) error {
	return binary.Write(buf, binary.BigEndian, v)
}

func decodeInt32(buf *bytes.Buffer) (int32, error) {
	var v int32
	err := binary.Read(buf, binary.BigEndian, &v)
	return v, err
}

func encodeInt16(buf *bytes.Buffer, v int16) error {
	return binary.Write(buf, binary.BigEndian, v)
}

func decodeInt16(buf *bytes.Buffer) (int16, error) {
	var v int16
	err := binary.Read(buf, binary.BigEndian, &v)
	return v, err
}

func encodeByte(buf *bytes.Buffer, v uint8) error {
	return buf.WriteByte(v)
}

func decodeByte(buf *bytes.Buffer) (uint8, error) {
	return buf.ReadByte()
}

func encodeBool(buf *bytes.Buffer, v bool) error {
	return buf.WriteByte(boolToByte(v))
}

func decodeBool(buf *bytes.Buffer) (bool, error) {
	b, err := buf.ReadByte()
	if err != nil {
		return false, err
	}
	return byteToBool(b), nil
}

func boolToByte(v bool) uint8 {
	if v {
		return uint8(1)
	}
	return uint8(0)
}

func byteToBool(v uint8) bool {
	return v != uint8(0)
}

func encodeColor(buf *bytes.Buffer, col utils.Color) error {
	r, g, b := col.ToRGB255()
	err := encodeByte(buf, r)
	if err != nil {
		return err
	}
	err = encodeByte(buf, g)
	if err != nil {
		return err
	}
	return encodeByte(buf, b)
}

func decodeColor(buf *bytes.Buffer) (utils.Color, error) {
	r, err := decodeByte(buf)
	if err != nil {
		return utils.Color{}, err
	}
	g, err := decodeByte(buf)
	if err != nil {
		return utils.Color{}, err
	}
	b, err := decodeByte(buf)
	if err != nil {
		return utils.Color{}, err
	}
	return utils.NewColor255(r, g, b), nil
}
