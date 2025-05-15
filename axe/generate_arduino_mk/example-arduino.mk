# Board definitions
CORE_DEBUG_LEVEL ?= 0
F_CPU ?= 240000000L
FLASH_MODE ?= dio
CDC_ON_BOOT ?= 
FLASH_SPEED ?= 40m
UPLOAD_RESET ?= --before default_reset --after hard_reset
UPLOAD_SPEED ?= 921600
COMP_WARNINGS ?= -w
MCU = esp32
INCLUDE_VARIANT = esp32
VTABLE_FLAGS?=-DVTABLES_IN_FLASH
MMU_FLAGS?=-DMMU_IRAM_SIZE=0x8000 -DMMU_ICACHE_SIZE=0x8000
SSL_FLAGS?=
BOOT_LOADER?=/Users/obnosis5/sdk/arduino/esp32/bootloaders/eboot/eboot.elf
# Commands
C_COM=$(C_COM_PREFIX) "/Users/obnosis5/sdk/arduino/esp32/tools/xtensa-esp-elf/bin/xtensa-esp32-elf-gcc"  -MMD -c "@/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/flags/c_flags" -w -Os -Werror=return-type -DF_CPU=$(F_CPU) -DARDUINO=10605 -DARDUINO_ESP32_DEV -DARDUINO_ARCH_$(UC_CHIP) -DARDUINO_BOARD=\"ESP32_DEV\" -DARDUINO_VARIANT=\"esp32\" -DARDUINO_PARTITION_default -DARDUINO_HOST_OS=\"$(OS)\" -DARDUINO_FQBN=\"generic\" -DESP32=ESP32 -DCORE_DEBUG_LEVEL=$(CORE_DEBUG_LEVEL)    -DARDUINO_USB_CDC_ON_BOOT=0  $(BUILD_EXTRA_FLAGS) "@/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/flags/defines" "-I$(dir $(SKETCH))" -iprefix "/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/include/" "@/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/flags/includes" "-I/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/dio_qspi/include" $(C_PRE_PROC_FLAGS) $(C_INCLUDES) "@$(BUILD_DIR)/build_opt.h" "@$(BUILD_DIR)/file_opts" 
CPP_COM=$(CPP_COM_PREFIX) "/Users/obnosis5/sdk/arduino/esp32/tools/xtensa-esp-elf/bin/xtensa-esp32-elf-g++"  -MMD -c "@/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/flags/cpp_flags" -w -Os -Werror=return-type -DF_CPU=$(F_CPU) -DARDUINO=10605 -DARDUINO_ESP32_DEV -DARDUINO_ARCH_$(UC_CHIP) -DARDUINO_BOARD=\"ESP32_DEV\" -DARDUINO_VARIANT=\"esp32\" -DARDUINO_PARTITION_default -DARDUINO_HOST_OS=\"$(OS)\" -DARDUINO_FQBN=\"generic\" -DESP32=ESP32 -DCORE_DEBUG_LEVEL=$(CORE_DEBUG_LEVEL)    -DARDUINO_USB_CDC_ON_BOOT=0  $(BUILD_EXTRA_FLAGS) "@/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/flags/defines" "-I$(dir $(SKETCH))" -iprefix "/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/include/" "@/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/flags/includes" "-I/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/dio_qspi/include" $(C_PRE_PROC_FLAGS) $(C_INCLUDES) "@$(BUILD_DIR)/build_opt.h" "@$(BUILD_DIR)/file_opts" 
S_COM="/Users/obnosis5/sdk/arduino/esp32/tools/xtensa-esp-elf/bin/xtensa-esp32-elf-gcc"  -MMD -c -x assembler-with-cpp "@/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/flags/S_flags" -w -Os -DF_CPU=$(F_CPU) -DARDUINO=10605 -DARDUINO_ESP32_DEV -DARDUINO_ARCH_$(UC_CHIP) -DARDUINO_BOARD=\"ESP32_DEV\" -DARDUINO_VARIANT=\"esp32\" -DARDUINO_PARTITION_default -DARDUINO_HOST_OS=\"$(OS)\" -DARDUINO_FQBN=\"generic\" -DESP32=ESP32 -DCORE_DEBUG_LEVEL=$(CORE_DEBUG_LEVEL)    -DARDUINO_USB_CDC_ON_BOOT=0  $(BUILD_EXTRA_FLAGS) "@/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/flags/defines" "-I$(dir $(SKETCH))" -iprefix "/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/include/" "@/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/flags/includes" "-I/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/dio_qspi/include" $(C_PRE_PROC_FLAGS) $(C_INCLUDES) "@$(BUILD_DIR)/build_opt.h" "@$(BUILD_DIR)/file_opts" 
LIB_COM="/Users/obnosis5/sdk/arduino/esp32/tools/xtensa-esp-elf/bin/xtensa-esp32-elf-gcc-ar" cr
CORE_LIB_COM="/Users/obnosis5/sdk/arduino/esp32/tools/xtensa-esp-elf/bin/xtensa-esp32-elf-gcc-ar" cr  "$(CORE_LIB)" 
LD_COM="/Users/obnosis5/sdk/arduino/esp32/tools/xtensa-esp-elf/bin/xtensa-esp32-elf-g++" "-Wl,--Map=$(BUILD_DIR)/$(MAIN_NAME).map" "-L/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/lib" "-L/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/ld" "-L/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/dio_qspi" "-Wl,--wrap=esp_panic_handler" "@/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/flags/ld_flags" "@/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/flags/ld_scripts"  -Wl,--start-group $^ $(BUILD_INFO_OBJ) "$(CORE_LIB)"   "@/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32/flags/ld_libs"  -Wl,--end-group -Wl,-EL -o "$(BUILD_DIR)/$(MAIN_NAME).elf"
PART_FILE?=/Users/obnosis5/sdk/arduino/esp32/tools/partitions/default.csv
GEN_PART_COM=python3 "/Users/obnosis5/sdk/arduino/esp32/tools/gen_esp32part.py" -q $(PART_FILE) "$(BUILD_DIR)/$(MAIN_NAME).partitions.bin"
OBJCOPY="$(dir $(ESPTOOL_FILE))/esptool" --chip esp32 elf2image --flash_mode "$(FLASH_MODE)" --flash_freq "$(FLASH_SPEED)" --flash_size "4MB" --elf-sha256-offset 0xb0 -o "$(BUILD_DIR)/$(MAIN_NAME).bin" "$(BUILD_DIR)/$(MAIN_NAME).elf"
SIZE_COM="/Users/obnosis5/sdk/arduino/esp32/tools/xtensa-esp-elf/bin/xtensa-esp32-elf-size" -A "$(BUILD_DIR)/$(MAIN_NAME).elf"
UPLOAD_COM?="$(dir $(ESPTOOL_FILE))/esptool"  --chip esp32 --port "$(UPLOAD_PORT)" --baud $(UPLOAD_SPEED)  --before default_reset --after hard_reset write_flash  -z --flash_mode keep --flash_freq keep --flash_size keep 0x1000 "$(BUILD_DIR)/$(MAIN_NAME).bootloader.bin" 0x8000 "$(BUILD_DIR)/$(MAIN_NAME).partitions.bin" 0xe000 "/Users/obnosis5/sdk/arduino/esp32/tools/partitions/boot_app0.bin" 0x10000 "$(BUILD_DIR)/$(MAIN_NAME).bin" 
COMMA=,
SPIFFS_SPEC:=$(subst $(COMMA), ,$(shell grep spiffs $(PART_FILE)))
SPIFFS_START:=$(word 4,$(SPIFFS_SPEC))
SPIFFS_SIZE:=$(word 5,$(SPIFFS_SPEC))
SPIFFS_BLOCK_SIZE?=4096
MK_FS_COM?="$(MK_FS_PATH)" -b $(SPIFFS_BLOCK_SIZE) -s $(SPIFFS_SIZE) -c $(FS_DIR) $(FS_IMAGE)
RESTORE_FS_COM?="$(MK_FS_PATH)" -b $(SPIFFS_BLOCK_SIZE) -s $(SPIFFS_SIZE) -u $(FS_RESTORE_DIR) $(FS_IMAGE)
FS_UPLOAD_COM?="$(dir $(ESPTOOL_FILE))/esptool"  --chip esp32 --port "$(UPLOAD_PORT)" --baud $(UPLOAD_SPEED)  --before default_reset --after hard_reset write_flash  -z --flash_mode keep --flash_freq keep --flash_size keep $(SPIFFS_START) $(FS_IMAGE)
PREBUILD=[ ! -f "$(dir $(SKETCH))"/partitions.csv ] || cp -f "$(dir $(SKETCH))"/partitions.csv "$(BUILD_DIR)"/partitions.csv && \
[ -f "$(BUILD_DIR)"/partitions.csv ] || [ ! -f "$(ESP_ROOT)/variants/esp32"/partitions.csv ] || cp "$(ESP_ROOT)/variants/esp32"/partitions.csv "$(BUILD_DIR)"/partitions.csv && \
[ -f "$(BUILD_DIR)"/partitions.csv ] || cp "/Users/obnosis5/sdk/arduino/esp32"/tools/partitions/default.csv "$(BUILD_DIR)"/partitions.csv && \
[ -f "$(dir $(SKETCH))"/bootloader.bin ] && cp -f "$(dir $(SKETCH))"/bootloader.bin "$(BUILD_DIR)"/$(MAIN_NAME).bootloader.bin || ( [ -f "$(ESP_ROOT)/variants/esp32"/bootloader.bin ] && cp "$(ESP_ROOT)/variants/esp32"/bootloader.bin "$(BUILD_DIR)"/$(MAIN_NAME).bootloader.bin || "$(dir $(ESPTOOL_FILE))"/esptool --chip esp32 elf2image --flash_mode $(FLASH_MODE) --flash_freq $(FLASH_SPEED) --flash_size 4MB -o "$(BUILD_DIR)"/$(MAIN_NAME).bootloader.bin "/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32"/bin/bootloader_dio_$(FLASH_SPEED).elf ) && \
[ ! -f "$(dir $(SKETCH))"/build_opt.h ] || cp -f "$(dir $(SKETCH))"/build_opt.h "$(BUILD_DIR)"/build_opt.h && \
[ -f "$(BUILD_DIR)"/build_opt.h ] || : > "$(BUILD_DIR)"/build_opt.h && \
: > '$(BUILD_DIR)/file_opts' && \
cp -f "/Users/obnosis5/sdk/arduino/esp32/tools/esp32-arduino-libs/esp32"/sdkconfig "$(BUILD_DIR)"/sdkconfig
PRELINK=
MEM_FLASH=^(?:\.iram0\.text|\.iram0\.vectors|\.dram0\.data|\.dram1\.data|\.flash\.text|\.flash\.rodata|\.flash\.appdesc|\.flash\.init_array|\.eh_frame|)\s+([0-9]+).*
MEM_RAM=^(?:\.dram0\.data|\.dram0\.bss|\.dram1\.data|\.dram1\.bss|\.noinit)\s+([0-9]+).*
FLASH_INFO=4MB (32Mb)
LWIP_INFO=
