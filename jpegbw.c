#include "jpegbw.h"

double func(double arg) {
  return pow(.5*(cos(M_PI*2.*arg)+1.), 2.);
}

double toon(double arg, double n) {
  return (double)((int)(arg*(n+1)))/n;
}

double vingette(double arg, double x, double y) {
  /* alternative */
  /*
  double hx, hy;
  hx = sqrt(2.)*(x-.5);
  hy = sqrt(2.)*(y-.5);
  return sqrt(arg)*(1.-sqrt(hx*hx+hy*hy));
  */
  return sqrt(arg)*(1.-hypot(sqrt(2.)*(x-.5), sqrt(2.)*(y-.5)));
}

double alpha(double arg, double period, double offset, double power) {
  return pow(.5*(cos(period*arg+offset)+1.), power);
}
