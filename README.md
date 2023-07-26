# xiaomi-camera-MJSXJ09CM-hack
we receive a video stream directly from the camera to the PC

The purpose of this project is to get the stream from the xiaomi MJSXJ09CM camera directly to the PC.

The essence of the problem: the xiaomi camera I bought suddenly turned out to be without rtsp support and it was possible to watch the video only from the mi-home application on the phone. Which is unacceptable.
The reverse engineering of the binary and the internals of the camera was carried out.

The code is written by a person who has never invaded IoT devices or dealt with video streams before.
Distributed AS IS, repeat steps at your own risk.

**TL;DR**

src directory has client.go and server.go

on the camera side, you need to get root access according to the instructions https://github.com/SungurLabs/sungurlabs.github.io/blob/6043366d497943e0a246a6a420ba8fb2adfcef31/_posts/2021-07-14-Xiaomi-Smart-Camera---Recovering-Firmware-and -Backdooring.md

I accessed by flashing the EEPROM with the addition of telnet from a third party busybox-armv7l.

We need to put inside the camera compiled golang for arm client.go with the clarification that we need exactly the statically assembled binary

go build -ldflags "-linkmode external -extldflags -static" client.go

I compiled natively on raspberry pi, but cross-compilation on amd64 is also possible.

the connection from the client to the server goes on UDP port 1053 and it is according to the precepts of UDP without confirmation of receipt of the packet.

inside camera run as ./client <your PC ip address>:<port>

on a PC - build server.go and run like this:

go build server.go

./server | ffmpeg -fflags nobuffer -flags low_delay -f hevc -i pipe:0 -map 0:v -f v4l2 /dev/video0

before that, to create /dev/loopX, you need to install the v4l2loopback kernel module on the PC, and you will also need ffmpeg

you can read the whole process of ordeals in more detail in the file full_log

the strace directory contains actual call traces from miio_record, miio_miss and fetch_av

I put the collected binaries of the corresponding architectures in the build directory

=======================================================================================================

Цель данного проекта - получить с камеры xiaomi MJSXJ09CM поток напрямую в ПК.

Суть проблемы: купленная мной сяоми-камера внезапно оказалась без поддержки rtsp и посмотреть видео можно было только с приложения mi-home на телефоне. Что неприемлимо.
Был проведен реверс-инжинеринг бинаря и внутренностей камеры.

Код написан человеком, который никогда ранее не вторгался в IoT-устройства и не имел дел с видеопотоками.
Распространяется AS IS, повторять действия на свой страх и риск.

**TL;DR**

в директории src есть client.go и server.go

на стороне камеры нужно получить root доступ по инструкции https://github.com/SungurLabs/sungurlabs.github.io/blob/6043366d497943e0a246a6a420ba8fb2adfcef31/_posts/2021-07-14-Xiaomi-Smart-Camera---Recovering-Firmware-and-Backdooring.md

Я получал доступ путем перепрошивки EEPROM с добавлением telnet из стороннего busybox-armv7l.

Нужно положить внутрь камеры скомпилированный golang-ом для arm client.go с уточнением о том, что нам нужен именно статически собранный бинарь

go build -ldflags "-linkmode external -extldflags -static" client.go

Я собирал нативно на raspberry pi, но кросс-компиляция на amd64 тоже возможна.

коннект от клиента к серверу идет по UDP порту 1053 и он по заветам UDP без подтверждения получения пакета.

внутри камеры запускать как ./client <your PC ip address>:<port>

на пк - собрать server.go и запустить вот таким образом:

go build server.go

./server | ffmpeg -fflags nobuffer -flags low_delay  -f hevc -i pipe:0 -map 0:v -f v4l2 /dev/video0

перед этим для создания /dev/loopX нужно установить в ПК модуль ядра v4l2loopback, а так же вам понадобится ffmpeg

более подробно весь процесс мытарств вы можете почитать в файле full_log


в директории strace находятся собственно трассировки вызовов от miio_record, miio_miss и fetch_av


собранные бинари соответствующих архитектур я положил в директорию build
