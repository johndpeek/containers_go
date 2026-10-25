package main

import (
	//"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func pivotRoot(newroot string) error {
	putold := filepath.Join(newroot, "/.pivot_root")
	fmt.Println(putold)
	fmt.Println(newroot)
	if err := syscall.Mount(newroot, newroot, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		fmt.Printf("Error on syscall mount bind")
		return err
	}
	if err := os.MkdirAll(putold, 0700); err != nil {
		fmt.Printf("error on mkdir putold 700")
		return err
	}
	if err := syscall.PivotRoot(newroot, putold); err != nil {
		fmt.Printf("error on pivot root new root put old")
		return err
	}
	if err := os.Chdir("/"); err != nil {
		fmt.Printf("error on chdir")
		return err
	}
	if err := syscall.Unmount("/.pivot_root", syscall.MNT_DETACH); err != nil {
		fmt.Printf("error on umount")
		return err
	}
	return nil
}

func main() {
	switch os.Args[1] {
	case "parent":
		parent()
	case "child":
		child()
	default:
		panic("help")
	}
}

// the parent function invoked from the main pgoram which setsup the needed namespaces
func parent() {
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	// statements below refer to the input and output and error streams of the processs created (cmd)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// set env variables
	cmd.Env = []string{"name=mycontainer"}

	// command below creats a UTS, PID and IPC, NETWORK and USERNAMESPACES for the process
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS |
			syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWUSER,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getgid(),
				Size:        1,
			},
		},
	}
	must(cmd.Run())
}

func child() {
	must(syscall.Sethostname([]byte("myhost")))
	//symlinked_command := symlink_check(os.Args[2])
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("Starting mount process")
	must(mountProc("/rfs"))
	must(syscall.Sethostname([]byte("myhost")))
	if err := pivotRoot("/rfs"); err != nil {
		fmt.Printf("Error running (child)pivot_root - %s\n", err)
		os.Exit(1)
	}
	// the command sets the hostname to myhost idea is to showcase the use of UTS namespace
	//this command executes the shell which is passed as a program argument
	//if err := os.Chdir("/proc"); err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}
	wd, _ := os.Getwd()
	fmt.Println("Working Directory:", wd)
	//fmt.Println(symlinked_command)

	//path := "/bin/busybox"
	//_, err := os.Stat(path)
	//switch {
	//switchase err == nil:
	//	fmt.Println(path, "exists")
	//case errors.Is(err, os.ErrNotExist):
	//	fmt.Println(path, "Does not exist")
	//default:
	//		fmt.Println(path, "stat failed:", err)
	//}
	must(cmd.Run())
}

func symlink_check(command_to_check string) string {
	symlinked_command, err := filepath.EvalSymlinks(command_to_check)
	if err != nil {
		fmt.Println(err)
		return (command_to_check)
	}
	return symlinked_command
}

func must(err error) {
	if err != nil {
		fmt.Printf("Error - %s\n", err)
	}
}

//function mounts the proc filesystem within the new mount namespace

func mountProc(newroot string) error {
	source := "proc"
	target := filepath.Join(newroot, "/proc")
	fstype := "proc"
	flags := 0
	data := ""
	// make a Mount system call to mount the proc filesystem within the mount namespace
	os.MkdirAll(target, 0755)
	if err := syscall.Mount(
		source,
		target,
		fstype,
		uintptr(flags),
		data,
	); err != nil {
		return err
	}
	return nil
}
