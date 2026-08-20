# Logs and crash reports

When something goes wrong there is usually nothing to look at. A sync that quietly stops, a server that rejects a command, a window that vanishes on startup: without a record, a bug report is a description from memory.

Pelton can keep that record. Both parts are **off by default** and both stay on your machine.

## Nothing is uploaded

There is no crash reporter, no telemetry endpoint and no "send this to the developer" checkbox anywhere in this feature. A log is a file in your data folder. It reaches us only if you open it, read it and paste it into an issue yourself.

## What a log contains, and what it never contains

With logging on, Pelton writes what it is doing: which folders it synced, how many messages moved, which operations failed and why.

Kept out of it, always:

- **Passwords, app passwords and OAuth tokens.** Every secret Pelton holds is registered with the log writer, which removes it wherever it appears, including inside an error message a server sent back with the password quoted in it.
- **Mail content.** Bodies, snippets and attachment names never go in.
- **Subjects, senders and addresses**, unless you turn on the separate switch described below.

## Turning it on

**Settings, Privacy and network, Logs and crash reports.**

- **Write a log file** starts the log. **Detail** picks how much goes in it: errors only through to debug, which is the most useful for a bug report and the noisiest.
- **Include message subjects and senders** is a second, separate opt-in. It only matters for debugging sync, and it is separate because it is the one switch here that writes something about your mail to disk. It records the subject, sender and date of each message a sync stores. Bodies and attachments stay out either way.
- **Write a crash report** leaves a file behind if Pelton crashes, holding the stack, the version, your OS and what the app was doing. The next launch tells you it happened and offers to open it.

Nightly builds default both the log and crash reports on. A nightly already warns that it is untested, and a crash report from one is the point of running it.

### When the app will not start

The settings toggle is no help if you cannot get into settings. Two overrides turn logging on at debug level before anything else runs, and both win over the setting:

- the `PELTON_DEBUG` environment variable, set to anything
- the `--debug` command line flag

## Finding the files

**Settings, Privacy and network** shows the folder and its current size, with **Open log folder** next to it. The same button is in **Settings, About**, alongside **Copy diagnostics**, which puts the version, platform and settings a bug report needs on your clipboard. That summary contains no mail and no addresses.

The folder is `logs` inside Pelton's data directory:

- macOS: `~/Library/Application Support/Pelton/logs`
- Linux: `~/.config/Pelton/logs`
- Windows: `%AppData%\Pelton\logs`

## Size and deletion

The active log rolls over at 2 MB and three older copies are kept, so the folder tops out around 8 MB no matter how long logging stays on.

**Delete logs** removes every log and crash report. Deleting a mailbox also offers it, since a log written while that mailbox existed is still a record of using it.
