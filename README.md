
# Transport-level LoadBalancer

Learn go by practice

setup:

* brew install dep
* brew install graphviz

TODO/Plans:

* Implement protection from known attack:
  1. <https://www.haproxy.com/blog/use-a-load-balancer-as-a-first-row-of-defense-against-ddos/>
  2. <https://www.haproxy.com/blog/application-layer-ddos-attack-protection-with-haproxy/>
  3. Application-level support?

go gotchas:
 -never call flag.Parse() in tests!!! I think it's called in tests framework

* [Markdown Cheatsheet](https://github.com/adam-p/markdown-here/wiki/Markdown-Cheatsheet)
* [Nice refresher on socket options](https://stackoverflow.com/questions/14388706/socket-options-so-reuseaddr-and-so-reuseport-how-do-they-differ-do-they-mean-t)
* [The complete guide to Go net/http timeouts](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/)
* [go amazing profiler](https://blog.golang.org/profiling-go-programs)
* [dependencies management tool](https://golang.github.io/dep/docs/introduction.html)
* show deps: dep status -dot | dot -T png | open -f -a /Applications/Preview.app
* run tests with logs: go test -v -args  -stderrthreshold=INFO -v=5 -logtostderr=true
* go test -v -args  -logtostderr=true -stderrthreshold=INFO -test.run=TestLB_With_ab # -test.cpuprofile=./cpuprofile -test.memprofile=./memprofile -v=3

performanse: ab tool results

<table>
<tr>
<th> direct call </th>  <th> call via LB </th>
</tr>
<tr>
<td>
<pre>
This is ApacheBench, Version 2.3 <$Revision: 1826891 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking localhost (be patient)


Server Software:
Server Hostname:        localhost
Server Port:            58814

Document Path:          /
Document Length:        100 bytes

Concurrency Level:      100
Time taken for tests:   1.912 seconds
Complete requests:      5000
Failed requests:        0
Total transferred:      1085000 bytes
HTML transferred:       500000 bytes
Requests per second:    2614.43 [#/sec] (mean)
Time per request:       38.249 [ms] (mean)
Time per request:       0.382 [ms] (mean, across all concurrent requests)
Transfer rate:          554.03 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0   18  17.5     16     139
Processing:     1   20  15.4     17     140
Waiting:        0   14  13.2     12     121
Total:          7   38  24.7     34     160

Percentage of the requests served within a certain time (ms)
  50%     34
  66%     39
  75%     41
  80%     42
  90%     49
  95%     56
  98%    147
  99%    152
 100%    160 (longest request) </pre> </td>
<td> <pre>This is ApacheBench, Version 2.3 <$Revision: 1826891 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking localhost (be patient)


Server Software:
Server Hostname:        localhost
Server Port:            58815

Document Path:          /
Document Length:        100 bytes

Concurrency Level:      100
Time taken for tests:   6.284 seconds
Complete requests:      5000
Failed requests:        0
Total transferred:      1084985 bytes
HTML transferred:       500000 bytes
Requests per second:    795.70 [#/sec] (mean)
Time per request:       125.676 [ms] (mean)
Time per request:       1.257 [ms] (mean, across all concurrent requests)
Transfer rate:          168.62 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0   45  21.8     43     123
Processing:     7   78  34.1     72     166
Waiting:        2   66  36.6     59     161
Total:         23  124  37.6    119     225

Percentage of the requests served within a certain time (ms)
  50%    119
  66%    140
  75%    155
  80%    162
  90%    177
  95%    187
  98%    194
  99%    199
 100%    225 (longest request) </pre> </td>
 <tr>
</table>