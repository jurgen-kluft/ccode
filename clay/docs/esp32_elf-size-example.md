# Example Output of ELF Size

// Example Output:
//     build/TestProject.elf  :
//     section                 size         addr
//     .rtc.text                  0   1074528256
//     .rtc.dummy                 0   1073217536
//     .rtc.force_fast           28   1073217536
//     .rtc_noinit                0   1342177792
//     .rtc.force_slow            0   1342177792
//     .rtc_fast_reserved        16   1073225712
//     .rtc_slow_reserved        24   1342185448
//     .iram0.vectors          1027   1074266112
//     .iram0.text            58615   1074267140
//     .dram0.data            15284   1073470304
//     .ext_ram_noinit            0   1065353216
//     .noinit                    0   1073485588
//     .ext_ram.bss               0   1065353216
//     .dram0.bss              5224   1073485592
//     .flash.appdesc           256   1061158944
//     .flash.rodata          69160   1061159200
//     .flash.text           135512   1074593824
//     .iram0.text_end            1   1074325755
//     .iram0.data                0   1074325756
//     .iram0.bss                 0   1074325756
//     .dram0.heap_start          0   1073490816
//     .debug_aranges         33920            0
//     .debug_info          2191837            0
//     .debug_abbrev         281430            0
//     .debug_line          1342862            0
//     .debug_frame           85216            0
//     .debug_str            361642            0
//     .debug_loc            438423            0
//     .debug_ranges          67832            0
//     .debug_line_str        12428            0
//     .debug_loclists       113443            0
//     .debug_rnglists        10220            0
//     .comment                 169            0
//     .xtensa.info              56            0
//     .xt.prop                2220            0
//     .xt.lit                   80            0
//     Total                5226925
