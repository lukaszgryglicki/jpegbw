#include "byname.h"

static void* handle = 0;
static size_t maxfn = 0;
static double complex (**fptra)(double complex) = 0;
static double complex (**fptra2)(double complex, double complex) = 0;
static double complex (**fptra3)(double complex, double complex, double complex) = 0;
static double complex (**fptra4)(double complex, double complex, double complex, double complex) = 0;
static char** fnames = 0;
static char** fnames2 = 0;
static char** fnames3 = 0;
static char** fnames4 = 0;
static int nptrs = 0;
static int nptrs2 = 0;
static int nptrs3 = 0;
static int nptrs4 = 0;

double complex byname(char* fname, double complex arg, int* res) {
  int i;
  double complex (*fptr)(double complex) = 0;
  if (!handle) {
    printf("byname %s,%f+%fi library not open\n", fname, creal(arg), cimag(arg));
    *res = 1;
    return 0.0;
  }
  for (i=0;i<nptrs;i++) {
    if (!strcmp(fnames[i], fname)) {
      fptr = fptra[i];
    }
  }
  if (!fptr) {
    if (nptrs >= maxfn) {
      printf("byname %s,%f+%fi function table full\n", fname, creal(arg), cimag(arg));
      *res = 2;
      return 0.0;
    }
    fptr = (double complex (*)(double complex))dlsym(handle, fname);
    if (!fptr) {
      printf("byname %s,%f+%fi function not found\n", fname, creal(arg), cimag(arg));
      *res = 3;
      return 0.0;
    }
    fptra[nptrs] = fptr;
    fnames[nptrs] = (char*)malloc((strlen(fname)+1)*sizeof(char));
    strcpy(fnames[nptrs], fname);
    nptrs ++;
  }
  *res = 0;
  return (*fptr)(arg);
}

double complex byname2(char* fname, double complex arg1, double complex arg2, int* res) {
  int i;
  double complex (*fptr)(double complex, double complex) = 0;
  if (!handle) {
    printf("byname2 %s,%f+%fi,%f+%fi: library not open\n", fname, creal(arg1), cimag(arg1), creal(arg2), cimag(arg2));
    *res = 1;
    return 0.0;
  }
  for (i=0;i<nptrs2;i++) {
    if (!strcmp(fnames2[i], fname)) {
      fptr = fptra2[i];
    }
  }
  if (!fptr) {
    if (nptrs2 >= maxfn) {
      printf("byname2 %s,%f+%fi,%f+%fi: function table full\n", fname, creal(arg1), cimag(arg1), creal(arg2), cimag(arg2));
      *res = 2;
      return 0.0;
    }
    fptr = (double complex (*)(double complex, double complex))dlsym(handle, fname);
    if (!fptr) {
      printf("byname2 %s,%f+%fi,%f+%fi: function not found\n", fname, creal(arg1), cimag(arg1), creal(arg2), cimag(arg2));
      *res = 3;
      return 0.0;
    }
    fptra2[nptrs2] = fptr;
    fnames2[nptrs2] = (char*)malloc((strlen(fname)+1)*sizeof(char));
    strcpy(fnames2[nptrs2], fname);
    nptrs2 ++;
  }
  *res = 0;
  return (*fptr)(arg1, arg2);
}

double complex byname3(char* fname, double complex arg1, double complex arg2, double complex arg3, int* res) {
  int i;
  double complex (*fptr)(double complex, double complex, double complex) = 0;
  if (!handle) {
    printf("byname3 %s,%f+%fi,%f+%fi,%f+%fi: library not open\n", fname, creal(arg1), cimag(arg1), creal(arg2), cimag(arg2), creal(arg3), cimag(arg3));
    *res = 1;
    return 0.0;
  }
  for (i=0;i<nptrs3;i++) {
    if (!strcmp(fnames3[i], fname)) {
      fptr = fptra3[i];
    }
  }
  if (!fptr) {
    if (nptrs3 >= maxfn) {
      printf("byname3 %s,%f+%fi,%f+%fi,%f+%fi: function table full\n", fname, creal(arg1), cimag(arg1), creal(arg2), cimag(arg2), creal(arg3), cimag(arg3));
      *res = 2;
      return 0.0;
    }
    fptr = (double complex (*)(double complex, double complex, double complex))dlsym(handle, fname);
    if (!fptr) {
      printf("byname3 %s,%f+%fi,%f+%fi,%f+%fi: function not found\n", fname, creal(arg1), cimag(arg1), creal(arg2), cimag(arg2), creal(arg3), cimag(arg3));
      *res = 3;
      return 0.0;
    }
    fptra3[nptrs3] = fptr;
    fnames3[nptrs3] = (char*)malloc((strlen(fname)+1)*sizeof(char));
    strcpy(fnames3[nptrs3], fname);
    nptrs3 ++;
  }
  *res = 0;
  return (*fptr)(arg1, arg2, arg3);
}

