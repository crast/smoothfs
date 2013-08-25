/*

Package smoothfs is a library for an IO-smoothing FUSE filesystem.

The reason SmoothFS was created is that it was observed that many high-latency
filesystems (such as network filesystems over the internet, and even some
specific applications of spinning disks, like CD-ROM) may have rather good
throughput and yet completely suck when used by certain applications (like
for example playing a movie.) This is because these applications issue small
reads spaced apart in time, and the fact of the matter is these kind of
filesystems have a very high latency of responding to commands. SmoothFS
acts as a wrapper "masking" these filesystems and should provide better
overall responsiveness from a user perspective.

The mechanism of the "smoothing" operation is instrumented via several
techniques:
	* Caching reads: Very often applications will request the same byte
	  range multiple times. SmoothFS caches these bytes to avoid more
	  network round-trips.
	* Larger block size reads: Many applications request e.g. 4kb at a time,
	  SmoothFS increases this size and tries to do byte-aligned reads as well.
	* Predictive reads (coming soon): Will try to prefetch data from the
	  backing filesystem such that the data is on hand before the application
	  requests it.


Current Status

SmoothFS is currently working, though under pretty heavy development. It can
be used right now as a read-only fronting filesystem and actually does speed
up high-latency filesystems such as SSHFS, though it will grow unboundedly with
memory usage as more files are read, so it's definitely not in a production
state at this point.

Planned / Coming Soon:
	* Tracking / clearing objects in memory (such as cached blocks)
	* On-disk cache
	* Pluggable predictive read engine (potentially algorithms per file type)
*/
package smoothfs
