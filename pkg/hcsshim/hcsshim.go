// +build windows

// Shim for Windows Containers Host Compute Service (HSC)

package hcsshim

import (
	"syscall"
	"unsafe"

	log "github.com/Sirupsen/logrus"
)

var (
	modvmcompute = syscall.NewLazyDLL("vmcompute.dll")

	procCreateComputeSystem             = modvmcompute.NewProc("CreateComputeSystem")
	procStartComputeSystem              = modvmcompute.NewProc("StartComputeSystem")
	procCreateProcessInComputeSystem    = modvmcompute.NewProc("CreateProcessInComputeSystem")
	procWaitForProcessInComputeSystem   = modvmcompute.NewProc("WaitForProcessInComputeSystem")
	procShutdownComputeSystem           = modvmcompute.NewProc("ShutdownComputeSystem")
	procTerminateProcessInComputeSystem = modvmcompute.NewProc("TerminateProcessInComputeSystem")
	procResizeTTY                       = modvmcompute.NewProc("ResizeTTY")
)

// The redirection devices as passed in from callers
type Devices struct {
	StdInPipe  string
	StdOutPipe string
	StdErrPipe string
}

// The redirection devices as passed used internally
type deviceInt struct {
	stdinpipe  *uint16
	stdoutpipe *uint16
	stderrpipe *uint16
}

// Configuration will be JSON such as:

//configuration := `{` + "\n"
//configuration += ` "SystemType" : "Container",` + "\n"
//configuration += ` "Name" : "test2",` + "\n"
//configuration += ` "RootDevicePath" : "C:\\Containers\\test",` + "\n"
//configuration += ` "IsDummy" : true` + "\n"
//configuration += `}` + "\n"

// Note that RootDevicePath MUST use \\ not \ as path separator

func CreateComputeSystem(ID string, Configuration string) error {

	log.Debugln("hcsshim::CreateComputeSystem")
	log.Debugln("ID:", ID)
	log.Debugln("Configuration:", Configuration)

	// Convert ID to uint16 pointers for calling the procedure
	IDp, err := syscall.UTF16PtrFromString(ID)
	if err != nil {
		log.Debugln("Failed conversion of ID to pointer ", err)
		return err
	}

	// Convert Configuration to uint16 pointers for calling the procedure
	Configurationp, err := syscall.UTF16PtrFromString(Configuration)
	if err != nil {
		log.Debugln("Failed conversion of Configuration to pointer ", err)
		return err
	}

	// Call the procedure itself.
	r1, _, _ := procCreateComputeSystem.Call(
		uintptr(unsafe.Pointer(IDp)), uintptr(unsafe.Pointer(Configurationp)))

	use(unsafe.Pointer(IDp))
	use(unsafe.Pointer(Configurationp))

	if r1 != 0 {
		return syscall.Errno(r1)
	}

	return nil
} // CreateComputeSystem

func StartComputeSystem(ID string) error {

	log.Debugln("hcsshim::StartComputeSystem")
	log.Debugln("ID:", ID)

	// Convert ID to uint16 pointers for calling the procedure
	IDp, err := syscall.UTF16PtrFromString(ID)
	if err != nil {
		log.Debugln("Failed conversion of ID to pointer ", err)
		return err
	}

	// Call the procedure itself.
	r1, _, _ := procStartComputeSystem.Call(uintptr(unsafe.Pointer(IDp)))

	use(unsafe.Pointer(IDp))

	if r1 != 0 {
		return syscall.Errno(r1)
	}

	return nil
} // StartComputeSystem

func CreateProcessInComputeSystem(ID string,
	ApplicationName string,
	CommandLine string,
	WorkingDir string,
	StdDevices Devices,
	EmulateTTY uint32) (PID uint32, err error) {

	log.Debugln("hcsshim::CreateProcessInComputeSystem")
	log.Debugln("ID:", ID)
	log.Debugln("CommandLine:", CommandLine)

	// Convert ID to uint16 pointer for calling the procedure
	IDp, err := syscall.UTF16PtrFromString(ID)
	if err != nil {
		log.Debugln("Failed conversion of ID to pointer ", err)
		return 0, err
	}

	// Convert ApplicationName to uint16 pointer for calling the procedure
	ApplicationNamep, err := syscall.UTF16PtrFromString(ApplicationName)
	if err != nil {
		log.Debugln("Failed conversion of ApplicationName to pointer ", err)
		return 0, err
	}

	// Convert CommandLine to uint16 pointer for calling the procedure
	CommandLinep, err := syscall.UTF16PtrFromString(CommandLine)
	if err != nil {
		log.Debugln("Failed conversion of CommandLine to pointer ", err)
		return 0, err
	}

	// Convert WorkingDir to uint16 pointer for calling the procedure
	WorkingDirp, err := syscall.UTF16PtrFromString(WorkingDir)
	if err != nil {
		log.Debugln("Failed conversion of WorkingDir to pointer ", err)
		return 0, err
	}

	// Need an instance of the redirection devices for internal use when calling the procedure
	var (
		stdinpipe  *uint16
		stdoutpipe *uint16
		stderrpipe *uint16
	)

	// Convert stdin, if supplied to uint16 pointer for calling the procedure
	if len(StdDevices.StdInPipe) != 0 {
		stdinpipe, err = syscall.UTF16PtrFromString(StdDevices.StdInPipe)
		if err != nil {
			log.Debugln("Failed conversion of StdInPipe to pointer ", err)
			return 0, err
		}
	}

	// Convert stdout, if supplied to uint16 pointer for calling the procedure
	if len(StdDevices.StdOutPipe) != 0 {
		stdoutpipe, err = syscall.UTF16PtrFromString(StdDevices.StdOutPipe)
		if err != nil {
			log.Debugln("Failed conversion of StdOutPipe to pointer ", err)
			return 0, err
		}
	}

	// Convert stderr, if supplied to uint16 pointer for calling the procedure
	if len(StdDevices.StdErrPipe) != 0 {
		stderrpipe, err = syscall.UTF16PtrFromString(StdDevices.StdErrPipe)
		if err != nil {
			log.Debugln("Failed conversion of StdErrPipe to pointer ", err)
			return 0, err
		}
	}

	internalDevices := &deviceInt{
		stdinpipe:  stdinpipe,
		stdoutpipe: stdoutpipe,
		stderrpipe: stderrpipe,
	}

	// To get a POINTER to the PID
	pid := new(uint32)

	log.Debugln("Calling the procedure itself")

	// Call the procedure itself.
	r1, _, _ := procCreateProcessInComputeSystem.Call(
		uintptr(unsafe.Pointer(IDp)),
		uintptr(unsafe.Pointer(ApplicationNamep)),
		uintptr(unsafe.Pointer(CommandLinep)),
		uintptr(unsafe.Pointer(WorkingDirp)),
		uintptr(0), // Environment to follow later
		uintptr(unsafe.Pointer(internalDevices)),
		uintptr(EmulateTTY),
		uintptr(unsafe.Pointer(pid)))

	use(unsafe.Pointer(internalDevices))
	use(unsafe.Pointer(IDp))
	use(unsafe.Pointer(ApplicationNamep))
	use(unsafe.Pointer(CommandLinep))
	use(unsafe.Pointer(WorkingDirp))

	log.Debugln("Returned from procedure call")

	if r1 != 0 {
		return 0, syscall.Errno(r1)
	}

	log.Debugln("hcsshim::CreateProcessInComputeSystem PID ", *pid)
	return *pid, nil
} // CreateProcessInComputeSystem

