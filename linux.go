package pybot

import "strings"

var linuxMarkers = []string{
	"linux", "ubuntu", "debian", "fedora", "arch linux", "red hat", "rhel", "centos", "rocky linux", "almalinux",
	"systemd", "systemctl", "journalctl", "apt-get", "apt ", "dnf", "pacman", "zypper", "sudo", "sshd", "ssh ",
	"chmod", "chown", "umask", "fstab", "grub", "kernel", "initramfs", "nftables", "firewalld", "ufw", "lvm", "ext4", "btrfs", "zfs",
	"uname", "lsblk", "findmnt", "journal", "crontab", "getent", "resolvectl", "ip route", "ss -", "df -", "du -", "ps aux", "ps -",
	"linux server", "disk space", "disk full", "disk usage", "large files", "filesystem", "file system", "file permissions", "permission denied", "listening port", "open ports", "process using", "process uses port", "package manager", "package updates", "shell command", "system service", "system logs", "network route", "network not working", "network interface", "no internet", "dns resolution", "high cpu", "cpu usage", "memory pressure", "memory usage", "ram usage", "ssh connection", "mount point", "firewall rules", "docker",
	"no space left on device", "connection refused", "connection timed out", "name or service not known", "temporary failure in name resolution", "could not resolve hostname", "command not found", "failed to start", "service failed", "unit not found", "cannot mount", "read-only file system", "out of memory", "oom killed", "killed process", "too many open files", "server is slow", "server slow",
	"inode", "disk io", "disk i/o", "empty director", "recursive delete", "large directory", "external ip", "wayland", "x11", "timezone", "set date", "date through command line", "add x days to date", "cron job", "cron jobs", "ram slot", "cpu core", "cpuinfo", "inotify", "swappiness", "block device", "uuid", "partition label", "filesystem label", "/proc/", "/dev/", "/etc/", "/run/", "disk model", "hardware model", "virtualization technology", "file timestamps", "timestamp", "created time", "logrotate", "dmesg", "ps aux", "sftp", "ssh-add", "pty", "tty", "filesystem cache", "files ordered by size", "ordered by size", "file sizes", "year while listing", "show the year while listing", "hostname to an ip", "resolve a hostname to an ip", "resolve hostname to an ip", "distribution name", "os and version", "find how long ago", "vim", "su -", "sync operation", "sync two folders", "move all files", "symbolic link", "symlink", "init 6", "shutdown -r", "shared object", ".so version", "too many levels of symbolic links", "symlink loop", "oom killer", "high and low memory", "which gpu", "ls to show only file name", "filesystem type", "log-rotate", "join files", "diff output", "unified diff", "youtube-dl", "yt-dlp", "ethernet interface", "interface called", "enp0s", "detached process", "attach a terminal", "split files", "block network access of a process", "/usr/bin", "partition labels", "filesystem label", "filesystem labels", "high memory and low memory", "processes created by specific user", "block network access of a process", "desktop freezes", "video tearing", "freebsd over linux", "linux is a unix", "mono", "join files",
}

var nonLinuxMarkers = []string{"windows", "powershell", "macos", "mac os", "os x"}

// isLinuxQuestion recognizes explicit Linux distributions, tools, and common
// administration terms. It supplies a retrieval-domain hint, never an answer.
func isLinuxQuestion(text string) bool {
	lower := strings.ToLower(text)
	if isSimpleLinuxDefinitionQuestion(lower) {
		return false
	}
	explicitLinux := strings.Contains(lower, "linux") || strings.Contains(lower, "ubuntu") || strings.Contains(lower, "debian") || strings.Contains(lower, "fedora") || strings.Contains(lower, "arch linux")
	if !explicitLinux {
		for _, marker := range nonLinuxMarkers {
			if strings.Contains(lower, marker) {
				return false
			}
		}
	}
	for _, marker := range linuxMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return hasCommandToken(lower, "ls", "tar", "cp", "mv", "du", "df", "top", "ps", "find", "kill", "sync", "crontab", "dmesg", "nproc", "lscpu", "lsblk", "findmnt", "mount", "getent", "ss", "nft", "ufw", "lsof", "stat", "umask", "chown", "chmod", "uname", "sftp", "screen", "timedatectl", "journalctl", "systemctl")
}

