package ring

import (
	"bytes"
	"log"
	"text/template"
	"time"
)

const ts = `[{{range $i, $a := .}} {{if $i}},{{end}}[new Date("{{$a.TimeStamp.Format "Mon Jan _2 15:04:05 2006"}}"),{{printf "%.2f" $a.Value}}]{{end}}]`

var (
	t *template.Template
	// DefaultCapacity is max size of ring
	DefaultCapacity = 256
)

func init() {
	t = template.Must(template.New("ring").Parse(ts))

}

// TimeMeasure compose Time and Measure
type TimeMeasure interface {
	Time
	Measure
}

// Measure value of the measure
type Measure interface {
	//Value from the measure
	Value() float64
}

//Time is the timestamp of enqueue
type Time interface {
	//Timestamp from the measure
	TimeStamp() time.Time
}

//Ring structure
type Ring struct {
	head int
	tail int
	buff []TimeMeasure
}

/*SetCapacity fix the maximum capacity of the ring*/
func (r *Ring) SetCapacity(size int) {
	r.init()
	r.extends(size)
}

/*Capacity returns the capacity of ringbuffer*/
func (r *Ring) Capacity() int {
	return len(r.buff)
}

//Enqueue enqueues measure of the ring
func (r *Ring) Enqueue(c TimeMeasure) {
	if r == nil {
		log.Fatal("Ring is nil, not initialised stores[ \"xx\"] = &ring.Ring{}")
	}
	r.init()
	r.set(r.head+1, c)
	old := r.head

	r.head = r.mod(r.head + 1)
	if old != -1 && r.head == r.tail {
		r.tail = r.mod(r.tail + 1)
	}

}

//Dequeue dequeues measure of the ring
func (r *Ring) Dequeue() Measure {
	r.init()
	if r.head == -1 {
		return nil
	}

	v := r.get(r.tail)

	if r.tail == r.head {
		r.head = -1
		r.tail = 0

	} else {
		r.tail = r.mod(r.tail + 1)
	}
	return v
}

//Values return array of timeMeasure
func (r *Ring) Values() []TimeMeasure {
	if r.head == -1 {
		return nil
	}

	arr := make([]TimeMeasure, 0, r.Capacity())

	for i := 0; i < r.Capacity(); i++ {
		idx := r.mod(i + r.tail)
		arr = append(arr, r.get(idx))
		if idx == r.head {
			break
		}
	}
	return arr
}

//DumpLine return String of ring
func (r *Ring) DumpLine() (string, error) {
	var result bytes.Buffer
	if err := t.Execute(&result, r.Values()); err != nil {
		return "", err
	}
	return result.String(), nil
}

/*------------------------------------ */
func (r *Ring) mod(p int) int {
	return p % len(r.buff)
}

func (r *Ring) init() {
	if r.buff == nil {
		r.buff = make([]TimeMeasure, DefaultCapacity)
		for i := 0; i < len(r.buff); i++ {
			r.buff[i] = nil
		}
		r.head = -1
		r.tail = 0
	}

}

func (r *Ring) extends(size int) {
	if size == len(r.buff) {
		return
	}
	if size < len(r.buff) {
		r.buff = r.buff[0:size]
		return
	}
	newbuffer := make([]TimeMeasure, size-len(r.buff))
	for i := 0; i < len(newbuffer); i++ {
		newbuffer[i] = nil
	}
	r.buff = append(r.buff, newbuffer...)
}

func (r *Ring) set(i int, b TimeMeasure) {
	r.buff[r.mod(i)] = b
}

func (r *Ring) get(i int) TimeMeasure {
	return r.buff[r.mod(i)]
}
