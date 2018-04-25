#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <math.h>
#include <tgmath.h>
#include <complex.h>
#include <dlfcn.h>

double complex byname(char* fname, double complex arg, int* res);
double complex byname2(char* fname, double complex arg1, double complex arg2, int* res);
double complex byname3(char* fname, double complex arg1, double complex arg2, double complex arg3, int* res);
double complex byname4(char* fname, double complex arg1, double complex arg2, double complex arg3, double complex arg4, int* res);
int init(char* lib, size_t mfn);
void tidy(void);