func linuxCandidateDocumentIDs(query string) map[string]struct{} {
	text := strings.ToLower(query)
	candidates := make(map[string]struct{})
	add := func(ids ...string) {
		for _, id := range ids {
			candidates[id] = struct{}{}
		}
	}
	match := func(markers ...string) bool { return containsAny(text, markers...) }
	command := func(commands ...string) bool { return hasCommandToken(text, commands...) }

	if match("no space left on device", "disk space", "disk full", "disk usage", "filesystem space", "filesystem full", "inode") && !match("docker", "docker volume", "docker system df") && !match("/var/log", "journalctl --disk-usage", "journal disk usage") {
		add("linux-filesystem-space")
	}
	if match("deleted open file", "disk full but", "lsof +l", "lsof +l1", "space not freed") {
		add("linux-deleted-open-files")
	}
	if match("recover deleted file", "recover deleted files", "undelete a file", "undelete files") {
		add("linux-file-recovery")
	}
	if match("docker", "docker prune", "docker volume", "docker system df", "build cache") {
		add("linux-docker-storage")
	}
	if match("journalctl --disk-usage", "/var/log", "log storage", "log files", "journal disk usage", "size of the journal") && !match("logrotate", "log-rotate") {
		add("linux-log-storage")
	}
	if match("large file", "big file", "huge file", "largest file", "biggest file", "files over", "files larger than", "find large") {
		add("linux-find-large-files")
	}
	if (match("ls -", "file sizes", "list files ordered by size", "files ordered by size", "year while listing", "file names with spaces", "only file names", "show file size", "display file size") || command("ls")) && !match("show the year while listing", "year while listing files", "ls to show only file name, date, and size") {
		add("linux-ls")
	}
	if match("empty director", "delete empty director", "remove empty director", "recursively delete director") {
		add("linux-empty-directories")
	}
	if match("delete large directory", "delete large folder", "thousands of files", "efficiently delete", "remove large directory") {
		add("linux-large-dir-delete")
	}
	if match("systemd service", "systemctl", "unit failed", "service failed", "failed service", "failed to start", "will not start", "won't start", "stopped service", "nginx service", "service error") {
		add("linux-systemd-service-failure")
	}
	if match("journalctl", "journal logs", "journal entries", "boot logs", "previous boot", "kernel messages", "dmesg", "syslog") {
		add("linux-journal-logs")
	}
	if match("cpu flags", "flags in /proc/cpuinfo", "/proc/cpuinfo flags") {
		add("linux-cpu-flags")
	} else if match("cpu core", "number of cores", "core count", "cpuinfo", "/proc/cpuinfo", "lscpu", "nproc") {
		add("linux-cpu-info")
	}
	if match("wayland", "x11", "xorg", "display server", "desktop session type") {
		add("linux-display-session")
	}
	if match("oom killer", "oom-killer", "which process to kill", "process to kill first") {
		add("linux-oom")
	}
	if match("kill all screens", "kill all screen sessions", "screen sessions") {
		add("linux-screen-sessions")
	}
	if match("kill all a user's processes", "kill all a users processes", "kill all processes using their uid", "kill all user's processes") {
		add("linux-process-user-stop")
	} else if match("processes by user", "process by user", "processes owned by", "process created by", "process created by specific user", "specific user", "user's processes", "users processes", "processes using their uid") && match("process", "processes") {
		add("linux-process-user")
	}
	if match("kill -9", "sigkill", "sigterm", "kill a process", "should i kill", "process signal") || command("kill") && !match("oom killer", "kill all screens", "kill all a user's processes", "kill all a users processes", "kill all processes using their uid") {
		add("linux-process-signals")
	}
	if match("ps aux", "aux mean", "what does aux", "ps -") {
		add("linux-ps-aux")
	}
	if match("memory usage for a single process", "limit memory usage for a single process", "limit memory usage for a single linux process", "limit memory for a process", "memory limit for one process", "memory limit for a single") {
		add("linux-process-memory-limit")
	}
	if match("top meminfo", "meminfo in", "memory in megabytes in top", "display meminfo", "top memory units") {
		add("linux-top-memory")
	}
	if match("cpu usage", "high cpu", "cpu consuming", "cpu-consuming", "top cpu", "processes by memory", "sort processes", "process using cpu", "load average") || command("ps") || command("top") && match("top cpu", "processes by memory", "sort processes", "cpu consuming") {
		add("linux-process-cpu")
	}
	if match("drop_caches", "drop caches", "buffers and cache", "clear cache", "page cache", "file system caching", "filesystem caching", "empty the buffers") {
		add("linux-memory-cache")
	}
	if match("high and low memory", "high memory and low memory") {
		add("linux-high-low-memory")
	}
	if match("oom killer", "oom-killer") {
		add("linux-oom")
	}
	if match("swappiness") {
		add("linux-swappiness")
	}
	if match("memory pressure", "memory usage", "memory and swap", "ram usage", "swap usage", "out of memory", "high memory", "low memory", "memory limit", "free memory", "available memory") && !match("single process", "single linux process", "one process", "oom killer", "oom-killer", "swappiness", "high and low memory", "high memory and low memory") {
		add("linux-memory-pressure")
	}
	if match("ram slot", "ram slots", "memory slots", "slots in use") {
		add("linux-ram-slots")
	}
	if match("virtualization technology", "detect virtualization", "virtualized environment", "vm or container") {
		add("linux-virtualization")
	}
	if match("hardware model", "system model", "disk model", "ssd", "hdd", "rotational", "hardware information", "hard disks", "hard drives") {
		add("linux-hardware-info")
	}
	if match("disk io", "disk i/o", "monitor disk", "disk input output") {
		add("linux-disk-io")
	}
	if match("top command", "cpu fields", "cpu states", "us, sy", "wa, hi", "stolen time") {
		add("linux-top-fields")
	}
	if match("external ip", "public ip address", "external ip address") {
		add("linux-external-ip")
	}
	if match("listening port", "listening on port", "process using a port", "pid of the process using", "port 8080", "port 443", "port 22", "listening socket", "listening sockets", "socket is listening") {
		add("linux-listening-ports")
	}
	if match("127.0.0.53", "systemd-resolved stub") {
		add("linux-systemd-resolved-stub")
	} else if match("dns", "resolve hostname", "resolve a hostname", "hostname to an ip", "resolve a hostname to an ip", "name resolution", "resolv.conf", "resolvectl", "getent ahosts", "getent hosts") {
		add("linux-dns")
	}
	if match("network interface", "ethernet interface", "default route", "ip address", "external ip", "no internet", "network connectivity", "remote host reachable", "network not working") {
		add("linux-network-basics")
	}
	if match("default file permissions", "default permissions for all", "files and folders in a directory", "setfacl", "default acl") {
		add("linux-setfacl-defaults")
	} else if match("chmod 755", "mode 755", "numerical permission", "numeric permission", "permission bits", "read bit", "execute bit", "directory permissions", "chmod numerical", "chmod numeric") {
		add("linux-chmod-755", "linux-permissions")
	}
	if match("permission denied", "access denied", "file ownership", "chown", "owner of a file", "ownership") && !match("symbolic link", "symbolic links", "symlink", "running find", "skip permission denied", "ignore permission denied") {
		add("linux-permission-denied")
	}
	if match("default file permissions", "default permissions") && !match("all files", "all folders", "all folders/files", "files and folders", "directory tree", "setfacl") {
		add("linux-umask")
	}
	if match("umask") {
		add("linux-umask")
	}
	if match("package update", "update packages", "upgrade packages", "apt update", "apt upgrade", "dnf upgrade", "pacman -syu", "package manager") {
		add("linux-package-updates")
	}
	if match("compress a folder", "compress a directory", "tar.gz", "archive a folder") || command("tar") {
		add("linux-tar")
	}
	if match("cp -r", "cp -a", "difference between cp", "copy files", "copy a directory") || command("cp") {
		add("linux-copy")
	}
	if match("permission denied", "running find") && match("skip", "ignore", "suppress", "hide") || match("permission denied errors when running find") {
		add("linux-find-permission-errors")
	} else if match("find file", "find a file", "find files", "search for a file", "search /etc", "sshd_config", "running find in linux", "find command") || match("find /", "find .", "find ~", "`find") {
		add("linux-find-files")
	}
	if (match("mounted filesystem", "mounted filesystems", "mount point", "filesystem type", "filesystem types", "partition label", "partition labels", "uuid", "block device", "lsblk", "findmnt", "ext2", "ext3", "ext4", "hard disks", "hard drives") || command("lsblk", "findmnt", "mount")) && !match("mount cifs", "cifs network drive", "cifs mount") {
		add("linux-mounts", "linux-block-devices")
	}
	if match("sftp", "uploading director", "upload a directory") || command("sftp") {
		add("linux-sftp")
	} else if match("clear text password", "password as a parameter", "password on command line", "sshpass", "ssh password") {
		add("linux-ssh-password-safety")
	} else if match("error connecting to agent", "ssh-add returns", "ssh agent", "ssh-add") {
		add("linux-ssh-agent")
	} else if match("ssh connection", "ssh login", "ssh key") || command("ssh") {
		add("linux-ssh-debug")
	} else if match("what is ssh", "what does ssh do", "what is secure shell") {
		add("linux-ssh-overview")
	}
	if match("block network access of a process", "block network access for a process", "network access for a process", "per-process firewall") {
		add("linux-process-firewall")
	} else if match("block network access of a process", "block network access for a process", "network access for a process", "per-process firewall") {
		add("linux-process-firewall")
	} else if match("firewall", "ufw", "nftables", "nft list", "firewall rules", "block network access") {
		add("linux-firewall-status")
	}
	if match("cron job", "cron jobs", "scheduled cron", "crontab") || command("crontab") {
		add("linux-cron")
	}
	if match("inotify", "watch limit", "max_user_watches") {
		add("linux-inotify-limit")
	}
	if match("timezone", "time zone") || command("timedatectl") {
		add("linux-timezone")
	}
	if match("/proc/$pid/mem", "proc_pid_mem", "process memory file", "read from /proc") {
		add("linux-proc-pid-mem")
	}
	if match("set date", "date through command line", "set the linux system clock", "set system clock") {
		add("linux-date-command")
	}
	if match("add x days to date", "days to date and get new date", "add days to date", "gnu date") {
		add("linux-date-cli-arithmetic")
	}
	if match("tty", "pty", "pseudo terminal", "/dev/console", "/dev/tty", "/dev/tty0") {
		add("linux-tty")
	}
	if match("logrotate", "log-rotate", "logrotate configuration", "logrotate config", "make logrotate") {
		add("linux-logrotate")
	} else if match("rotate log", "log rotation", "journalctl --disk-usage", "/var/log", "log storage", "log files", "size of the journal") {
		add("linux-log-storage")
	}
	if match("distro", "distribution name", "os version", "os and version", "kernel version", "kernel architecture", "architecture of linux", "identify the linux distribution", "what version of linux") {
		add("linux-identify-system")
	}
	if match("vim", "edit multiple files in vim") {
		add("linux-vim-multiple-files")
	}
	if match("su -", "su vs", "login shell", "why do we use su") {
		add("linux-login-shell")
	}
	if match("watch the progress of a sync", "sync operation", "sync progress", "sync operation progress", "progress of a sync") || command("sync") && match("progress", "watch") {
		add("linux-sync-progress")
	}
	if match("predictable interface names", "enp0s", "interface called", "instead of eth0") {
		add("linux-interface-names")
	}
	if match("password for /etc/shadow", "generate password", "manually generate password") {
		add("linux-shadow-password")
	}
	if match("sync two folders", "sync two directories", "synchronize two directories", "synchronise two directories") {
		add("linux-rsync-folders")
	}
	if match("move all files", "move files and folders", "via mv", "mv command") {
		add("linux-move-files")
	}
	if match("symbolic links", "symbolic link", "symlink") && match("change ownership", "owner", "chown") {
		add("linux-symlink-ownership")
	}
	if match("/run/user", "runtime directory") {
		add("linux-runtime-dir")
	}
	if match("attach a terminal", "detached process", "attach to a detached", "reconnect to a process") {
		add("linux-detached-terminal")
	}
	if match("processes by user", "process by user", "processes owned by", "process created by", "processes using their uid", "specific user in unix", "processes of a user") && !match("kill all a user's processes", "kill all a users processes", "kill all processes using their uid") {
		add("linux-process-user")
	}
	if match("kill all screens", "kill all screen sessions", "screen sessions") {
		add("linux-screen-sessions")
	}
	if match("/bin and /usr/bin", "difference between /bin", "/usr/bin") {
		add("linux-bin-paths")
	}
	if match("difference reboot", "reboot , init 6", "init 6", "shutdown -r", "reboot versus") {
		add("linux-reboot-commands")
	}
	if match("how long ago", "when linux was installed", "when was linux installed", "installation date") {
		add("linux-install-time")
	}
	if match("virtualization technology", "detect virtualization", "virtualized environment", "vm or container") {
		add("linux-virtualization")
	}
	if match("set default kernel", "default kernel in grub", "grub default kernel", "grub menu entry") {
		add("linux-grub-default")
	}
	if match("portion of video", "download portion", "download sections", "youtube-dl", "yt-dlp") {
		add("linux-video-sections")
	}
	if match("diff output", "unified diff", "diff output", "understanding of diff") {
		add("linux-diff-output")
	}
	if match("tainted linux kernel", "tainted kernel", "kernel taint") {
		add("linux-kernel-taint")
	}
	if match("mount cifs", "cifs network drive", "cifs mount", "write permissions and chown") {
		add("linux-cifs-permissions")
	}
	if match("how /dev files are created", "device nodes", "/dev files", "udev", "dev files created", "are /dev") || match("/dev") && match("created", "how are") {
		add("linux-dev-devices")
	}
	if match("gpu", "graphics adapter", "graphics renderer", "active renderer") {
		add("linux-gpu")
	}
	if match("why is swappiness", "swappiness set to 60", "vm.swappiness") {
		add("linux-swappiness")
	}
	if match("shared object", ".so version", "soname", "so numbers") {
		add("linux-so-version")
	}
	if match("too many levels of symbolic links", "symlink loop", "symbolic link loop") {
		add("linux-symlink-loop")
	}
	if match("high memory and low memory", "high and low memory") {
		add("linux-high-low-memory")
	}
	if match("show the year while listing", "year while listing files", "ls to show only file name, date, and size") {
		add("linux-list-year-files")
	}
	if match("set default file permissions for all", "default file permissions for all", "default permissions for all folders") {
		add("linux-setfacl-defaults")
	}
	if match("join files again", "join files after splitting", "files again after splitting", "split files") {
		add("linux-join-files")
	}
	if match("mono", "how is mono magical") {
		add("linux-mono-runtime")
	}
	if match("find a manual", "man page", "manual page", "command line options", "--help", "option mean", "flag mean") || command("man") {
		add("linux-man-pages")
	}
	if match("what is linux", "is linux a unix", "linux a unix", "linux is a unix") {
		add("linux-overview")
	}
	if match("desktop freezes", "linux desktop freezes", "frozen graphical session") {
		add("linux-desktop-freeze")
	}
	if match("video tearing", "screen tearing", "tearing in linux") {
		add("linux-video-tearing")
	}
	if match("freebsd over linux", "choose freebsd", "choose FreeBSD") {
		add("linux-freebsd-comparison")
	}
	if match("file timestamps", "modification time", "created time", "timestamp of a file", "timestamp") {
		add("linux-file-metadata")
	}
	if match("get a username by id", "username by uid", "look up a username by id") {
		add("linux-user-lookup")
	}
	return candidates
}

