#include "jpegbw.h"

double complex func(double complex arg) {
  return cpow(.5*(ccos(M_PI*2.*arg)+1.), 2.);
}

double complex toon(double complex arg, double complex n) {
  return (double complex)((int)(arg*(n+1)))/n;
}

double complex vingette(double complex arg, double complex x, double complex y) {
  double complex hx, hy;
  hx = csqrt(2.)*(x-.5);
  hy = csqrt(2.)*(y-.5);
  return csqrt(arg)*(1.-csqrt(hx*hx+hy*hy));
}

double complex alpha(double complex arg, double complex period, double complex offset, double complex power) {
  return cpow(.5*(ccos(period*arg+offset)+1.), power);
}

double complex saturate(double complex arg, double complex lo, double complex hi) {
  double rarg = creal(arg);
  double rlo = creal(lo);
  double rhi = creal(hi);
  double rlov = cimag(lo);
  double rhiv = cimag(hi);
  if (rarg < rlo) return (double complex)rlov;
  if (rarg > rhi) return (double complex)rhiv;
  return (double complex)rarg;
}

double complex gsrainbowr(double complex arg) {
  double x = creal(arg);
  if (x > 0. && x < 1./7.) {
    return (double complex)(x * 7.);
  } else if (x >= 1./7. && x < 2./7.) {
    return (double complex)((2./7. - x) * 7.);
  } else if (x >= 2./7. && x < 4./7.) {
    return (double complex)0.;
  } else if (x >= 4./7. && x < 5./7.) {
    return (double complex)((x - 4./7.) * 7.);
  } else if (x >= 5./7. && x <= 1.) {
    return (double complex)1.;
  } else {
    return (double complex)0.;
  }
}

double complex gsrainbowg(double complex arg) {
  double x = creal(arg);
  if (x > 0. && x < 2./7.) {
    return (double complex)0.;
  } else if (x >= 2./7. && x < 3./7.) {
    return (double complex)((x - 2./7.) * 7.);
  } else if (x >= 3./7. && x < 5./7.) {
    return (double complex)1.;
  } else if (x >= 5./7. && x < 6./7.) {
    return (double complex)((6./7. - x) * 7.);
  } else if (x >= 6./7. && x <= 1.) {
    return (double complex)((x - 6./7.) * 7.);
  } else {
    return (double complex)0.;
  }
}

double complex gsrainbowb(double complex arg) {
  double x = creal(arg);
  if (x > 0. && x < 1./7.) {
    return (double complex)(x * 7.);
  } else if (x > 1./7. && x < 3./7.) {
    return (double complex)1.;
  } else if (x > 3./7. && x < 4./7.) {
    return (double complex)((4./7. - x) * 7.);
  } else if (x > 4./7. && x < 6./7.) {
    return (double complex)0.;
  } else if (x >= 6./7. && x <= 1.) {
    return (double complex)((x - 6./7.) * 7.);
  } else {
    return (double complex)0.;
  }
}
