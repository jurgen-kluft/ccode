# TODO
 
- MEDIOR: Clay; add MsDev toolchain for supporting Windows.
- MINOR:  Clay; how to deal with response files, since they should be tracked by
          the dependency tracking system.
- MAJOR:  Clay; parse the boards.txt file and be able to fill in all the necessary
          vars in ArduinoEsp32 toolchain
- MINOR:  Clay; Executable Stats (size, .ram, .text, .data, .bss, etc.)
          - ESP32 Elf files
          - PE files

Future:

- MAJOR: Code coverage analysis
- 


# DONE

- MAJOR: Clay; To reduce compile/link time we need Dependency Tracking (Database)
         - source file <-> [command-line args], [object-file + header-files]
- MINOR: ESP32 S3 Target (could be automatic if we are able to fully parse the boards.txt file)
- MINOR: Clay build CLI-APP for the user
- MINOR: Clay build CLI-APP for the user
         build, clean, build-info, list libraries, list boards, list flash sizes (done)
         flash (done)
- MINOR: Clay; Build Info (build_info.h and build_info.cpp as a Library)
- MINOR: Clay; Archiver
- MINOR: Clay; Linker
- MINOR: Clay; Burner
