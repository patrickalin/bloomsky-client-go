package ring

// "fmt"
import (
	"fmt"
	"testing"
	"time"
)

type measure struct {
	f float64
}

func (m measure) Value() float64 {
	return m.f
}

func (m measure) TimeStamp() time.Time {
	return time.Now()
}
func TestSetsSize(t *testing.T) {
	r := &Ring{}
	r.SetCapacity(10)
	if r.Capacity() != 10 {
		t.Fatal("Size of ring was not 10", r.Capacity())
	}
}

func TestSavesSomeData(t *testing.T) {
	r := Ring{}
	r.SetCapacity(7)
	for i := 0; i <= 14; i++ {
		r.Enqueue(measure{float64(i)})
	}
	fmt.Println(r.head)
	fmt.Println(r.tail)

	for i := 0; i < 7; i++ {
		fmt.Println(r.Dequeue().Value())

	}
}

func TestDumpLine(t *testing.T) {
	r := Ring{}
	r.SetCapacity(7)
	for i := 0; i <= 14; i++ {
		r.Enqueue(measure{float64(i)})
	}
	s, err := r.DumpLine()
	if err != nil {
		panic(err)
	}
	fmt.Println(s)

}
