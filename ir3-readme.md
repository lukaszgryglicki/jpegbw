IR3 v4 defaults

Full map:
  IRRATIO=2.5
  IRS=1.0
  IRL=1.0
  IRSH=0.08
  IRGM=1.0
  IRSPLIT=0.55
  IRSEND=(0.00, 1.00, 0.00)
  IRLSPLIT=0.685
  IRLVIOLET=(0.445, 0.00, 1.00)
  IRLEND=(0.75, 0.18, 0.87)

Green-only:
  IRRATIO=2.5
  IRS=1.0
  IRL=1.0
  IRSH=0.08
  IRGM=1.0
  IRSPLIT=0.55
  IRGSHORTEND=1.0
  IRGLONGMID=0.60
  IRGLONGEND=1.0
  IRGLONGSPLIT=0.80

Blue tail change vs v3 full map:
  v3 split=0.82, violet=(0.34,0.00,1.00), end=(0.50,0.06,0.92)
  v4 split=0.685, violet=(0.445,0.00,1.00), end=(0.75,0.18,0.87)

Artifacts:
  - actual binary outputs were generated with jpeg-v4 and the v4 wrappers
  - gradient outputs were generated from the same binary using GLO=0 GHI=0
