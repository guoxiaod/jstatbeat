#include <stdio.h>
#include <stdbool.h>
#include <unistd.h>
#include "wrapper.h"

#define MAX_NAME 128

// read total cpu time
unsigned long long int read_cpu_tick() {
  unsigned long long int usertime, nicetime, systemtime, idletime;
  unsigned long long int ioWait, irq, softIrq, steal, guest, guestnice;
  usertime = nicetime = systemtime = idletime = 0;
  ioWait = irq = softIrq = steal = guest = guestnice = 0;

  FILE *fp;
  fp = fopen("/proc/stat", "r");
  if (fp != NULL) {
    if (fscanf(fp,   "cpu  %16llu %16llu %16llu %16llu %16llu %16llu %16llu %16llu %16llu %16llu",
               &usertime, &nicetime, &systemtime, &idletime,
               &ioWait, &irq, &softIrq, &steal, &guest, &guestnice) == EOF) {
      fclose(fp);
      return 0;
    } else {
      fclose(fp);
      return usertime + nicetime + systemtime + idletime + ioWait + irq + softIrq + steal + guest + guestnice;
    }
  }else{
    return 0;
  }
}

// read cpu tick for a specific process
cpustat read_time_from_pid(int pid) {

  char fn[MAX_NAME + 1];
  char name[MAX_NAME + 1];
  snprintf(fn, sizeof fn, "/proc/%i/stat", pid);

  cpustat cs = {0};

  FILE *fp;
  fp = fopen(fn, "r");
  if (fp != NULL) {
    bool ans = fscanf(fp, 
    //   1   2  3   4   5   6   7   8   9   10   11   12   13   14  15  16   17   18   19   20 
        "%*d %s %*c %*d %*d %*d %*d %*d %*u %*lu %*lu %*lu %*lu %lu %lu %*ld %*ld %*ld %*ld %*ld"
    //   21   22    23   24   25   26   27   28   29   30   31   32   33   34   35   36   37   38  39  40
        "%*ld %*llu %*lu %*ld %*lu %*lu %*lu %*lu %*lu %*lu %*lu %*lu %*lu %*lu %*lu %*lu %*lu %*d %*d %*u"
    //   41   42
        "%*u %llu",
        name, &cs.utime, &cs.stime, &cs.iotime) != EOF;
    if (!ans) {
      fclose(fp);
      return cs;
    } else {
      fclose(fp);
      return cs;
    }
  } else {
    return cs;
  }
}

// return number of cores
unsigned int num_cores(){
  return sysconf(_SC_NPROCESSORS_ONLN);
}

