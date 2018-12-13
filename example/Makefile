jaeger:
	@docker run \
      --rm \
      --name jaeger \
      -p6831:6831/udp \
      -p16686:16686 \
      jaegertracing/all-in-one:1.6

postgres:
	@docker run -d -p 5432:5432 postgres