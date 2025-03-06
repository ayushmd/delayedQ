# DelayedQ

Recieve messages from queue after a certain time or on a certain time. Useful for scheduling jobs, notifications.

```go
func main() {
	ttlq := NewTTLQueue()

	// Background queue listner
	go func() {
		for {
			select {
			case job := <-ttlq.Subscribe():
				jobj := job.(*TTLItem)
				fmt.Printf(
                    "Recieved Job %d: Created At: %d Recieved At: %d\n",
                    jobj.id, jobj.createdAt, time.Now().Unix(),
                )
			}
		}
	}()

	ttlq.Push(&TTLItem{
		id:        1,
		createdAt: time.Now().Unix(),
	}, time.Now().Add(10*time.Second).Unix())

	ttlq.Push(&TTLItem{
		id:        2,
		createdAt: time.Now().Unix(),
	}, time.Now().Add(5*time.Second).Unix())

	ttlq.Push(&TTLItem{
		id:        3,
		createdAt: time.Now().Unix(),
	}, time.Now().Add(2*time.Second).Unix())
	select {}
}
```
