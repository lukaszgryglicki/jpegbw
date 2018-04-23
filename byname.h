#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <dlfcn.h>

#define MAXFN 2048

double byname(char* fname, double arg, int* res);
double byname2(char* fname, double arg1, double arg2, int* res);
double byname3(char* fname, double arg1, double arg2, double arg3, int* res);
double byname4(char* fname, double arg1, double arg2, double arg3, double arg4, int* res);
int init(char* lib);
void tidy();
