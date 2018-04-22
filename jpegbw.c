#include "jpegbw.h"

double func(double arg) {
  return pow(.5*(cos(M_PI*2.*arg)+1.), 2.);
}

double toon(double arg, double n) {
  return (double)((int)(arg*(n+1)))/n;
}
