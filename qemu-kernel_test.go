// Copyright 2018 Mikhail Klementev. All rights reserved.
// Use of this source code is governed by a GPLv3 license
// (or later) that can be found in the LICENSE file.

package qemukernel

import (
	"net"
	"strings"
	"testing"
	"time"
)

func TestQemuSystemNew_InvalidKernelPath(t *testing.T) {
	kernel := Kernel{Name: "Invalid", KernelPath: "/invalid/path"}
	if _, err := NewQemuSystem(X86_64, kernel, "/bin/sh"); err == nil {
		t.Fatal(err)
	}
}

func TestQemuSystemNew_InvalidQemuArch(t *testing.T) {
	// FIXME put kernel image to path not just "any valid path"
	kernel := Kernel{Name: "Valid path", KernelPath: "/bin/sh"}
	if _, err := NewQemuSystem(unsupported, kernel, "/bin/sh"); err == nil {
		t.Fatal(err)
	}
}

func TestQemuSystemNew_InvalidQemuDrivePath(t *testing.T) {
	// FIXME put kernel image to path not just "any valid path"
	kernel := Kernel{Name: "Valid path", KernelPath: "/bin/sh"}
	if _, err := NewQemuSystem(X86_64, kernel, "/invalid/path"); err == nil {
		t.Fatal(err)
	}
}

func TestQemuSystemNew(t *testing.T) {
	// FIXME put kernel image to path not just "any valid path"
	kernel := Kernel{Name: "Valid path", KernelPath: "/bin/sh"}
	if _, err := NewQemuSystem(X86_64, kernel, "/bin/sh"); err != nil {
		t.Fatal(err)
	}
}

func TestQemuSystemStart(t *testing.T) {
	kernel := Kernel{Name: "Test kernel", KernelPath: testConfigVmlinuz}
	qemu, err := NewQemuSystem(X86_64, kernel, "/bin/sh")
	if err != nil {
		t.Fatal(err)
	}

	if err = qemu.Start(); err != nil {
		t.Fatal(err)
	}

	qemu.Stop()
}

func TestQemuSystemStart_Timeout(t *testing.T) {
	kernel := Kernel{Name: "Test kernel", KernelPath: testConfigVmlinuz}
	qemu, err := NewQemuSystem(X86_64, kernel, "/bin/sh")
	if err != nil {
		t.Fatal(err)
	}

	qemu.Timeout = time.Second

	if err = qemu.Start(); err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	if !qemu.Died {
		t.Fatal("qemu does not died :c")
	}

	if !qemu.KilledByTimeout {
		t.Fatal("qemu died not because of timeout O_o")
	}
}

func TestGetFreeAddrPort(t *testing.T) {
	addrPort := getFreeAddrPort()
	ln, err := net.Listen("tcp", addrPort)
	if err != nil {
		t.Fatal(err)
	}
	ln.Close()
}

func startTestQemu() (q *QemuSystem, err error) {
	kernel := Kernel{
		Name:       "Test kernel",
		KernelPath: testConfigVmlinuz,
		InitrdPath: testConfigInitrd,
	}
	q, err = NewQemuSystem(X86_64, kernel, testConfigRootfs)
	if err != nil {
		return
	}

	q.Timeout = 10 * time.Second

	if err = q.Start(); err != nil {
		return
	}

	return
}

func TestQemuSystemCommand(t *testing.T) {
	qemu, err := startTestQemu()
	if err != nil {
		t.Fatal(err)
	}
	defer qemu.Stop()

	output, err := qemu.Command("root", "cat /etc/shadow")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "root::") {
		t.Fatal("Wrong output from `cat /etc/shadow` by root")
	}

	output, err = qemu.Command("user", "cat /etc/passwd")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "root:x:0:0:root:/root:/bin/bash") {
		t.Fatal("Wrong output from `cat /etc/passwd` by user")
	}

	output, err = qemu.Command("user", "cat /etc/shadow")
	if err == nil { // unsucessful is good because user must not read /etc/shadow
		t.Fatal("User have rights for /etc/shadow. WAT?!")
	}
}
