# REST Server Performance

Compare simple REST server performance in Node.js and Go.

Both servers do the same - parse the JSON body and return it back in the response.
No external libraries are used, just the builtin APIs.

Install
- Node.js
- Go
- [wrk](https://github.com/wg/wrk)

The POST request is defined in [post.lua](post.lua).
The content is ~4KB JSON.

```
$ node -v
v10.1.0

$ go version
go version go1.10.1 linux/amd64

$ cat /etc/issue
Ubuntu 16.04.4 LTS \n \l
```
CPU: Intel® Core™ i7-7500U CPU @ 2.70GHz × 4

## Single thread

### Node.js
Start the server
```
$ node server.js
Listening on port 3000
```
Run the test in a separate console
```
$ wrk -d 20s -s post.lua http://localhost:3000
Running 20s test @ http://localhost:3000
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   793.29us  221.36us  11.96ms   92.88%
    Req/Sec     6.36k   513.96    12.33k    83.04%
  253629 requests in 20.10s, 564.55MB read
Requests/sec:  12618.86
Transfer/sec:     28.09MB
```
During the test the _node_ process runs at ~100% CPU and ~70MB memory.

### Go
Start the server
```
$ GOMAXPROCS=1 go run server.go
2018/05/19 23:45:39 Listening on port 3000
```
To get comparable results we start the go server with `GOMAXPROCS=1` as Node.js executes JavaScript in a single thread.

Run the test in a separate console
```
$ wrk -d 20s -s post.lua http://localhost:3000
Running 20s test @ http://localhost:3000
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.40ms  735.96us   7.03ms   59.42%
    Req/Sec     3.59k   249.95     7.36k    93.77%
  143198 requests in 20.10s, 318.06MB read
Requests/sec:   7124.41
Transfer/sec:     15.82MB
```
During the test the _server_ process runs at ~100% CPU and ~10MB memory.

The Go server parses the JSON request into fixed structures.
It is interesting to test what happens when we use unstructured JSON parsing with `interface{}`. So, open [server.go](server.go) and change this line
```
	var body Body
```
to
```
	var body interface{}
```
Run again the test
```
$ wrk -d 20s -s post.lua http://localhost:3000
Running 20s test @ http://localhost:3000
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     2.03ms    1.08ms  22.77ms   63.43%
    Req/Sec     2.48k   128.59     2.73k    74.00%
  98894 requests in 20.01s, 219.65MB read
Requests/sec:   4943.11
Transfer/sec:     10.98MB
```

### Results
Surprisingly Node.js is faster at HTTP and JSON handling. Probably the reason is that most of the work is done in the native (C++) parts of node and V8.

![chart](chart.png)

This is a fair comparison as both use almost the same amount of CPU (~100% ±3pp).

We see also that in Go structured JSON is faster.
You can find more details about this [here](https://github.com/dotchev/go-json-bench).

## Run in Docker
As an alternative here we run each server in a docker container limited to one CPU.
This way we ensure both servers have equal resources.

### Node
Build server image
```sh
docker build -t rest-node -f node.docker .
```
Start container and limit it to 1 CPU
```
$ docker run -it -p 3000:3000 --cpus=1 --rm rest-node
Node version: v10.1.0
CPU: 4 x Intel(R) Core(TM) i7-7500U CPU @ 2.70GHz
Listening on port 3000
```
Run the load test
```
$ wrk -d 20s -s post.lua http://localhost:3000
Running 20s test @ http://localhost:3000
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.31ms    1.96ms  59.60ms   97.28%
    Req/Sec     4.56k   785.46     5.30k    88.25%
  181603 requests in 20.00s, 404.23MB read
Requests/sec:   9078.68
Transfer/sec:     20.21MB
```

### Go
Here we test only the faster structured JSON handling so restore server.go
to its original version.

Build server image
```sh
docker build -t rest-go -f go.docker .
```
Start container and limit it to 1 CPU
```
$ docker run -it -p 3000:3000 --cpus=1 --rm rest-go
Version: go1.10.2
NumCPU: 4
GOMAXPROCS: 4
Listening on port 3000
```
Run the load test
```
$ wrk -d 20s -s post.lua http://localhost:3000
Running 20s test @ http://localhost:3000
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    25.13ms   25.78ms 179.49ms   77.94%
    Req/Sec   291.19     79.29     1.28k    95.50%
  11614 requests in 20.03s, 25.80MB read
Requests/sec:    579.77
Transfer/sec:      1.29MB
```
It seems running 4 threads on a single core is not a good idea.
Let's limit the threads to 1.
```
$ docker run -it -p 3000:3000 --cpus=1 -e "GOMAXPROCS=1" --rm rest-go
Version: go1.10.2
NumCPU: 4
GOMAXPROCS: 1
Listening on port 3000
```
Run again the load test
```
$ wrk -d 20s -s post.lua http://localhost:3000
Running 20s test @ http://localhost:3000
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.69ms    1.70ms  58.44ms   98.22%
    Req/Sec     3.12k   231.49     3.48k    86.57%
  124807 requests in 20.10s, 277.21MB read
Requests/sec:   6209.90
Transfer/sec:     13.79MB
```
### Results
|     | Node.js | Go (unlimited) | Go (GOMAXPROCS=1)
|-----|---------|----------------|------------------
|Req/s| 9078    | 579            | 6209

## Use all CPU cores
Here we use all the CPUs on the machine - no limits.

### Node.js
Start node cluster with one master process + one worker process per CPU
```
$ CLUSTER=1 node server.js
Node version: v10.1.0
CPU: 4 x Intel(R) Core(TM) i7-7500U CPU @ 2.70GHz
Master 23652 is running
Listening on port 3000
Worker 23658 started
Worker 23663 started
Worker 23664 started
Worker 23669 started
```
Run the load test
```
$ wrk -d 20s -s post.lua http://localhost:3000
Running 20s test @ http://localhost:3000
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   546.62us    0.97ms  44.91ms   97.55%
    Req/Sec    11.04k     1.62k   15.20k    84.50%
  439361 requests in 20.03s, 0.96GB read
Requests/sec:  21938.23
Transfer/sec:     48.83MB
```
Here 4 _node_ processes run at ~90% CPU each.

### Go
```
$ go run server.go
Version: go1.10.1
NumCPU: 4
GOMAXPROCS: 4
Listening on port 3000
```
```
$ wrk -d 20s -s post.lua http://localhost:3000
Running 20s test @ http://localhost:3000
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.09ms    1.58ms  26.92ms   93.73%
    Req/Sec     6.42k   495.44    12.33k    84.04%
  256084 requests in 20.10s, 568.79MB read
Requests/sec:  12740.96
Transfer/sec:     28.30MB
```
Here the _server_ process runs at ~360% CPU.

### Results
|     | Node.js x 4 | Go (unlimited) |
|-----|-------------|----------------|
|Req/s| 21938       | 12740          |

