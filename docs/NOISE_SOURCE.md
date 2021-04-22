# Quick reference (cont.)

- **Use it for**: https://github.com/yomorun/yomo
- **Where to file issues**: https://github.com/yomorun/yomo-source-mqtt-starter/issues
- **Source of this description**: https://github.com/yomorun/yomo-source-mqtt-starter
- **Related Projects**: 
  - https://hub.docker.com/r/yomorun/noise-source
  - https://hub.docker.com/r/yomorun/noise-web
  - https://hub.docker.com/r/yomorun/noise-emit



# What is noise-web?

The [noise-source](https://github.com/yomorun/yomo-source-mqtt-starter/blob/main/cmd/noise/main.go) project collects data from IoT devices, and send data to zipper engine.



# How to use this image



## run the container

```bash
docker run --rm --name noise-source -p 1883:1883 \
  -e YOMO_SOURCE_MQTT_ZIPPER_ADDR={YOUR-ZIPPER-ADDR}:9999 \
  -e YOMO_SOURCE_MQTT_SERVER_ADDR=0.0.0.0:1883 \
  yomorun/noise-source:latest
```

- YOUR-ZIPPER-ADDR: address of your zipper engine. 



# License

View [license information](https://github.com/yomorun/yomo/blob/master/LICENSE) for the software contained in this image.

As with all Docker images, these likely also contain other software which may be under other licenses (such as Bash, etc from the base distribution, along with any direct or indirect dependencies of the primary software being contained).

Some additional license information which was able to be auto-detected might be found in [`noise-source`](https://github.com/yomorun/yomo-source-mqtt-starter).

As for any pre-built image usage, it is the image user's responsibility to ensure that any use of this image complies with any relevant licenses for all software contained within.