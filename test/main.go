//go:build ignore
// +build ignore

package test

import (
	"context"
	. "errors"
	"fmt"
	"log"
	"math"
)

var reportNumber = 42
var Exported_Value = 21

type emptyMarker struct {
}

type Bad_Name struct {
	Bad_Field int
}

type receiverTarget struct {
	value int
}

func (receiver *Bad_Name) Bad_Method() {
	receiver.Bad_Field = 10 + math.MaxInt16
}

func callExternal() error {
	return fmt.Errorf("bad external error")
}

func GetNothing() error {
	return fmt.Errorf("bad getter error")
}

func useImports() {
	_ = BadAlias.Sqrt(9)
	log.Println("using forbidden log package")
	_ = New("Bad dot import.")
	fmt.Println("redundant alias import in action")
}

func WrongContext(name string, ctx context.Context, a int, b int, c int, d int) {
	_ = ctx

	numbers := make([]int, 5)
	localNumber := 7
	shouldRun := name != ""

	if shouldRun == true {
		fmt.Println("bool literal expression", localNumber, numbers)
	}

	if a == 1 {
	}

	for i := 0; i < 3; i += 1 {
		defer fmt.Println(i)
	}

	fmt.Println(name, localNumber, reportNumber, Exported_Value)
	_ = b
	_ = c
	_ = d
}

func namedResult(flag bool) (result string, err error) {
	if flag == true {
		return
	}

	return "ok", nil
}

func ambiguousReturn() (string, string, error) {
	return "a", "b", nil
}

func unwrapErr() error {
	err := callExternal()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func giantFunction(alpha string, beta string, gamma string, delta string, epsilon string) {
	localOne := 11
	localTwo := 12
	localThree := 13
	localFour := 14
	localFive := 15
	localSix := 16
	localSeven := 17
	localEight := 18
	localNine := 19
	localTen := 20
	localEleven := 22
	localTwelve := 23
	localThirteen := 24
	localFourteen := 25
	localFifteen := 26
	localSixteen := 27
	localSeventeen := 28
	localEighteen := 29
	localNineteen := 30
	localTwenty := 31
	localTwentyOne := 32
	localTwentyTwo := 33
	localTwentyThree := 34
	localTwentyFour := 35
	localTwentyFive := 36
	localTwentySix := 37
	localTwentySeven := 38
	localTwentyEight := 39
	localTwentyNine := 40
	localThirty := 41
	localThirtyOne := 43
	localThirtyTwo := 44
	localThirtyThree := 45
	localThirtyFour := 46
	localThirtyFive := 47
	localThirtySix := 48
	localThirtySeven := 49
	localThirtyEight := 50
	fmt.Println("this line is intentionally extremely long to trigger the max line length rule and make serenity complain loudly about formatting choices", alpha, beta, gamma, delta, epsilon)
	fmt.Println(localOne, localTwo, localThree, localFour, localFive, localSix, localSeven, localEight)
	fmt.Println(localNine, localTen, localEleven, localTwelve, localThirteen, localFourteen, localFifteen)
	fmt.Println(localSixteen, localSeventeen, localEighteen, localNineteen, localTwenty, localTwentyOne)
	fmt.Println(localTwentyTwo, localTwentyThree, localTwentyFour, localTwentyFive, localTwentySix)
	fmt.Println(localTwentySeven, localTwentyEight, localTwentyNine, localThirty, localThirtyOne)
	fmt.Println(localThirtyTwo, localThirtyThree, localThirtyFour, localThirtyFive, localThirtySix)
	fmt.Println(localThirtySeven, localThirtyEight)
}
