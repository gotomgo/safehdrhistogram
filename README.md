# safehdrhistogram

The safehdrhistogram package is a wrapper around https://github.com/HdrHistogram/histogram-go that provides safe concurrent
access to a hdrhistogram.Histogram, or a collection of named instances of hdrhistogram.Histogram.

The wrapped histogram-go package is a functional version of HdrHistogram, but does not provide any support for 
concurrent access. While there are many use cases where you might not need (or want) concurrent access to a Histogram, 
there are many uses cases where having a concurrency safe implementation is a good thing.

In addition, the histogram-go package has some fairly minor flaws which are easy to correct without forking the 
original package and having to maintain those changes. Most of these flaws are around maintaining minor histogram state
(start time, end time, and tag) through the export/import process (typically referred to as snapshots), and to some 
degree, histogram resets. Again, these are minor flaws, and while I can say with confidence that I would update the 
original package to address these issues, others would likely argue for something different. I also consider the
existing support for exporting Percentiles to be fairly hokey, as it does not support structured output (unless 
parsing text is your thing.) I added exporting Percentile information in a structured way, which is ideal for 
interval based collection and reporting.

There are a variety of ways to create safe concurrent access to data structures in GO, channels is generally the best
option when maximum concurrency is desirable. Although a simple locking system could have been used (similar to other
atomic implementations), using channels seems very natural and idiomatic. In addition, I expect that channels is the
best way to maximize concurrent throughput for my use cases. Actual measurement to prove that conjecture will have to wait until I
have had more time to play-test the API in real world production scenarios.

## About HdrHistogram
A good summary of HdrHistogram can be found here: https://github.com/HdrHistogram/HdrHistogram

One of the many presentations by the creator of HdrHistogram, Gil Tene, called 'How NOT to Measure Latency'
can be found here: https://www.youtube.com/watch?v=lJ8ydIuPFeU&feature=youtu.be

A great blog post by Tyler Treat, 'Everything you know about latency is Wrong' which does an excellent job of 
summarizing a Gile Tene presentation, can be found here: 
https://bravenewgeek.com/everything-you-know-about-latency-is-wrong/

