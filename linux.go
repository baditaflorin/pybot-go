package pybot

import "strings"

var linuxMarkers = []string{
	"linux", "ubuntu", "debian", "fedora", "arch linux", "red hat", "rhel", "centos", "rocky linux", "almalinux",
	"systemd", "systemctl", "journalctl", "apt-get", "apt ", "dnf", "pacman", "zypper", "sudo", "sshd", "ssh ",
	"chmod", "chown", "umask", "fstab", "grub", "kernel", "initramfs", "nftables", "firewalld", "ufw", "lvm", "ext4", "btrfs", "zfs",
	"uname", "lsblk", "findmnt", "journal", "crontab", "getent", "resolvectl", "ip route", "ss -", "df -", "du -", "ps aux", "ps -",
	"linux server", "disk space", "disk full", "disk usage", "large files", "filesystem", "file system", "file permissions", "permission denied", "listening port", "open ports", "process using", "process uses port", "package manager", "package updates", "shell command", "system service", "system logs", "network route", "network not working", "network interface", "no internet", "dns resolution", "high cpu", "cpu usage", "memory pressure", "memory usage", "ram usage", "ssh connection", "mount point", "firewall rules", "docker",
}

var nonLinuxMarkers = []string{"windows", "powershell", "macos", "mac os", "os x"}

// isLinuxQuestion recognizes explicit Linux distributions, tools, and common
// administration terms. It supplies a retrieval-domain hint, never an answer.
func isLinuxQuestion(text string) bool {
	lower := strings.ToLower(text)
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
	return false
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
