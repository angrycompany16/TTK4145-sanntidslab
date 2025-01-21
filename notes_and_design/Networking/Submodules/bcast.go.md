This module handles general passing of structs (messages) between peers. It functions similarly to [[peers.go]], but has more general methods and uses reflection for arbitrary data type passing.

```go
func Transmitter(port int, chans ...interface{})
```
This function creates a connection for sending whatever is passed into the `chans` over `port`. `chans` can be channels of any data type, but there is a max size of `bufSize = 1024`. **If a struct that is too large is passed, the thread panics with an error message.** 

```go
func Receiver(port int, chans ...interface{})
```
This function receives whatever is sent on `port` and tries to parse it and pass it into the datatype(s) which are given in `chans`.

Both these functions also use the private methods `checkArgs` and `checkTypeRecursive`, which use reflection to convert the bytes passed over the network into Go structs or datatypes. 

The usage pattern looks like this:
```go
// Setup
helloTx := make(chan HelloMsg)
helloRx := make(chan HelloMsg)
go bcast.Transmitter(16569, helloTx)
go bcast.Receiver(16569, helloRx)

// Send a message every second
go func() {
	helloMsg := HelloMsg{"Hello from " + id, 0}
	for {
		helloMsg.Iter++
		helloTx <- helloMsg
		time.Sleep(1 * time.Second)
	}
}()

// Read messages
for {
	select {
		case a := <-helloRx:
			fmt.Printf("Received: %#v\n", a)
	}
}
```
