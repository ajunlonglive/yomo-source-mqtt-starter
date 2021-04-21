# Quick reference (cont.)

- **Use it for**: https://github.com/yomorun/yomo
- **Where to file issues**: https://github.com/yomorun/yomo-sink-socket-io-example/issues
- **Source of this description**: https://github.com/yomorun/yomo-sink-socket-io-example
- **Related Projects**: https://hub.docker.com/r/yomorun/noise-source



# What is noise-web?

The [noise-source](https://hub.docker.com/r/yomorun/noise-source) project collects data from IoT devices, and the data is processed by flow on zipper and finally displayed through the `noise-web` project.



# How to use this image



## run the container

```bash
docker run --rm --name noise-web -p 3000:3000 yomorun/noise-web:latest
```

access the page via `http://{IP}:3000`



# License

View [license information](https://github.com/yomorun/yomo/blob/master/LICENSE) for the software contained in this image.

As with all Docker images, these likely also contain other software which may be under other licenses (such as Bash, etc from the base distribution, along with any direct or indirect dependencies of the primary software being contained).

Some additional license information which was able to be auto-detected might be found in [the `repo-info` repository's `quic-mqtt/` directory](https://github.com/yomorun/yomo-sink-socket-io-example).

As for any pre-built image usage, it is the image user's responsibility to ensure that any use of this image complies with any relevant licenses for all software contained within.