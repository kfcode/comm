package invitecode

import (
	"math/rand"
	"time"
)

var Char32 = []string { "T","9","C", "D", "E", "Y", "G", "6", "G", "L", "K", "2","N",
	"8", "Q", "S","R", "4", "V","W", "U", "X", "F", "Z","N", "3", "A", "5", "H", "7", "P","B"}

func internalGenInviteCode(number int32) string {

	resultNum := ""
	for i := 5; i >= 0; i-- {
		switch i {
		case 5:
			hexNum := int32(0x3E000000)
			maxNum := number & hexNum >> 25
			resultNum = resultNum + Char32[maxNum]
		case 4:
			hexNum := int32(0x01F00000)
			maxNum := number & hexNum >> 20
			resultNum = resultNum + Char32[maxNum]
		case 3:
			hexNum := int32(0x000F8000)
			maxNum := number & hexNum >> 15
			resultNum = resultNum + Char32[maxNum]
		case 2:
			hexNum := int32(0x00007B00)
			maxNum := number & hexNum >> 10
			resultNum = resultNum + Char32[maxNum]
		case 1:
			hexNum := int32(0x000003E0)
			maxNum := number & hexNum >> 5
			resultNum = resultNum + Char32[maxNum]
		case 0:
			hexNum := int32(0x0000001F)
			maxNum := number & hexNum
			resultNum = resultNum + Char32[maxNum]
		}
	}
	return resultNum
}

func GenerateSixNumInviteCode() string {

	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	randNum := r.Int31n(1024*1024*1024)
	return internalGenInviteCode(randNum)
}
