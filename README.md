# smart  home
基于当前已有的硬件做一个尽量内网运行的智能家居系统

## 一、工具和原理
现有的硬件：
小爱智能音箱（某鱼50元)，绿米网关(某鱼60元)，小米人体传感器(某鱼20)
海康卡片摄像机CDX2HD（某宝28元）
ESP8266（某宝10元), MAX98357 D2A音频放大器(某宝7元), 腔体扬声器(某宝3元)
电脑

用到的软件(全部使用开源软件):
### 使用的开源软件列表。
- 项目(https://github.com/aler9/rtsp-simple-server),代理摄像头的rtsp流, 防止多路取流累死摄像机。
  
- 项目(https://github.com/AlexxIT/go2rtc), 因为rtsp-simple-server目前不支持webrtc, 需要用这个项目转一下。

- 项目(https://github.com/FFmpeg/FFmpeg), 把视频流转成一张张的照片, 用于人脸识别或者存储。

- 项目(https://www.home-assistant.io/), 用来管理小米设备并做mqtt与小米协议的一些转换。

- 项目(https://github.com/schmurtzm/MrDiy-Audio-Notifier), 用来把ESP8266和MAX98357做成一个MQTT播放器。

- 项目(https://github.com/fatedier/frp), 用来穿透内网, 把视频流发送到外网服务器, 方便远程查看。

- 项目(https://github.com/al-one/hass-xiaomi-miot), home-assistant的小米协议集成。

### 下面三个本程序的依赖库, 可以不用理会, 已经集成到本项目代码中。
- 项目(https://github.com/esimov/pigo), golang库, 图片人脸检测库, 性能好, 用来筛选有人脸的照片交给人脸识别程序。
- 项目(https://github.com/Kagami/go-face), golang库, 人脸识别程序, 识别出图片中的人脸是谁。
- 项目(http://github.com/lucacasonato/mqtt), golang库, mqtt客户端。


## 二、安装方法
### 1. 安装rtsp-simple-server, 代理摄像头的rtsp流。
- 下载https://github.com/aler9/rtsp-simple-server, 中的最新二制程序, 在配置文件rtsp-simple-server.yml中找到对应位置, 加入如下配置。
    ```bash
    paths:
      gate:
        source: rtsp://摄像头1的IP:554/streaming/channels/1  #摄像头的取流地址，海康摄像头是：rtsp://摄像头IP:554/streaming/channels/1, 内网建议设置摄像机使用UDP协议
        sourceOnDemand: yes #按需取流, 有访问时才去拉流, 减轻摄像机负担
      home:
        source: rtsp://摄像头2的IP:554/streaming/channels/1 #客厅还有一个摄像头
        sourceOnDemand: yes
    ```
- 运行命令开启视频流代理服务
  ```bash
  rtsp-simple-server  rtsp-simple-server.yml
  ```
    以后, 访问摄像头1将使用`rtsp://rtsp-simple-serverIP:8554/gate`， 可以设置密码等其他选项, 具体看rtsp-simple-server的配置说明。

### 2. 安装go2rtc把摄像机的视频流转为webRTC方便用浏览器查看
- 还是直接下载https://github.com/AlexxIT/go2rtc，对应操作系统的二进制文件即可，配置文件是：go2rtc.yaml
    ```bash
    streams:
      gate: rtsp://rtsp-simple-serverIP:8554/gate #从rtsp-simple-server代理那里取流
      home: rtsp://rtsp-simple-serverIP:8554/home
    rtsp:
      listen: ""  #这句是为了禁用rtsp发布模块，防止与rtsp-simple-server的8554端口冲突
    ```
- 运行命令启动go2rtc
    ```bash
    go2rtc  -config   go2rtc.yaml 
    ````
    以后就可以用浏览器访问`http://go2rtcIP:1984`来访问和查看视频流了。

### 3. 安装MQTT服务
- 使用的服务软件是mosquitto，软件里包括server和客户端程序。找到自己对应系统的安装程序安装就可以， 我使用的是archlinux系统， 安装方法是：
    ```bash
    pacman -Syu  mosquitto
    ```
- 修改mosquitto的配置文件。
配置文件在`/etc/mosquitto/mosquitto.conf`以及`/etc/mosquitto/conf.d/中以.conf结尾的文件`里，在/etc/mosquitto/conf.d目录下添加配置文件myconfig.conf 配置文件：
    ```bash
    #添加监听端口（很重要，否则只能本机访问）
    listener 1883
    #关闭匿名访问，客户端必须使用用户名密码
    allow_anonymous false
    #指定 用户名-密码 文件
    password_file /etc/mosquitto/pwfile.txt
    ```
- 添加账户及密码
    ```bash
    sudo mosquitto_passwd -c /etc/mosquitto/pwfile.txt lab37
    回车后连续输入2次用户密码即可
    ```

- 启动mosquitto
    ```bash
    sudo systemctl start  mosquitto
    ```

### 4. 安装home-assistant及小米协议集成插件与MQTT集成插件
- 到`https://www.home-assistant.io`， 查看安装home-assistant，选择适合的安装方式。因为我内网找了台老笔记本装了archlinux，所以我使用的是archlinux系统内的安装软件包方法。
    ```bash
    pacman  -Syu  home-assistant
    ```
- home-assistant启动配置好以后要在其中安装小米的控制集成`https://github.com/al-one/hass-xiaomi-miot`。 可以用hacs安装， 我用的是官方的安装方法：
    ```bash
    #进入到archlinux安装软件包home-assistant的配置目录(hass的家目录，这是一个链接目录)
    cd  /var/lib/hass
    #安装hass-xiaomi-miot集成
    wget -q -O - https://raw.githubusercontent.com/al-one/hass-xiaomi-miot/master/    install.sh | ARCHIVE_TAG=latest bash -
    ```
- 从home-assistant的集成管理中搜索MQTT，安装MQTT插件，配置连接上面建立的mqtt服务器， hass默认的订阅主题是 `homeassistant/#`。

### 5. 设置米家app中的自动化，当门口人体传感器感应到有人时打开绿米网关的灯并延迟10秒后关闭网关灯。
- 这一设置可以在门口有人时亮起灯光，做为提示作用。
- 这一设置可以使绿米网关在网络中发送组播消息， 实时性要优于home-assistant中的设备轮循。

### 6. 编写程序接收绿米网关亮灯时的组播信息并以mqtt的形式发送出去。
- 程序地址: https://github.com/lab37/lumi-gate-multicast2mqtt

### 7. 编写程序在收到有人到来的mqtt消息后开始识别来客人脸，并以mqtt的形式将识别结果发送出去。
- 程序地址：即本项目程序
- 配置文件：config.json
    ```bash
    {
        "imgFileName": "F:\\tools\\ffmpeg\\rtsp.jpg",  # 程序保存来客截图的位置        
        "mqttServer": "192.168.31.96:1883",   # mqtt服务器地址
        "mqttUserName": "lab37",   # mqtt用户名
        "mqttPassword": "142857",   # mqtt密码
        "mqttSubTopic": "homeassistant/security/gate/motion",   #订阅的主题，用于订阅门口有人时multicast2mqtt发送的mqtt主题
        "mqttPubTopic": "homeassistant/camera/facerec",       # 发布的主题， 用于在识别出人脸后发出的主题，附带的消息是一句文本，用于home-assistant接收后调用小爱同学的TTS播报出来。
        "ffmpegScriptFile": "F:\\tools\\ffmpeg\\start-ffmpeg-for-30s.bat",  # mqtt有客到来时启动的ffmpeg截图脚本(示例为本项目start-ffmpeg-for-30s.bat)。
        "faceFinder": "F:\\Program Files\\new-face-detect\\cascade\\facefinder",  #用于人脸检测的分类模型文件。
        "faceData": "F:\\Program Files\\new-face-detect\\face-data.json",  # 人脸数据库, 可由https://github.com/lab37/generate-face-128D-tools  生成
        "testDataDir": "F:\\Program Files\\new-face-detect\\testdata"  # 用于人脸识别的模型文件夹。github访问较慢也可从此处下载： https://www.aliyundrive.com/s/VQTwUeysrU3
    }
    ```
- 运行环境配置
    - windows下运行需要安装msys2，`https://www.msys2.org/`， 安装完成后
    ```bash
    #运行msys2 msys shell
    pacman   -Syu  #如果要求关闭shell那就关了， 重新打开再执行一遍这个命令。
    pacman   -S mingw-w64-x86_64-gcc mingw-w64-x86_64-dlib
    #把msys2的程序目录加入系统环境变量中的PATH。(我的是'F:\msys64\mingw64\bin')
    #安装ffmpeg，并把ffmepg的程序目录加入PATH。
    ```
   - linux下运行(以ubuntu为例)
    ```bash
    sudo apt update  && sudo apt upgrade -y
    sudo apt install ffmpeg  -y
    #安装go-face用到的库dlib依赖包
    sudo apt install libdlib-dev libblas-dev liblapack-dev libjpeg-turbo8-dev     libatlas-base-dev -y
    ```

- 运行命令，可从本项目release页面下载编译好的可执行文件，也可从https://www.aliyundrive.com/s/xU3oRwRhryB  下载windows下的可执行程序。
    ```bash
    new-face-detect   -config   config.json
    ```

- 如果要自己编译本程序，要求安装git和golang，golang版本至少要1.19，如果系统的golang不是最新的可以手动安装golang。

    - 对于windows直接下载安装程序安装即可，安装完后加入PATH，因为我们把程序直接安装在了windows里，不是msys2的环境中，所以要改一个wsys2的配置：msys2安装目录中的配置文件`mys2_shell.cmd`，去掉`set MSYS2_PATH_TYPE=inherit`这句前面的注释(rem)。
如果在msys2的环境中安装, 运行msys2 msys shell使用如下命令
        ```bash
        pacman -S mingw-w64-x86_64-go git
        ```
  - 对于linux系统，直接用下面的命令安装即可：
    ```bash
    sudo apt install   git   golang
    ```
    因为golang的版本要求1.18以上， 如果你系统中的golang版本不满足要求，可以用下面的方法安装新版golang。
    ```bash
    sudo apt remove golang #删除已经安装的， 
    apt autoremove  && apt autoclean
    #手动下载golang的安装包(网站：https://studygolang.com/dl)，
    wget  https://studygolang.com/dl/golang/go1.18.3.linux-amd64.tar.gz
    #解压安装
    tar xf go1.18.3.linux-amd64.tar.gz -C /usr/local
    sudo ln -sf /usr/local/go/bin/* /usr/bin/
    #把环境变量加到/etc/profile里
    vim /etc/profile #在最后面添加下面两行
    export GOPATH="$HOME/go"
    export PATH=$PATH:/usr/local/go/bin
    #执行命令
    source  /etc/profile
    sudo apt install gcc g++
    go  env  -w GOPROXY=https://goproxy.cn,direct
    mkdir  -p   ~/go/src
    ```


### 8. 设置home-assistant的自动化，调用小爱同学播报有人以及识别到的来客姓名
- 第一个自动化由multicast2mqtt有人来时发送的mqtt触发, 调用小爱同学的TTS, 播报门口有人来了。
    ```bash
    - id: '1673673951022'
      alias: 门口有人播报
      description: ''
      trigger:
      - platform: mqtt
        topic: ' homeassistant/security/gate/motion'
      condition:
      - condition: time
        after: 07:00:00
        before: '22:30:00'
        weekday:
        - sun
        - mon
        - tue
        - wed
        - thu
        - fri
        - sat
      action:
      - parallel:
        - service: mqtt.publish
          data:
            topic: speaker/play
            payload: http://192.168.31.96/men_kou_you_ren.mp3
        - service: xiaomi_miot.intelligent_speaker
          data:
            entity_id: media_player.xiaomi_s12_e10f_play_control
            text: 门口有人来了
      mode: single  
    ```
- 第二个自动化由人脸识别程序发送的mqtt触发, 调用小爱同学的TTS, 播报门口的人是谁。
    ```bash
    - id: '1673674153526'
      alias: 播报人脸识别到的人员
      description: ''
      trigger:
      - platform: mqtt
        topic: homeassistant\camera\facerec
      condition:
      - condition: time
        after: 07:00:00
        before: '22:30:00'
        weekday:
        - sun
        - mon
        - tue
        - wed
        - thu
        - fri
        - sat
      action:
      - service: xiaomi_miot.intelligent_speaker
        data:
          entity_id: media_player.xiaomi_s12_e10f_play_control
          text: '{{ trigger.payload }}'
      mode: single
    ```

## 三、一些说明
- 人脸数据库文件face-data.json解释一下：
    ```bash
    [{"name":"不认识","descriptor":[-0.086402895,0.062449184,0.025281772,-0.009759135,    -0.076491865,-0.064023325,-0.040512057,-0.157585205,0.137831153,-0.106937215,0.    24599106,-0.12431721,-0.191693885,-0.08147765,-0.11236507,0.255513065,-0.23072328,    -0.112731145,-0.03360492,-0.032514189,0.07166216,0.014985042,0.003708921,0.    088701943,-0.119844575,-0.324208085,-0.074863277,-0.118751262,-0.021223529,-0.    073263458,-0.001774842,0.13434302,-0.14869849,-0.041081177,0.040243549,0.    060301265,-8.71335E-05,-0.040631263,0.215542345,0.001988182,-0.22200241,0.    036800291,0.104752845,0.265269695,0.129051788,0.039235579,0.005628372,-0.    147551495,0.12083431,-0.14273316,-0.034898025,0.125377785,0.087508404,0.069152422,    0.021131133,-0.13263075,0.060253563,0.07951126,-0.188309025,0.022543276,0.    080808008,-0.115117245,0.009295596,-0.07170093,0.21518027,0.028408698,-0.    109116545,-0.131378645,0.10627839,-0.14796075,-0.10411008,0.043264372,-0.    110221635,-0.191448145,-0.36577347,-0.030245966,0.364821325,0.098646549,-0.    190443975,0.080678125,0.023957553,-0.053486413,0.130695993,0.19222275,-0.    022794363,0.019647024,-0.06424305,-0.027084,0.174070805,-0.043240131,-0.014647769,    0.21065591,-0.038219288,0.032520585,0.010140643,0.017078017,-0.110517103,0.    05279175,-0.085269262,-0.002775514,0.016622106,-0.02873386,0.011908661,0.    061259529,-0.15876859,0.112683435,-0.030069245,0.003977443,-0.020185009,0.    034421181,-0.10986702,-0.026641504,0.141332565,-0.242618605,0.195348715,0.    16779973,0.034253812,0.129698445,0.117613575,0.097841295,-0.05830558,0.00868199,    -0.185684965,-0.04591856,0.131664145,-0.064163616,0.116699835,-0.005825664]
    },
    {"name":"杨老师","descriptor":[-0.1159001,0.09263086,0.0035315836,-0.05554646,-0.    12823397,-0.06515299,-0.056935456,-0.16845298,0.17980391,-0.12318635,0.24090745,    -0.14977786,-0.17863044,-0.07701772,-0.16076702,0.31423423,-0.24079233,-0.    10512808,-0.02752084,-0.04456017,0.04039625,-0.0074459766,0.021002386,0.13501738,    -0.14683314,-0.3867429,-0.060152464,-0.12205729,-0.06866231,-0.06346952,0.    002346863,0.19863264,-0.15344058,-0.026065439,0.03514251,0.08698663,0.017542275,    -0.03569204,0.20831482,0.0054261843,-0.2556521,0.021006221,0.1502506,0.25756842,0.    116998866,0.019827325,0.03345234,-0.13459077,0.144613,-0.14897116,-0.07909307,0.    10195791,0.060774878,0.019279614,0.0580177,-0.13151145,0.042931125,0.09834204,-0.    18176971,-0.026620638,0.052940886,-0.09984687,-0.026971336,-0.12301793,0.24215722,    0.07947438,-0.11810785,-0.10671544,0.14092323,-0.14055648,-0.0759744,0.023722753,    -0.09704798,-0.20509142,-0.39267457,-0.040415365,0.3663631,0.14050837,-0.15795466,    0.09033715,0.037092745,-0.048454795,0.092696995,0.2019112,-0.046125486,0.    040454436,-0.03608166,-0.025072712,0.16190164,-0.029369472,0.017310027,0.21564318,    -0.047905046,0.049919203,-0.04239881,-0.0065827826,-0.16572133,0.058782343,-0.    094047084,0.021720782,-0.0022751018,-0.011840454,0.020022474,0.054937173,-0.    19528738,0.06903831,-0.04523987,-0.021006031,-0.041255984,0.084276654,-0.09607051,    -0.034084138,0.12846239,-0.24399525,0.16876882,0.17153835,0.011270084,0.15957515,    0.1532768,0.07538282,-0.042439796,-0.0006385939,-0.16411084,-0.050656885,0.    16210714,-0.12667571,0.10326011,-0.013752969]
    }]
    ```
    这个是人脸特征的数据文件，每个人的人脸特征数据由128个浮点数组成。我编写了一个程序用来根据人脸的照片生成这组数`https://github.com/lab37/generate-face-128D-tools`

## 四、扩展功能
### 1、将视频流转发到外网，方便外网查看摄像头。
- 当然可以直接使用海康的莹石云或者配合录像机的云功能查看
- 为了减少延时或者没有录像机和云功能的情况下，可以使用项目：https://github.com/fatedier/frp， 做内网穿透，把go2rtc的web管理页面转到公网服务器。
  1. 下载frp的二进制程序，里面有服务端frps程序和客户端程序frpc，
  2. 在服务器布署服务端程序frps，配置文件`frps.ini`内容如下：
     ```bash
     [common]
     bind_port = 7000  #服务端的监听端口，服务端程序通过这个端口和客户通信，用于客户端管理和数据交换。
     vhost_http_port = 8889  #在服务端开启http服务，这里设置http的端口。数据流：8889<-->7000<--->客户端连接时用的端口。
     authentication_method = token  #验证方式，出于安全考虑，要求连接到服务端的客户需要提供一个token凭证，否则不予连接。
     token = 827ccb0eea8a706c4c34a16891f84e7b #连接本服务端需要提供的token值。
     ```
     运行服务端程序
     ```bash
     sudo  frps -c frps.ini     
     ```
  3. 在内网布署客户端程序frpc，配置文件frpc.ini内容如下：
     ```bash
     [common]
     server_addr = 101.23.203.212 #frps服务端的ip地址
     server_port = 7000 #frps服务的管理端口
     authentication_method = token  #frps服务端的验证方式，要与服务端配置一致。
     token = 827ccb0eea8a706c4c34a16891f84e7b #连接服务端要提供的token，要与服务端配置一致。
     
     [web] #名字不重要，主要用[]符号来分隔不同区块，起个有意义的名字最好。
     type = http  #本区块的类型是http，用于转发http协议。
     local_port = 1984  #本地提供http服务的端口，这里是go2rtc的服务端口。数据流：本客户端连接frps的端口<---->1984
     custom_domains = www.abcde.cn  #向服务端注册自己的域名，以便服务端正确转发此域名的数据过来。
     http_user = admin  #本地http需要basic验证，这里是用户名
     http_pwd = P@ssw0rd #本地http需要basic验证，这里是密码
     ```
     运行服务端程序
     ```bash
     sudo  frpc -c frpc.ini 
     ```
     现在可以访问`http://外网IP:8889`来查看到go2rtc的管理页面了。
  4. 如果外网服务器安装了ngnix，可以用ngnix做反向代理来复用域名，但是go2rtc提供的webRTC用的是javascript，走的是websocket协议复用的http服务端口，如果在通过ngnix代理需要配置nginx代理websocket协议，配置如下：
      ```bash
      map $http_upgrade $connection_upgrade {
          default upgrade;
          ''   close;
      }
      
      
      server {
      
        listen 80;
        listen [::]:80;
        listen 443 ssl;      
        #nnnxxx文件夹中放此网站的静态文件，通常目录为/var/www/html/下
        root /var/www/html/nnnxxx;
        index  index.html;
        server_name www.abcde.cn abcde.cn;
        # /cctv/用于转发正常的http访问
        location  /cctv/  {
           proxy_pass http://127.0.0.1:8889/;
           proxy_redirect http://$host/ http://$http_host/;
           proxy_set_header  X-Real-IP  $remote_addr;
           proxy_set_header X-Forwarded- $proxy_add_x_forwarded_for;
           proxy_set_header Host $host;      
        }
        # /cctv/api/用来转发websocket流量
        location /cctv/api/ {
            proxy_pass http://127.0.0.1:8889/api/;
            proxy_read_timeout 300s;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection $connection_upgrade;
        } 
      }  
      ```
### 2. 制作一个mqtt协议控制的播放器，用于在无外网时播报。
- 可以使用ESP8266加MAX98357，结合项目：https://github.com/schmurtzm/MrDiy-Audio-Notifier， 做一个mp3播放器，目前只能播放mp3，这是一个播放器，没有TTS功能。
  1. 下载项目代码，固件在项目`_Schmurtz_ESP_Flasher`文件夹中，使用此文件下的`Schmurtz_ESP_Flasher.bat`批处理把对应的ESP8266固件刷入ESP8266开发板即可。
  2. 上电开发板，连接开发板的配置热点：`MrDIYNotifier`，密码：`mrdiy.ca`，打开浏览器登陆：`192.168.4.1`可以修改配置，包括修改登陆和管理密码、配置家里的wifi名字与连接密码、mqtt订阅的主题、 使用内部的I2C模拟DAC还是使用external DAC。这里选择使用external DAC也就是外部DAC。
  3. 修改完后保存，开发板会去连接家里的无线路由器，当它连接成功后会从路由器获取地址，并县且不会再发出管理热点，以后管理需要登陆它从路由器获取的地址。如果连接路由器失败他会回滚，继续发出管理热点。
  4. 断开开发板电源，根据项目main.cpp的文件中所描述的样子焊接MAX98357和扬声器。ESP8266与MAX98357的连接对应方式如下：
     ```bash
     +---------------------------------+-------------------------------+
     |         MAX98357                |            ESP8266            | 
     +---------------------------------+-------------------------------+
     | DAC - LRC                       | GPIO2  (D4)                   |
     | DAC - BCLK                      | GPIO15 (D8)                   |
     | DAC - DIN                       | GPIO3  (RX)                   |
     | DAC - GND                       | 8266   (GND)                  |
     | DAC - Vin                       | 8266   (3V)                   |     
     | speaker -  接扬声器负极          |                               |
     | speaker +  接扬声器正极          |                               |
     +---------------------------------+-------------------------------+
     ```
  5. 可用的控制主题如下：
     ```bash
      - 播放MP3                   MQTT 主题: "你配置的主题/play"
                                  MQTT 消息: http://url-to-the-mp3-file/file.mp3
                                  PS:  仅支持 HTTP , 不支持HTTPS. -> 可以参考:
                                  https://github.com/earlephilhower/ESP8266Audio/pull/410
      
      - 播放    AAC               MQTT 主题: "你配置的主题/aac" (支持ESP32, 不支持esp8266)
                                  MQTT 消息: http://url-to-the-aac-file/file.aac
      
      - 播放 an Icecast Stream    MQTT 主题: "你配置的主题/stream"
                                  MQTT 消息: http://url-to-the-icecast-stream/file.mp3
                                  example: http://icecast.radiofrance.fr/fiprock-midfi.mp3
      
      - 播放 a Flac               MQTT 主题: "你配置的主题/flac" (支持ESP32, 不支持esp8266)
                                  MQTT 消息: http://url-to-the-flac-file/file.flac
      
      - 播放                      MQTT 主题: "你配置的主题/tone"
                                  MQTT 消息: RTTTL formated text
                                  example: Soap:d=8,o=5,b=125:g,a,c6,p,a,4c6,4p,a,g,e,c,4p,4g,a
      
      - 停止播放                  MQTT 主题: "你配置的主题/stop"
      
      - 设置音量                  MQTT 主题: "你配置的主题/volume"
                                  MQTT 消息: 0.0到1.0
                                  example: 0.7
      							
      - 音量+/音量-                MQTT 主题: "你配置的主题/volume"
                                  MQTT 消息: + 或 -
                                  example: +  音量增加0.1，步进单位就是0.1
      							
      - TTS                       MQTT 主题: "你配置的主题/samvoice"
                                  MQTT 消息: 要读出来的文字
                                  example: 目前仅支持英文
      
      - Google TTS转语音          MQTT 主题: "你配置的主题/googlevoice"
                                  MQTT 消息: 要用google的TTS转语音的文字
                                  example: 目前用不了
     ```
  6. 在内网建立nginx服务器，用于提供http方式访问mp3的服务。把要播放的mp3放到nginx的web目录，想要播放时发送主题：配置的订阅主题/play即可。
   
## 五、一些Tips
### 1、一些优化
- 为了提高读写性能，最好用内存文件系统：
    ```bash
    mkdir /home/lab37/faceImg
    sudo mount -t tmpfs -o size=100M tmpfs /home/lab37/faceImg
    #系统重启后内存挂载的文件系统会消失，可以写入fstab长期挂载
    #在/etc/fstab文件中增加挂载配置，可以实现系统启动时自动挂载。具体如下：
    sudo vim /etc/fstab
    #在文件中增加如下内容并保存。
    tmpfs	/home/lab37/faceImg	tmpfs	defaults,size=100M	0 0   
     ```
### 2、用来测试的命令（备忘）
  - 系统命令
    ```bash
    #在windows下用ffmpeg播放RTSP视频流
    ffplay   "rtsp://192.168.31.225:554/ch0_1.h264"
     
    #每隔1秒截取一张图片并都按一定的规则命名来生成图片
    ffmpeg -i "rtsp://192.168.31.225:554/ch0_1.h264" -y -f image2 -r 1/1 /home/lab37/faceImg/img%03d.jpg
    
    #每隔1秒截取一张指定分辨率的图片并覆盖在同一张图片上
    ffmpeg -i "rtsp://192.168.31.225:554/ch0_1.h264" -y -f image2 -r 1/1  -update  1 -s 640x480 /home/lab37/     faceImg/rtsp.jpg
    
    #每隔1秒截取5张指定分辨率的图片并覆盖在同一张图片上
    ffmpeg -i "rtsp://192.168.31.225:554/ch0_1.h264" -y -f image2 -r 5/1 -update 1  /home/lab37/faceImg/rtsp.jpg
    
    #每隔1秒截取5张指定分辨率960*540的图片, 转成灰度图并覆盖在同一张图片上(-vf format=gray是转为灰度图)
    ffmpeg -i "rtsp://192.168.31.225:554/ch0_1.h264" -y -f image2 -r 5/1 -update 1   -s 960x540  -vf format=gray  /     home/lab37/faceImg/rtsp.jpg
  
    #只运行ffmpeg一定的时间，比如只运行30秒 -t 30
    ffmpeg -i "rtsp://192.168.31.225:554/ch0_1.h264" -y -f image2 -r 1/1 -t 30  /home/lab37/faceImg/img%03d.jpg
    
    
    #只输出错误到文件
    nohup command -c -b -d aaa.txt  > /dev/null 2 > log &
    
    #ffmpeg有时会异常退出, 需要监控ffmpeg运行, 编写脚本：ffmpeg2jpg.sh
    timeout 20 ffmpeg -i "rtsp://127.0.0.1:8554/gate" -y -f image2 -r 3/1 -update 1   -vf format=gray  /home/lab37/     faceImg/rtsp.jpg 2> /dev/null &
    
    #再编写一个监控ffmpeg的脚本, check_ff_mp_eg_live.sh
    #!/bin/sh 
    num=`ps -ef | grep ffmpeg | grep -v grep | wc -l`
    if [ $num -lt 1 ]
    then
     . /home/lab37/ffmpeg2jpg.sh
    fi
    #上面那个.不要落了，这是一个脚本调用另一个脚本的方法，或者用source. 因为脚本名字中有ffmpeg，所以要分开写,不然麻烦, 
    #把脚本添加crontab
    crontab -e 
    */1 * * * *  /home/lab37/check_ff_mp_eg_live.sh
    
    #Ubuntu默认没有开启cron定时任务的执行日志，需手动打开
    #编辑 rsyslog 配置文件，如果没有就新建一个
    sudo vim /etc/rsyslog.d/50-default.conf
    #取消 cron 注释，变成如下（如果没有此行配置就下入如下配置）
    cron.*          /var/log/cron.log
    #重启 rsyslog 服务
    sudo service rsyslog restart
    #然后执行crontab的任务，比如设置一个每分钟执行一次的，
    #过一分钟之后就可以看到生成了 /var/log/cron.log 文件
    #查看没有问题后最好关掉这个日志。
    ```
### 3、把程序注册为服务，以便开机自启动
  - Linux系统下使用systemd
    1. /etc/systemd/system/abcde.service的写法
        ```bash
        [Unit]
        Wants= network-online.target
        After = network.target  network-online.target  syslog.target
        
        [Service]
        Type=simple
        ExecStart=/home/lab37/user-systemd-services/aaa  -config  /home/lab37/user-systemd-services/config.json 
        Restart=always
        RestartSec=5
        StartLimitInterval=0
        RestartPreventExitStatus=SIGKILL
        
        [Install]
        WantedBy=multi-user.target
        ```
    2. 把程序和配置文件放到对应目录下，激活服务并启用服务
        ```bash
        #设置服务开机启动，
        sudo systemctl  enable  abcde
        #启动服务
        sudo systemctl  start   abcde
        ```
  - Windows下使用`https://github.com/winsw/winsw`， 把应用程序注册为服务。
    1. 下载https://github.com/winsw/winsw， 二进制程序`WinSW.NET461.exe`，把这个`WinSW.NET461.exe`程序复制到你的应用程序目录下面，改名为`lanuch_myapp.exe`，名字随便，这个程序就服务的启动程序，通过他来代理启动你自己的程序
    2. 在同目录下编写配置文件`lanuch_myapp.xml`
        ```bash
        <service>
          <id>lanuch_myapp服务的ID</id>
          <name>lanuch_myapp服务的名字</name>
          <description>This service discripe服务描述.</description>
          <env name="eviroment-var" value="%BASE%"/>
          <executable>F:\programe\myapp.exe可执行文件路径</executable>
          <arguments>-x -an --httpPort=8080命令的参数</arguments>
          <log mode="roll"></log>
        </service>
        ```
    3. 运行服务安装程序
       ```bash
       #在本目录执行下面的命令安装服务，安装完成后可以在系统的服务管理里看到
       lanuch_myapp.exe   install       
       #要卸载这个服务使用命令：
       lanuch_myapp.exe   uninstall
       #注意服务的运行是在系统的环境变量中，不是用户的环境变量，注意把用到的path加到系统环境变量的path中。
       ```