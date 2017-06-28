package main

var DefaultCapacity = 256

type Ring struct {
	head int
	tail int
	buff []interface{}
}

/**
Set the maximum capacity of the ring
*/
func (r *Ring) SetCapacity(size int) {
	r.init()
	r.extends(size)
}

/**
Capacity returns the capacity of ringbuffer
*/
func (r *Ring) Capacity() int {
	return len(r.buff)
}

func (r *Ring) Enqueue(c interface{}) {
	r.init()
	r.set(r.head+1, c)
	old := r.head

	r.head = r.mod(r.head + 1)
	if old != -1 && r.head == r.tail {
		r.tail = r.mod(r.tail + 1)
	}

}

func (r *Ring) Dequeue() interface{} {
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

func (r *Ring) Values() []interface{} {
	if r.head == -1 {
		return nil
	}

	arr := make([]interface{}, 0, r.Capacity())

	for i := 0; i < r.Capacity(); i++ {
		idx := r.mod(i + r.tail)
		arr = append(arr, r.get(idx))
		if idx == r.head {
			break
		}
	}
	return arr
}

/*------------------------------------ */
func (r *Ring) mod(p int) int {
	return p % len(r.buff)
}

func (r *Ring) init() {
	if r.buff == nil {
		r.buff = make([]interface{}, DefaultCapacity)
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
	newbuffer := make([]interface{}, size-len(r.buff))
	for i := 0; i < len(newbuffer); i++ {
		newbuffer[i] = nil
	}
	r.buff = append(r.buff, newbuffer...)
}

func (r *Ring) set(i int, b interface{}) {
	r.buff[r.mod(i)] = b
}

func (r *Ring) get(i int) interface{} {
	return r.buff[r.mod(i)]
}
