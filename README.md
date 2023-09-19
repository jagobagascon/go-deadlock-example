This is an example project to reproduce this issue: https://github.com/golang/go/issues/55015

The project contains two folders.
Both of them contain the exact same code: a (not so) small example of an application scanning for BLE devices using the WinRT BLE APIs.

Both applications start a BLE scan operation, and wait indefinitely. Windows will execute the callback everytime a device is found.

The only difference is that one of them uses `malloc` to allocate memory into the heap, and thus, uses CGO. And the other one uses `HeapAlloc` via syscall, so it does not need CGO.

> [!NOTE]
> `example-malloc` is pointing to the latest release of https://github.com/saltosystems/winrt-go, which uses `malloc` for allocation of delegates.
> `example-heapalloc` is pointing to this PR https://github.com/saltosystems/winrt-go/pull/60, which replaces it for the `HeapAlloc` function.

## Expected behaviour
Both applications should list the address of the devices they find.

## What really happens

- `example-malloc` is working as expected: listing the address of the bluetooth devices it scans.
- `example-heapalloc` is failing with the following error:
```
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [chan receive]:
main.runExample()
        C:/projects/go-deadlock-example/example-heapalloc/main.go:79 +0x3ee
main.main()
        C:/projects/go-deadlock-example/example-heapalloc/main.go:14 +0x19 
exit status 2
```

If you add a goroutine to `example-heapalloc`, then it works as expected. Even if the goroutine does nothing but wait:
```go
go func() {
    <-time.After(5 * time.Second)
}()
```

This issue seems related to https://github.com/golang/go/issues/6751