func WaitForProcessInComputeSystem(ID string, ProcessId uint32) (ExitCode uint32, err error) {

	log.Debugln("hcsshim::WaitForProcessInComputeSystem")
	log.Debugln("ID:", ID)
	log.Debugln("ProcessID:", ProcessId)

	var (
		// Infinite
		Timeout uint32 = 0xFFFFFFFF // (-1)
	)

	// Convert ID to uint16 pointer for calling the procedure
	IDp, err := syscall.UTF16PtrFromString(ID)
	if err != nil {
		log.Debugln("Failed conversion of ID to pointer ", err)
		return 0, err
	}

	// To get a POINTER to the ExitCode
	ec := new(uint32)

	// Call the procedure itself.
	r1, _, err := procWaitForProcessInComputeSystem.Call(
		uintptr(unsafe.Pointer(IDp)),
		uintptr(ProcessId),
		uintptr(Timeout),
		uintptr(unsafe.Pointer(ec)))

	use(unsafe.Pointer(IDp))

	if r1 != 0 {
		return 0, syscall.Errno(r1)
	}

	log.Debugln("hcsshim::WaitForProcessInComputeSystem ExitCode ", *ec)
	return *ec, nil
} // WaitForProcessInComputeSystem

func TerminateProcessInComputeSystem(ID string, ProcessId uint32) (err error) {

	log.Debugln("hcsshim::TerminateProcessInComputeSystem")
	log.Debugln("ID:", ID)
	log.Debugln("ProcessID:", ProcessId)

	// Convert ID to uint16 pointer for calling the procedure
	IDp, err := syscall.UTF16PtrFromString(ID)
	if err != nil {
		log.Debugln("Failed conversion of ID to pointer ", err)
		return err
	}

	// Call the procedure itself.
	r1, _, err := procTerminateProcessInComputeSystem.Call(
		uintptr(unsafe.Pointer(IDp)),
		uintptr(ProcessId))

	use(unsafe.Pointer(IDp))

	if r1 != 0 {
		return syscall.Errno(r1)
	}

	return nil
} // TerminateProcessInComputeSystem

func ShutdownComputeSystem(ID string) error {

	log.Debugln("hcsshim::ShutdownComputeSystem")
	log.Debugln("ID:", ID)

	// Convert ID to uint16 pointers for calling the procedure
	IDp, err := syscall.UTF16PtrFromString(ID)
	if err != nil {
		log.Debugln("Failed conversion of ID to pointer ", err)
		return err
	}

	timeout := uint32(0xffffffff)

	// Call the procedure itself.
	r1, _, err := procShutdownComputeSystem.Call(
		uintptr(unsafe.Pointer(IDp)), uintptr(timeout))

	use(unsafe.Pointer(IDp))

	if r1 != 0 {
		return syscall.Errno(r1)
	}

	return nil
} // ShutdownComputeSystem

func ResizeTTY(ID string, h, w int) error {
	log.Debugf("hcsshim::ResizeTTY %s (%d,%d) - NOT IMPLEMENTED", ID, h, w)
	return nil
	/*
		// Make sure ResizeTTY is supported
		err := procResizeTTY.Find()
		if err != nil {
			return err
		}

		// Convert ID to uint16 pointers for calling the procedure
		IDp, err := syscall.UTF16PtrFromString(ID)
		if err != nil {
			log.Debugln("Failed conversion of ID to pointer ", err)
			return err
		}

		h32 := uint32(h)
		w32 := uint32(w)

		r1, _, _ := procResizeTTY.Call(uintptr(unsafe.Pointer(IDp)), uintptr(h32), uintptr(w32))
		if r1 != 0 {
			return syscall.Errno(r1)
		}

		return nil
	*/
}

// use is a no-op, but the compiler cannot see that it is.
// Calling use(p) ensures that p is kept live until that point.
//go:noescape
func use(p unsafe.Pointer) {}
