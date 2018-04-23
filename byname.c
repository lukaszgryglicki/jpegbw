#include "byname.h"

static void* handle = 0;
static double (**fptra)(double) = 0;
static double (**fptra2)(double, double) = 0;
static double (**fptra3)(double, double, double) = 0;
static double (**fptra4)(double, double, double, double) = 0;
static char** fnames = 0;
static char** fnames2 = 0;
static char** fnames3 = 0;
static char** fnames4 = 0;
static int nptrs = 0;
static int nptrs2 = 0;
static int nptrs3 = 0;
static int nptrs4 = 0;

double byname(char* fname, double arg, int* res) {
  int i;
  double (*fptr)(double) = 0;
  if (!handle) {
    printf("byname %s,%f: library not open\n", fname, arg);
    *res = 1;
    return 0.0;
  }
  for (i=0;i<nptrs;i++) {
    if (!strcmp(fnames[i], fname)) {
      fptr = fptra[i];
    }
  }
  if (!fptr) {
    if (nptrs >= MAXFN) {
      printf("byname %s,%f: function table full\n", fname, arg);
      *res = 2;
      return 0.0;
    }
    fptr = (double (*)(double))dlsym(handle, fname);
    if (!fptr) {
      printf("byname %s,%f: function not found\n", fname, arg);
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

double byname2(char* fname, double arg1, double arg2, int* res) {
  int i;
  double (*fptr)(double, double) = 0;
  if (!handle) {
    printf("byname2 %s,%f,%f: library not open\n", fname, arg1, arg2);
    *res = 1;
    return 0.0;
  }
  for (i=0;i<nptrs2;i++) {
    if (!strcmp(fnames2[i], fname)) {
      fptr = fptra2[i];
    }
  }
  if (!fptr) {
    if (nptrs2 >= MAXFN) {
      printf("byname2 %s,%f,%f: function table full\n", fname, arg1, arg2);
      *res = 2;
      return 0.0;
    }
    fptr = (double (*)(double, double))dlsym(handle, fname);
    if (!fptr) {
      printf("byname2 %s,%f,%f: function not found\n", fname, arg1, arg2);
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

double byname3(char* fname, double arg1, double arg2, double arg3, int* res) {
  int i;
  double (*fptr)(double, double, double) = 0;
  if (!handle) {
    printf("byname3 %s,%f,%f,%f: library not open\n", fname, arg1, arg2, arg3);
    *res = 1;
    return 0.0;
  }
  for (i=0;i<nptrs3;i++) {
    if (!strcmp(fnames3[i], fname)) {
      fptr = fptra3[i];
    }
  }
  if (!fptr) {
    if (nptrs3 >= MAXFN) {
      printf("byname3 %s,%f,%f,%f: function table full\n", fname, arg1, arg2, arg3);
      *res = 2;
      return 0.0;
    }
    fptr = (double (*)(double, double, double))dlsym(handle, fname);
    if (!fptr) {
      printf("byname3 %s,%f,%f,%f: function not found\n", fname, arg1, arg2, arg3);
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

double byname4(char* fname, double arg1, double arg2, double arg3, double arg4, int* res) {
  int i;
  double (*fptr)(double, double, double, double) = 0;
  if (!handle) {
    printf("byname4 %s,%f,%f,%f,%f: library not open\n", fname, arg1, arg2, arg3, arg4);
    *res = 1;
    return 0.0;
  }
  for (i=0;i<nptrs4;i++) {
    if (!strcmp(fnames4[i], fname)) {
      fptr = fptra4[i];
    }
  }
  if (!fptr) {
    if (nptrs4 >= MAXFN) {
      printf("byname4 %s,%f,%f,%f,%f: function table full\n", fname, arg1, arg2, arg3, arg4);
      *res = 2;
      return 0.0;
    }
    fptr = (double (*)(double, double, double, double))dlsym(handle, fname);
    if (!fptr) {
      printf("byname4 %s,%f,%f,%f,%f: function not found\n", fname, arg1, arg2, arg3, arg4);
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

int init(char* lib) {
  handle = dlopen(lib, RTLD_LAZY);
  if (!handle) {
    printf("%s load result: %p\n", lib, handle);
    return 0;
  }
  fptra = malloc(MAXFN*sizeof(void*));
  fptra2 = malloc(MAXFN*sizeof(void*));
  fptra3 = malloc(MAXFN*sizeof(void*));
  fptra4 = malloc(MAXFN*sizeof(void*));
  fnames = (char**)malloc(MAXFN*sizeof(char*));
  fnames2 = (char**)malloc(MAXFN*sizeof(char*));
  fnames3 = (char**)malloc(MAXFN*sizeof(char*));
  fnames4 = (char**)malloc(MAXFN*sizeof(char*));
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

