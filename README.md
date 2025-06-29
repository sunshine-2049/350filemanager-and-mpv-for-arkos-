# 1.Build MPV for arkos
`(by the way, if you use the binary file in the release, ignore the step 1-4)`<br/>
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
sudo chmod +x /opt/app/mpv
# to test dynamic link is ok 
ldd /opt/app/mpv
# to test file play
/opt/app/mpv -vo=drm --input xxx.mp4
# to test gamepad/joystick is ok 
/opt/app/mpv --input-test  --input-gamepad --idle
```
5.update the es_system.conf and restart the emulationstation
vi  /etc/emulationstation/es_systems.cfg
update the fllowing content:
```xml
 <system>
         <name>videos</name>
         <fullname>Movies</fullname>
         <path>/roms2/movies/</path>
         <extension>.mp4 .MP4 .avi .AVI .mpeg .MPEG .mkv .MKV .mov .MOV</extension>
         <command>sudo perfmax %GOVERNOR% %ROM%; /opt/app/mpv -vo=drm --input-gamepad %ROM%; sudo perfnorm</command>
         <platform>videos</platform>
         <theme>vhs</theme>
</system>
```
6.put the mpv.conf and input.conf to ~/.config/mpv

# 2.Use 350file to show medias and call the mpv program to play
1.such an example:
```shell
/opt/app/350file/350file \
  --conf='/opt/app/350file/config.json' \
  --start='/roms2/media/movies' \
  --cmds='/opt/app/mpv --input-gamepad --vo=drm  $__FILE__' \
  --filters='mkv,mov,flv,mp4,avi,wmv,ts,mpg,mpeg,3gp,rmvb'
```
2.cmd arguments
```
to --conf option, set the path of config file
to --start option, set the path of 350files up
to --cmds option, set the commands when the game controller click the B button
to --filters option, the 350 file only shows these in with the extendname string
```
3.buttons：
```
left/l1：the previous page 
right/r1: the next page
up: the previous file
down: the next file
b: to exec the command 
a: back to the path when 350file begining
```

Have func!
