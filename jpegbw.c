#include "jpegbw.h"

double func(double arg) {
  return pow(.5*(cos(M_PI*2.*arg)+1.), 2.);
}

double toon(double arg, double n) {
  return (double)((int)(arg*(n+1)))/n;
}

double vingette(double arg, double x, double y) {
  return sqrt(arg)*(1.-hypot(sqrt(2.)*(x-.5), sqrt(2.)*(y-.5)));
}
