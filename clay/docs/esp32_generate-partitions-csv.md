# Partitions CSV Generation

1. If buildPath/partitions.img not found, generate it:
    a. USB needs to be connected to an ESP32 device.
    b. "$(ESP_SDK)/esptool/esptool" read_flash 0x9000 0xc00 partitions.img
    c. 0x9000 is the default partition table address, and 0xc00 is the size of the partition table.
    d. older ESP32 devices may use 0x8000 as the partition table address.
2. "python3" $(ESP_SDK)/tools/gen_esp32part.py partitions.img > partitions.csv
