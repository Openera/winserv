## Building

### winserv.exe
From the top of the repository:

```
go build
```

If the generated packages in the `p` subdirectory can be regenerated using:

```
go generate
```

### launchserv
```
cd cmd/launchserv
go build
```

## Usage & Installation

`winserv.exe` contains a large number of unused standard package imports, primarily for their initialization code, and outputs to a logfile to indicate when different steps in the initialization and execution are reached. Itcan function as a standalone Windows service, or as a 'normal' executable launched by `launchserv.exe`. This latter method is preferred, as then `launchserv.exe` will log immediately before starting `winserv.exe`, giving an indication of the time delta between the process starting and the first `winserv.exe` log.

To use `launchserv.exe` and `winserv.exe` in this manner, place both executables together in a writable directory (such as `\Users\<user>\` or subdirectory) and use this command in an Administrator Command Prompt in that directory:

```launchserv.exe install```

Logs will be written to `output.txt` in the same directory, including `launchserv.exe` installing itself as a Windows Service. After this is done, Windows will start `launchserv.exe` as a Windows Service during system start-up, and it will in turn start `winserv.exe`. Both will write to the same log file (the last log from `launchserv.exe` is just before it starts `winserv.exe`).

To remove the Windows Service, in an Administrator Command Prompt:

```
launchserv.exe stop
launchserv.exe remove
```