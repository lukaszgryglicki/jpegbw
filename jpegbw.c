#include "jpegbw.h"

double complex func(double complex arg) {
  return pow(.5*(cos(M_PI*2.*arg)+1.), 2.);
}

double complex toon(double complex arg, double complex n) {
  return (double complex)((int)(arg*(n+1)))/n;
}

double complex vingette(double complex arg, double complex x, double complex y) {
  double complex hx, hy;
  hx = sqrt(2.)*(x-.5);
  hy = sqrt(2.)*(y-.5);
  return sqrt(arg)*(1.-sqrt(hx*hx+hy*hy));
}

double complex alpha(double complex arg, double complex period, double complex offset, double complex power) {
  return pow(.5*(cos(period*arg+offset)+1.), power);
}
