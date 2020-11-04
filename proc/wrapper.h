#ifndef GOCODE_WRAPPER_H
#define GOCODE_WRAPPER_H

typedef struct cpustat {
    unsigned long utime;
    unsigned long stime;
    unsigned long long iotime;
} cpustat;

unsigned long long int read_cpu_tick();
cpustat read_time_from_pid(int pid);
unsigned int num_cores();

#endif // GOCODE_WRAPPER_H
