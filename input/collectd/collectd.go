// Input-plug: collectdinput
// The plug's function is real-time monitoring of the system include cpu,memory,disk,host,net infomation.
package collectdinput

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"

	"github.com/wgliang/logcool/utils"
)

const (
	ModuleName = "collectd"
)

// Define collectdinput' config.
type InputConfig struct {
	utils.InputConfig

	hostname string
}

type CpuInfo struct {
	Info []cpu.InfoStat  `json:"info"`
	Time []cpu.TimesStat `json:"time"`
}

type MemoryInfo struct {
	VirtualMemory string `json:"virtualMemory"`
	SwapMemory    string `json:"swapMemory"`
}

type DiskInfo struct {
	Usage      string                         `json:"usage"`
	Partition  []disk.PartitionStat           `json:"partition"`
	IOCounters map[string]disk.IOCountersStat `json:"iOCounters"`
}

type HostInfo struct {
	Info *host.InfoStat  `json:"info"`
	User []host.UserStat `json:"user"`
}

type NetInfo struct {
	IOCounters    []net.IOCountersStat    `json:"iOCounters"`
	Connection    []net.ConnectionStat    `json:"connection"`
	ProtoCounters []net.ProtoCountersStat `json:"protoCounters"`
	Interface     []net.InterfaceStat     `json:"interface"`
	Filter        []net.FilterStat        `json:"filter"`
}

type ProcessInfo struct {
	Name           string                      `json:"name"`
	Pid            int32                       `json:"pid"`
	Ppid           int32                       `json:"ppid"`
	Exe            string                      `json:"exe"`
	Cmdline        string                      `json:"cmdline"`
	CmdlineSlice   []string                    `json:"cmdlineSlice"`
	CreateTime     int64                       `json:"createTime"`
	Cwd            string                      `json:"cwd"`
	Parent         *process.Process            `json:"parent"`
	Status         string                      `json:"status"`
	Uids           []int32                     `json:"uids"`
	Gids           []int32                     `json:"gids"`
	Terminal       string                      `json:"terminal"`
	Nice           int32                       `json:"nice"`
	IOnice         int32                       `json:"iOnice"`
	Rlimit         []process.RlimitStat        `json:"rlimit"`
	IOCounters     *process.IOCountersStat     `json:"iOCounters"`
	NumCtxSwitches *process.NumCtxSwitchesStat `json:"numCtxSwitches"`
	NumFDs         int32                       `json:"numFDs"`
	NumThreads     int32                       `json:"numThreads"`
	Threads        map[string]string           `json:"threads"`
	Times          *cpu.TimesStat              `json:"times"`
	CPUAffinity    []int32                     `json:"cpuAffinity"`
	MemoryInfo     *process.MemoryInfoStat     `json:"memoryInfo"`
	MemoryInfoEx   *process.MemoryInfoExStat   `json:"memoryInfoEx"`
	Children       []*process.Process          `json:"children"`
	OpenFiles      []process.OpenFilesStat     `json:"openFiles"`
	Connections    []net.ConnectionStat        `json:"connections"`
	NetIOCounters  []net.IOCountersStat        `json:"netIOCounters"`
	IsRunning      bool                        `json:"isRunning"`
	MemoryMaps     *[]process.MemoryMapsStat   `json:"memoryMaps"`
}

type SysInfo struct {
	Host    HostInfo    `json:"host"`
	Cpu     CpuInfo     `json:"cpu"`
	Mem     MemoryInfo  `json:"mem"`
	Disk    DiskInfo    `json:"disk"`
	Net     NetInfo     `json:"net"`
	Process ProcessInfo `json:"process"`
}

func init() {
	utils.RegistInputHandler(ModuleName, InitHandler)
}

// Init fileinput Handler.
func InitHandler(confraw *utils.ConfigRaw) (retconf utils.TypeInputConfig, err error) {
	conf := InputConfig{
		InputConfig: utils.InputConfig{
			CommonConfig: utils.CommonConfig{
				Type: ModuleName,
			},
		},
	}
	if err = utils.ReflectConfig(confraw, &conf); err != nil {
		return
	}
	// get hostname.
	if conf.hostname, err = os.Hostname(); err != nil {
		return
	}

	retconf = &conf
	return
}

// Input's start,and this is the main function of input.
func (ic *InputConfig) Start() {
	ic.Invoke(ic.monitor)
}

