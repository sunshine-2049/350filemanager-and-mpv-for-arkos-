# MPV for arkos
1.we'd better do this in a aarch64 docker env（I'm not suggestion that you do these on your gameplayer）
```
# pull a ubuntu19.10 docker images
sudo docker pull ubuntu:19.10 
# run the docker images
sudo docker run -it --name test ubuntu:19.10 bash
```
2.config and install the packages
```shell
# change the apt source
sed -i 's|http://ports.ubuntu.com/ubuntu-ports|http://old-releases.ubuntu.com/ubuntu|g' /etc/apt/sources.list
aot update && apt install git python3  ffmpeg-dev libinput-dev libdrm-dev libsdl2-dev libmujs-dev  libavcodec-dev libavformat-dev libavutil-dev libswscale-dev libavfilter-dev libavdevice-dev liblua5.2-dev
ln -s /usr/bin/python3 /usr/bin/python
git clone https://github.com/mpv-player/mpv.git
cd mpv
git checkout release/0.31
```
3.build the mpv
```shell
./bootstrap.py
./waf configure   --enable-drm   --enable-sdl2  --enable-sdl2-gamepad   --enable-lua   --disable-x11   --disable-wayland   --disable-vulkan   --disable-gl   --disable-libmpv-shared   --prefix=/usr/local
./waf build
```
 4.test the mpv
```
# could copy the ./build/mpv file to your gameplayer and to retry:
sudo chmod +x ./mpv
# to test dynamic link is ok 
ldd ./mpv
# to test file play
./mpv -vo=drm --input xxx.mp4
# to test gamepad/joystick is ok 
./mpv --input-test  --input-gamepad --idle
```
