
# LoadBalancer

Learn go by practice

run tests with logs: go test -v -args  -stderrthreshold=INFO -v=5 -logtostderr=true

setup:
brew install dep
brew install graphviz


TODO/Plans:

* Introduce [dep](https://github.com/golang/dep)
* Implement protection from known attack:
  1. <https://www.haproxy.com/blog/use-a-load-balancer-as-a-first-row-of-defense-against-ddos/>
  2. <https://www.haproxy.com/blog/application-layer-ddos-attack-protection-with-haproxy/>

go gotchas:
 -never call flag.Parse() in tests!!! I think it's called in tests framework

[Markdown Cheatsheet](https://github.com/adam-p/markdown-here/wiki/Markdown-Cheatsheet)
