package rewledis

//go:generate genny -in=genny-deque/deque.go -out=slotdeque_gen.go -pkg=rewledis gen "ValueType=Slot"
