iso8601duration
===============

ISO8601 Duration Parser for Golang

Provides a partial implementation of ISO8601 durations. Parsing doesn't take into consideration variable lengths
of days (24 vs 25), months (28 vs 31) nor years (365 vs 366). It will default to the following:

* 1 Day = 24 Hours
* 1 Month = 30 Days
* 1 Year = 365 Days
 
Also, converting back to string will omit months and work with days and weeks, e.g.

```
INPUT     OUTPUT
P1M       P30D
P1M5D     P5W
P1Y1M5D   P1Y35D
```

Zero value is set to `P0D`

Forked from https://github.com/ChannelMeter/iso8601duration (refers to http://github.com/BrianHicks/finch)
Updated from https://github.com/retzkek/iso8601duration
