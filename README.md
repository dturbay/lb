# LoadBalancer
run tests with logs: go test -v -args  -stderrthreshold=INFO -v=5 -logtostderr=true


go gotchas:
 - never call flag.Parse() in tests!!! I think it's called in tests framework