### How can i test it?

1. First you need to build the project, by running:

```SHELL
docker build -t go-stress-test . # The "go-stress-test" can be overwritted by any name
```

2. Then you just need to pass some value to the parameters, like:

```SHELL
docker run go-stress-test --url=https://www.google.com.br --concurrency=1 --requests=100
```