func containsAny(text string, markers ...string) bool {
	for _, marker := range markers {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}

func hasCommandToken(text string, commands ...string) bool {
	for _, token := range strings.Fields(text) {
		token = strings.Trim(token, "`'\"()[]{}<>,.:;?!")
		for _, command := range commands {
			if token == command || strings.HasPrefix(token, command+"-") {
				return true
			}
		}
	}
	return false
}

func isSimpleLinuxDefinitionQuestion(text string) bool {
	normalized := strings.TrimSpace(strings.TrimRight(text, "?!.,"))
	switch normalized {
	case "what is linux", "what is ubuntu", "what is debian", "what is fedora", "what is arch linux":
		return true
	default:
		return false
	}
}

func linuxConceptDocumentID(text string) string {
	normalized := strings.TrimSpace(strings.TrimRight(strings.ToLower(text), "?!.,"))
	switch normalized {
	case "what is linux", "what is ubuntu", "what is debian", "what is fedora", "what is arch linux":
		return "linux-overview"
	case "what is systemd", "what does systemd do":
		return "linux-systemd-overview"
	case "what is ssh", "what does ssh do", "what is secure shell":
		return "linux-ssh-overview"
	default:
		return ""
	}
}

func isGenericLinuxHelpQuery(text string) bool {
	lower := strings.ToLower(text)
	asksForHelp := false
	for _, phrase := range []string{"help me", "can you help", "need help", "linux help", "what can you help", "what kind of help"} {
		if strings.Contains(lower, phrase) {
			asksForHelp = true
			break
		}
	}
	if !asksForHelp {
		return false
	}
	for _, topic := range []string{
		"disk", "space", "inode", "storage", "file", "systemd", "systemctl", "service", "journal", "log",
		"process", "cpu", "memory", "ram", "load", "port", "socket", "tcp", "udp", "dns", "resolve",
		"route", "gateway", "internet", "interface", "network", "chmod", "permission", "owner", "chown",
		"apt", "dnf", "pacman", "package", "update", "upgrade", "ssh", "key", "login", "authentication",
		"firewall", "ufw", "nft", "mount", "partition", "fstab", "lsblk", "docker", "volume", "container",
	} {
		if strings.Contains(lower, topic) {
			return false
		}
	}
	return true
}
