
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