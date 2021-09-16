package ifchanged

import "fmt"

func ExampleNewLineTuple() {
	tuple, err := NewLineTuple("./linetuple_test.txt")
	defer tuple.Close()
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	tuple.Put([]byte("testKey"), []byte("TESTValue"))
	tuple.Put([]byte("testKey2"), []byte("TESTValue2"))
	tuple.Put([]byte("testKey3"), []byte("TESTValue3"))
	get, _ := tuple.Get([]byte("testKey"))
	fmt.Printf("%s", get)
	// Output: TESTValue
}
