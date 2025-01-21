Just a single important function
```go
func LocalIP() (string, error) {
	...
}
```
This function automatically finds 
the IP address of the machine and returns it. It returns an error if, dialing `255.255.255.255:53` fails.