// monitor all system information
func (t *InputConfig) monitor(logger *logrus.Logger, inchan utils.InChan) (err error) {
	defer func() {
		if err != nil {
			logger.Errorln(err)
		}
	}()
	for {
		info := SysInfo{
			Host:    HostStat(),
			Cpu:     CpuStat(),
			Mem:     MemStat(),
			Disk:    DiskStat(),
			Net:     NetStat(),
			Process: ProcessStat(),
		}
		stsinfo, err := json.Marshal(info)
		if err != nil {
			fmt.Println(err)
			break
		}

		message := string(stsinfo)

		event := utils.LogEvent{
			Timestamp: time.Now(),
			Message:   message,
			Extra: map[string]interface{}{
				"host": t.hostname,
			},
		}
		inchan <- event
		// take a event every 3 seconds
		time.Sleep(3 * time.Second)
	}
	return
}

func CpuStat() CpuInfo {
	info, _ := cpu.Info()
	times, _ := cpu.Times(true)

	cpustat := CpuInfo{
		Info: info,
		Time: times,
	}
	return cpustat
}

func MemStat() MemoryInfo {
	virt, _ := mem.VirtualMemory()
	virtmem := virt.String()
	swap, _ := mem.SwapMemory()
	swapmem := swap.String()

	memstat := MemoryInfo{
		VirtualMemory: virtmem,
		SwapMemory:    swapmem,
	}

	return memstat
}

func DiskStat() DiskInfo {
	usage, _ := disk.Usage("./")
	partitions, _ := disk.Partitions(true)
	iOCounters, _ := disk.IOCounters()

	ue := usage.String()

	diskstat := DiskInfo{
		Usage:      ue,
		Partition:  partitions,
		IOCounters: iOCounters,
	}

	return diskstat
}

func NetStat() NetInfo {
	iOCounters, _ := net.IOCounters(true)
	protoCounters, _ := net.ProtoCounters([]string{"tcp", "http", "udp", "snmp", "ftp"})
	filterCounters, _ := net.FilterCounters()
	connections, _ := net.Connections("tcp")

	interfaces, _ := net.Interfaces()

	netstat := NetInfo{
		IOCounters:    iOCounters,
		Connection:    connections,
		ProtoCounters: protoCounters,
		Interface:     interfaces,
		Filter:        filterCounters,
	}

	return netstat

}
func getSelfProcess() process.Process {
	checkPid := os.Getpid() // process.test
	ret, _ := process.NewProcess(int32(checkPid))
	return *ret
}

func ProcessStat() ProcessInfo {
	pro := getSelfProcess()
	processinfo := new(ProcessInfo)

	processinfo.Name, _ = pro.Name()
	processinfo.Pid = int32(os.Getpid())
	processinfo.Ppid, _ = pro.Ppid()
	processinfo.Exe, _ = pro.Exe()
	processinfo.Cmdline, _ = pro.Cmdline()
	processinfo.CmdlineSlice, _ = pro.CmdlineSlice()
	processinfo.CreateTime, _ = pro.CreateTime()
	processinfo.Cwd, _ = pro.Cwd()
	processinfo.Parent, _ = pro.Parent()
	processinfo.Status, _ = pro.Status()
	processinfo.Uids, _ = pro.Uids()
	processinfo.Gids, _ = pro.Gids()
	processinfo.Terminal, _ = pro.Terminal()
	processinfo.Nice, _ = pro.Nice()
	processinfo.IOnice, _ = pro.IOnice()
	processinfo.Rlimit, _ = pro.Rlimit()
	processinfo.IOCounters, _ = pro.IOCounters()
	processinfo.NumCtxSwitches, _ = pro.NumCtxSwitches()
	processinfo.NumFDs, _ = pro.NumFDs()
	processinfo.NumThreads, _ = pro.NumThreads()
	processinfo.Threads, _ = pro.Threads()
	processinfo.Times, _ = pro.Times()
	processinfo.CPUAffinity, _ = pro.CPUAffinity()
	processinfo.MemoryInfo, _ = pro.MemoryInfo()
	processinfo.MemoryInfoEx, _ = pro.MemoryInfoEx()
	processinfo.Children, _ = pro.Children()
	processinfo.OpenFiles, _ = pro.OpenFiles()
	processinfo.Connections, _ = pro.Connections()
	processinfo.NetIOCounters, _ = pro.NetIOCounters(true)
	processinfo.IsRunning, _ = pro.IsRunning()
	processinfo.MemoryMaps, _ = pro.MemoryMaps(true)

	p := ProcessInfo(*processinfo)

	return p
}

func HostStat() HostInfo {
	info, _ := host.Info()
	users, _ := host.Users()

	hoststat := HostInfo{
		Info: info,
		User: users,
	}
	return hoststat
}
