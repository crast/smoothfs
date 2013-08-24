// Package smoothfs is the library upon which SmoothFS, an IO-smoothing filesystem, is based on.
// The purpose of SmoothFS is based on an observation with using high-latency
// filesystems in that the filesystem is not constrained by throughput but
// rather is using a small portion of its possible throughput due to the large
// latency on completing each IO request. SmoothFS uses caching and lookahead
// requests in order to do the smoothing out.
package smoothfs
