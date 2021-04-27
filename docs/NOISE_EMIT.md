# Quick reference (cont.)

- **Use it for**: https://github.com/yomorun/yomo
- **Where to file issues**: https://github.com/yomorun/yomo-source-mqtt-starter/issues
- **Source of this description**: https://github.com/yomorun/yomo-source-mqtt-starter
- **Related Projects**: 
  - https://hub.docker.com/r/yomorun/noise-source
  - https://hub.docker.com/r/yomorun/noise-web
  - https://hub.docker.com/r/yomorun/noise-emit

# What is noise-emit?

If you don't have an MQTT device, you can simulate the data with [noise-emit](https://github.com/yomorun/yomo-source-mqtt-starter/tree/main/cmd/emitter), and send it to the [noise-source](https://github.com/yomorun/yomo-source-mqtt-starter/blob/main/cmd/noise/main.go) service.

# How to use this image

## run the container

```bash
docker run --rm --name noise-emit \
  -e YOMO_SOURCE_MQTT_BROKER_ADDR=tcp://localhost:1883 \
  yomo/noise-emit:latest
```

# License

View [license information](https://github.com/yomorun/yomo/blob/master/LICENSE) for the software contained in this image.

As with all Docker images, these likely also contain other software which may be under other licenses (such as Bash, etc from the base distribution, along with any direct or indirect dependencies of the primary software being contained).

Some additional license information which was able to be auto-detected might be found in [`noise-emit`](https://github.com/yomorun/yomo-source-mqtt-starter).

As for any pre-built image usage, it is the image user's responsibility to ensure that any use of this image complies with any relevant licenses for all software contained within.