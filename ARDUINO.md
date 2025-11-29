# CCODE - Arduino

Clay as a buildsystem supports Arduino, ESP8266 and ESP32 boards. You can create a package that holds source code that can be 
compiled, linked and uploaded to an Arduino board.

- rcore; This is the core library that supports Arduino development and should always be a dependency.
- rwifi; This is an extension library that supports WiFi, Tcp and Udp on Arduino boards.
- rsensors; This is an extension library that supports specific sensors on Arduino boards.
- respnow; This is an extension library that supports ESP-NOW communication on Arduino boards.

Furthermore, clay provides some further helpful functionality:

- ./clay identify; This command will identify a USB connected Arduino board and list information like its MAC address, the
  information is written to a file called 'board_info.json' in the target/clay folder, this file is then used when calling 
  the flash command.
- ./clay flash; This command will build and upload an application to the connected Arduino board.
- ./clay list-boards; This command will list Arduino boards closely matching a given name

## Partition Table Configuration

Clay also will help you to configure the partition table of an Arduino board. You can specify some parameters in
your package to customize the partition table. For example, you can specify the following:

- OTA (true or false)
- APP size (xxKB or xxMB)
- SPIFFS size (xxKB or xxMB)
- NVS size (xxKB or xxMB)
- CoreDump (true or false)

When flashing the application, clay will load the `board_info.json` file to identify the connected board and
will generate a partition table based on the specified parameters. 
It will also identify the application and determine if SRAM, Flash and other resources are sufficient, Clay
will give warnings when resources are low or insufficient.