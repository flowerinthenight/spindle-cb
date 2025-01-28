#include <clockbound.h>
#include <stdio.h>

char const *shm_path = CLOCKBOUND_SHM_DEFAULT_PATH;
clockbound_ctx* ctx = NULL;

/* Open ClockBound's shared memory segment. */
int cb_open() {
  clockbound_err open_err;
  if (ctx == NULL) {
    ctx = clockbound_open(shm_path, &open_err);
    if (ctx == NULL) return open_err.sys_errno;
  }

  return 0;
}

/* Close ClockBound's shared memory segment. */
int cb_close() {
  if (ctx != NULL) {
    clockbound_err const *err;
    err = clockbound_close(ctx);
    if (err) {
      return err->kind;
    } else {
      ctx = NULL;
    }
  }

  return 0;
}

/* Read a bounded timestamp from ClockBound. */
int cb_now(int *e_s, int *e_ns, int *l_s, int *l_ns, int *s) {
  if (ctx == NULL) return 1;

  clockbound_err const *err;
  clockbound_now_result now;
  err = clockbound_now(ctx, &now);
  if (err) return err->kind;

  *e_s = now.earliest.tv_sec;
  *e_ns = now.earliest.tv_nsec;
  *l_s = now.latest.tv_sec;
  *l_ns = now.latest.tv_nsec;
  *s = now.clock_status;

  return 0;
}
