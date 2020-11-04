package proc

/*
#include "wrapper.h"
*/
import "C"

type Cpustat struct {
	Utime  uint64
	Stime  uint64
	Iotime uint64
}

func CpuTick() (t int64) {
	return int64(C.read_cpu_tick())
}

func TimeFromPid(pid int) (t Cpustat) {
	stat := C.read_time_from_pid(C.int(pid))
	return Cpustat{uint64(stat.utime), uint64(stat.stime), uint64(stat.iotime)}
}

func NumCores() (n int) {
	return int(C.num_cores())
}
