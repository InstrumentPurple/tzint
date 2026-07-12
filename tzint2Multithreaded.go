package main


import("os"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"strconv"
)


//Copyright InstrumentPurple(Github) Dec. 21, 2025
/* Permission is hereby granted, free of charge,
 *  to any person obtaining a copy of this software
 *  and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including
 *  without limitation the rights to use, copy, modify,
 *  merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject
 *  to the following conditions:
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
 *  EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
 *  OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
 *  NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
 *  HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 *  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
 *  OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */







// google ai vibed AtomicFloat64


// AtomicFloat64 is a 64-bit float variable that can be safely used concurrently.
type AtomicFloat64 struct {
	bits uint64
}

// Add atomically adds a delta to the float value and returns the new value.
func (af *AtomicFloat64) Add(delta float64) (newf float64) {
	for {
		oldbits := atomic.LoadUint64(&af.bits) // Atomically load the current bits
		oldf := math.Float64frombits(oldbits)  // Convert bits to float64
		newf = oldf + delta                    // Calculate the new float value
		newbits := math.Float64bits(newf)      // Convert the new float value back to bits

		// Atomically compare and swap; if the value hasn't changed, store the new bits
		if atomic.CompareAndSwapUint64(&af.bits, oldbits, newbits) {
			return newf
		}
		// If the value changed in between, the loop retries
	}
}

// Load atomically loads the float value.
func (af *AtomicFloat64) Load() float64 {
	return math.Float64frombits(atomic.LoadUint64(&af.bits))
}

// Store atomically stores the float value.
func (af *AtomicFloat64) Store(val float64) {
	atomic.StoreUint64(&af.bits, math.Float64bits(val))
}




////////////////////////////////////////









func eval(poly sync.Map, at float64)float64{
	total := 0.0
	poly.Range(func (expo, coef interface{})bool{
		total += coef.(float64)*math.Pow(at, expo.(float64))
		return true
	})

	return total
}

func sumRange(poly sync.Map, wg *sync.WaitGroup, TOTAL *AtomicFloat64, isEnd bool, left, right, stepSize float64){
	i := left

	if isEnd{
		for i < right-stepSize{ /* within ends variable. discard the last one */
			TOTAL.Add(eval(poly,i))
			i += stepSize
		}
	} else {
		for i < right{
			TOTAL.Add(eval(poly,i))
			i += stepSize
		}
	}
	wg.Done()
}

// tzint2Multithreaded.exe [trapezoid count] [left bound] [right bound] [exponent] [coef] ...
/* EXAMPLE:  .\tzint2Multithreaded.exe 90000000 -1.5 1.5 2 -4 0 9 */
func main(){
	if len(os.Args) < 4 {
		fmt.Println("tzint2Multithreaded.exe [trapezoid count] [left bound] [right bound] [exponent] [coef] ...")
		os.Exit(0)
	}
	if ((len(os.Args) - 4) % 2) == 1 {
		fmt.Println("odd number of polynomial arguments. Each exponent must be mapped to one coeffcient.")
		os.Exit(0)
	}


	var wg sync.WaitGroup
	var f sync.Map

	cur:= 4 // start with the polynomial
	for cur < len(os.Args){
		expo,_ := strconv.ParseFloat(os.Args[cur], 64)
		coef,_ := strconv.ParseFloat(os.Args[cur+1], 64)

		f.Store(expo,coef)
		cur += 2
	}


	count,_ := strconv.ParseFloat(os.Args[1], 64)

	left ,_ := strconv.ParseFloat(os.Args[2], 64)
	right,_ := strconv.ParseFloat(os.Args[3], 64)

	if left >= right{
		fmt.Println("invalid bounds")
	}

	ends := eval(f, left) + eval(f, right)

	i := left
	stepSize := math.Abs(right - left)/count
	i += stepSize //skip left bound. It is accounted for in ends variable



	tcount := 16
	wg.Add(tcount)
	var TOTAL AtomicFloat64
	chunkSize := math.Abs(right - left)/float64(tcount)

	for i < right{
		if (i+chunkSize) >= right{ /* the last iteration of the loop */
			go sumRange(f,&wg, &TOTAL, true, i, i + chunkSize, stepSize)
		} else {
			go sumRange(f,&wg, &TOTAL, false, i, i + chunkSize, stepSize)
		}

		i += chunkSize
	}


	wg.Wait()
	total := TOTAL.Load()

	total *= 2.0 // distributive property on all terms except the ends
	total += ends
	total *= (right-left)/(2.0*count) // multiply the factored out term

	fmt.Println(total)
}
