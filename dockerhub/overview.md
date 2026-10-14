The image runs the fejkdata CLI: it reads a template on stdin and prints it rendered.

```sh
echo '{/sv_SE.person}' | docker run --rm -i larvit/fejkdata                # Sara Eriksson
echo '{/sv_SE.address}' | docker run --rm -i larvit/fejkdata --seed 42     # the same address every run
docker run --rm larvit/fejkdata --list                                     # every path the data offers
```

Tags: `latest`, and one per release, such as `0.1.0`. Seeded output can change between versions, so pin one where you compare it. In CI on shared runners, pull `ghcr.io/larvit/fejkdata` instead: Docker Hub limits anonymous pulls per IP address.

Documentation and source: https://github.com/larvit/fejkdata
