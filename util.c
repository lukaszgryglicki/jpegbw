#include "util.h"

double complex clr(double complex z) {
  return cimag(z) * I;
}

double complex cli(double complex z) {
  return creal(z);
}
