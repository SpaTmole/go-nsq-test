======
This is demo case of broken NSQD compression.


When using SNAPPY on very big data or Deflate (lvl3 compression) on relativelly small data- NSQD goes with `IO error - unexpected EOF` error.

###
Try yourself.

Simply clone this and run from your shell: `docker-compose up`

Play around with flags `USE_DEFLATE` and `USE_SNAPPY` in `main.go` to check different plots.

Note: NSQ Cannot use both Deflate and Sanppy options.
