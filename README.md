# Secco [Poc]

trace syscalls of a specific container

## Prerequisites
* kinda [recent kernel version](https://github.com/iovisor/bcc/blob/master/docs/kernel-versions.md)
* kernel headers
* libbcc installed [see here](https://github.com/iovisor/bcc/blob/master/INSTALL.md)

## Running
```sh
# run in different shell
$ docker run -it alpine:3.10
[alpine] $

$ make binary
$ docker ps
CONTAINER ID [...]
79f589ed1d8c [...]
$ sudo ./bin/secco --id 79f589ed1d8c

# inside container
[alpine] $ apk add curl

# stop secco and see syscalls
map[access:3 arch_prctl:8 bind:1 brk:3234 chroot:2 close:381 connect:3 dup2:4 execve:9 exit_group:5 fallocate:3 fchdir:2 fchmod:1 fchownat:20 fcntl:41 flock:1 fork:5 fstat:161 fstatfs:1 getcwd:4 getdents64:13 geteuid:2 getpid:4 getppid:3 getsockname:1 gettid:5 getuid:5 ioctl:11 lseek:12 lstat:152 madvise:4 mkdirat:5 mmap:27 mprotect:20 munmap:7 newfstatat:98 open:329 openat:52 poll:53 read:1967 readlink:2 recvfrom:3 rename:1 renameat:61 rt_sigaction:35 rt_sigprocmask:23 rt_sigreturn:2 sendfile:149 sendto:26 set_tid_address:8 setpgid:2 setsockopt:10 socket:4 stat:9 statfs:1 symlink:456 symlinkat:2 umask:4 uname:3 unlinkat:22 utimensat:17 vfork:1 wait4:14 write:239 writev:57]
```
