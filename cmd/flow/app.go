package main

import (
	"context"
	"fmt"
	"time"

	y3 "github.com/yomorun/y3-codec-golang"
	"github.com/yomorun/yomo/pkg/rx"
)

// NoiseDataKey represents the Tag of a Y3 encoded data packet.
const NoiseDataKey = 0x10

// ThresholdSingleValue is the threshold of a single value.
const ThresholdSingleValue = 16

// ThresholdAverageValue is the threshold of the average value after a sliding window.
const ThresholdAverageValue = 13

// NoiseData represents the structure of data
type NoiseData struct {
	Noise float32 `y3:"0x11"`
	Time  int64   `y3:"0x12"`
	From  string  `y3:"0x13"`
}

// Print every value and alert for value greater than ThresholdSingleValue
var computePeek = func(_ context.Context, i interface{}) (interface{}, error) {
	value := i.(NoiseData)
	rightNow := time.Now().UnixNano() / int64(time.Millisecond)
	fmt.Println(fmt.Sprintf("[%s] %d > value: %f ⚡️=%dms", value.From, value.Time, value.Noise, rightNow-value.Time))

	// Compute peek value, if greater than ThresholdSingleValue, alert
	if value.Noise >= ThresholdSingleValue {
		fmt.Println(fmt.Sprintf("❗ value: %f reaches the threshold %d! 𝚫=%f", value.Noise, ThresholdSingleValue, value.Noise-ThresholdSingleValue))
	}

	return value, nil
}

// Unserialize data to `NoiseData` struct, transfer to next process
var decode = func(v []byte) (interface{}, error) {
	var mold NoiseData
	err := y3.ToObject(v, &mold)
	if err != nil {
		return nil, err
	}
	mold.Noise = mold.Noise / 10
	return mold, nil
}

// Handler will handle data in Rx way
func Handler(rxstream rx.RxStream) rx.RxStream {
	stream := rxstream.
		Subscribe(NoiseDataKey).
		OnObserve(decode).
		Map(computePeek).
		Encode(0x10)

	return stream
}
