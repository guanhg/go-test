package pkg

import (
	"fmt"
)

var testCnt int

// 收集日志 要求：1 每收集size条日志才写入文件 2 使用管道
type LogCollector struct {
	size int32
	c    chan string
}

func NewLogCollector(size int32) *LogCollector {
	l := &LogCollector{
		size: size,
		c:    make(chan string, size+1),
	}

	go l.doCollect()

	return l
}

func (l *LogCollector) Collect(str string) {
	l.c <- str
}

func (l *LogCollector) doCollect() {
	var ary []string

	for log := range l.c {
		ary = append(ary, log)

		if int32(len(ary)) >= l.size {
			h := ary[:l.size]
			ary = ary[l.size:]
			// 写日志文件
			testCnt++
			fmt.Println(testCnt, h)
		}
	}

	if len(ary) > 0 {
		fmt.Println(ary)
	}
}

func (l *LogCollector) Close() {
	close(l.c)
}
