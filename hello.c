#include <clockbound.h>
#include <stdio.h>

char const *shm_path = CLOCKBOUND_SHM_DEFAULT_PATH;
clockbound_ctx* ctx = NULL;

int open() {
  clockbound_err open_err;

  if (ctx == NULL) {
    ctx = clockbound_open(shm_path, &open_err);
    if (ctx == NULL) {
      print_clockbound_err("clockbound_open", &open_err);
      return 1;
    } else {
      printf("ctx created");
    }
  }

  return 0;
}

int close() {
  clockbound_err const *err;

  if (ctx != NULL) {
    err = clockbound_close(ctx);
    if (err) {
      print_clockbound_err("clockbound_close", err);
      return 1;
    } else {
      printf("ctx destroyed");
      ctx = NULL;
    }
  }

  return 0;
}

int now() {
  if (ctx == NULL) {
    printf("not init");
    return 1;
  }

  clockbound_now_result first;
  err = clockbound_now(ctx, &first);
  if (err) {
    print_clockbound_err("clockbound_now", err);
    return 1;
  }

  printf("When clockbound_now was called true time was somewhere within "
         "%ld.%09ld and %ld.%09ld seconds since Jan 1 1970. The clock status "
         "is %d.\n",
         first.earliest.tv_sec, first.earliest.tv_nsec, first.latest.tv_sec,
         first.latest.tv_nsec, first.clock_status);

  return 0;
}

/*
 * Helper function to print out errors returned by libclockbound.
 */
void print_clockbound_err(char const* detail, const clockbound_err *err) {
        fprintf(stderr, "%s: ", detail);
        switch (err->kind) {
                case CLOCKBOUND_ERR_NONE:
                        fprintf(stderr, "Success\n");
                        break;
                case CLOCKBOUND_ERR_SYSCALL:
                        if (err->detail) {
                                fprintf(stderr, "%s: %s\n", err->detail, strerror(err->sys_errno));
                        } else {
                                fprintf(stderr, "%s\n", strerror(err->sys_errno));
                        }
                        break;
                case CLOCKBOUND_ERR_SEGMENT_NOT_INITIALIZED:
                        fprintf(stderr, "Segment not initialized\n");
                        break;
                case CLOCKBOUND_ERR_SEGMENT_MALFORMED:
                        fprintf(stderr, "Segment malformed\n");
                        break;
                case CLOCKBOUND_ERR_CAUSALITY_BREACH:
                        fprintf(stderr, "Segment and clock reads out of order\n");
                        break;
        }
}
