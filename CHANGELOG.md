## 0.3.1 (2021-11-01)


### Bug Fixes

* **common**: resolver order (e872e8121e824d8c1c7cb73f74d0861767254525)
  > The weighting of flag values changed from `flags > config > env-vars` to
  > `flags > env-vars > config`. This means values from flags overwrite
  > values from env-vars which overwrite values from config.



## 0.3.0 (2021-08-11)


### New Features

* **common**: add Config.Variables to be added to kong.Vars (d5a8c1735b43793fae461e3f8d02ea09af040ad7)
* **common**: add the possibility to get version struct (9bee9cd01a950f0eed35421e2b18020d5c218847)



## 0.2.0 (2021-07-13)


### New Features

* **common**: add toml config support (e0f13a9717566f73f2504f62540f818fe110f9d0)



## 0.1.1 (2021-07-05)


### Dependencies

* **common**: update go dependencies (8b71edb0d5a280722e4547360d9d0afaf6e58346)



## 0.1.0 (2021-04-14)


### New Features

* **common**: `ShowConfig` flag shows config files which are not readable (c38e88689bd263313eb9bc0e1bf2807dcf05847e)