double complex byname4(char* fname, double complex arg1, double complex arg2, double complex arg3, double complex arg4, int* res) {
  int i;
  double complex (*fptr)(double complex, double complex, double complex, double complex) = 0;
  if (!handle) {
    printf("byname4 %s,%f+%fi,%f+%fi,%f+%fi,%f+%fi: library not open\n", fname, creal(arg1), cimag(arg1), creal(arg2), cimag(arg2), creal(arg3), cimag(arg3), creal(arg4), cimag(arg4));
    *res = 1;
    return 0.0;
  }
  for (i=0;i<nptrs4;i++) {
    if (!strcmp(fnames4[i], fname)) {
      fptr = fptra4[i];
    }
  }
  if (!fptr) {
    if (nptrs4 >= maxfn) {
      printf("byname4 %s,%f+%fi,%f+%fi,%f+%fi,%f+%fi: function table full\n", fname, creal(arg1), cimag(arg1), creal(arg2), cimag(arg2), creal(arg3), cimag(arg3), creal(arg4), cimag(arg4));
      *res = 2;
      return 0.0;
    }
    fptr = (double complex (*)(double complex, double complex, double complex, double complex))dlsym(handle, fname);
    if (!fptr) {
      printf("byname4 %s,%f+%fi,%f+%fi,%f+%fi,%f+%fi: function not found\n", fname, creal(arg1), cimag(arg1), creal(arg2), cimag(arg2), creal(arg3), cimag(arg3), creal(arg4), cimag(arg4));
      *res = 3;
      return 0.0;
    }
    fptra4[nptrs4] = fptr;
    fnames4[nptrs4] = (char*)malloc((strlen(fname)+1)*sizeof(char));
    strcpy(fnames4[nptrs4], fname);
    nptrs4 ++;
  }
  *res = 0;
  return (*fptr)(arg1, arg2, arg3, arg4);
}

int init(char* lib, size_t mfn) {
  if (mfn < 1) {
    printf("init(%s, %ld): mfn must be >= 1\n", lib, mfn);
    return 0;
  }
  maxfn = mfn;
  handle = dlopen(lib, RTLD_LAZY);
  if (!handle) {
    printf("init(%s, %ld): cannot load library\n", lib, mfn);
    return 0;
  }
  fptra = malloc(maxfn*sizeof(void*));
  fptra2 = malloc(maxfn*sizeof(void*));
  fptra3 = malloc(maxfn*sizeof(void*));
  fptra4 = malloc(maxfn*sizeof(void*));
  fnames = (char**)malloc(maxfn*sizeof(char*));
  fnames2 = (char**)malloc(maxfn*sizeof(char*));
  fnames3 = (char**)malloc(maxfn*sizeof(char*));
  fnames4 = (char**)malloc(maxfn*sizeof(char*));
  if (!fptra || !fnames || !fptra2 || !fnames2 || !fptra3 || !fnames3 || !fptra4 || !fnames4) {
    printf("%s malloc failed\n", lib);
    return 0;
  }
  return 1;
}

void tidy(void) {
  if (handle) {
    dlclose(handle);
    handle = 0;
  }
  if (fptra) {
    free((void*)fptra);
    fptra = 0;
  }
  if (fptra2) {
    free((void*)fptra2);
    fptra2 = 0;
  }
  if (fptra3) {
    free((void*)fptra3);
    fptra3 = 0;
  }
  if (fptra4) {
    free((void*)fptra4);
    fptra4 = 0;
  }
  if (fnames) {
    free((void*)fnames);
    fnames = 0;
  }
  if (fnames2) {
    free((void*)fnames2);
    fnames2 = 0;
  }
  if (fnames3) {
    free((void*)fnames3);
    fnames3 = 0;
  }
  if (fnames4) {
    free((void*)fnames4);
    fnames4 = 0;
  }
  nptrs = 0;
  nptrs2 = 0;
  nptrs3 = 0;
  nptrs4 = 0;
}

