// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package monitor

import (
	"context"
	"runtime"
	"time"

	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/zeromicro/go-zero/core/logx"
)

type MonitorStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMonitorStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MonitorStatusLogic {
	return &MonitorStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MonitorStatusLogic) MonitorStatus() (resp *types.MonitorStatusResp, err error) {
	// 获取CPU信息
	cpuInfo, err := l.getCPUInfo()
	if err != nil {
		l.Errorf("获取CPU信息失败: %v", err)
		cpuInfo = types.CPUInfo{
			Usage: 0,
			Cores: runtime.NumCPU(),
		}
	}

	// 获取内存信息
	memoryInfo, err := l.getMemoryInfo()
	if err != nil {
		l.Errorf("获取内存信息失败: %v", err)
		memoryInfo = types.MemoryInfo{
			Total:     0,
			Used:      0,
			Available: 0,
			Usage:     0,
		}
	}

	// 获取磁盘信息
	diskInfo, err := l.getDiskInfo()
	if err != nil {
		l.Errorf("获取磁盘信息失败: %v", err)
		diskInfo = types.DiskInfo{
			Total:     0,
			Used:      0,
			Available: 0,
			Usage:     0,
		}
	}

	// 获取网络信息
	networkInfo, err := l.getNetworkInfo()
	if err != nil {
		l.Errorf("获取网络信息失败: %v", err)
		networkInfo = types.NetworkInfo{
			BytesSent:   0,
			BytesRecv:   0,
			PacketsSent: 0,
			PacketsRecv: 0,
		}
	}

	return &types.MonitorStatusResp{
		Cpu:     cpuInfo,
		Memory:  memoryInfo,
		Disk:    diskInfo,
		Network: networkInfo,
	}, nil
}

// getCPUInfo 获取CPU信息
func (l *MonitorStatusLogic) getCPUInfo() (types.CPUInfo, error) {
	// 获取CPU使用率（1秒内的平均值）
	percentages, err := cpu.PercentWithContext(l.ctx, 1*time.Second, false)
	if err != nil {
		return types.CPUInfo{}, err
	}

	usage := 0.0
	if len(percentages) > 0 {
		usage = percentages[0]
	}

	// 获取CPU核心数
	cores, err := cpu.Counts(true)
	if err != nil {
		cores = runtime.NumCPU()
	}

	return types.CPUInfo{
		Usage: usage,
		Cores: cores,
	}, nil
}

// getMemoryInfo 获取内存信息
func (l *MonitorStatusLogic) getMemoryInfo() (types.MemoryInfo, error) {
	vmStat, err := mem.VirtualMemoryWithContext(l.ctx)
	if err != nil {
		return types.MemoryInfo{}, err
	}

	return types.MemoryInfo{
		Total:     vmStat.Total,
		Used:      vmStat.Used,
		Available: vmStat.Available,
		Usage:     vmStat.UsedPercent,
	}, nil
}

// getDiskInfo 获取磁盘信息
func (l *MonitorStatusLogic) getDiskInfo() (types.DiskInfo, error) {
	// 获取根目录磁盘信息
	diskStat, err := disk.UsageWithContext(l.ctx, "/")
	if err != nil {
		// Windows系统使用 "C:"
		diskStat, err = disk.UsageWithContext(l.ctx, "C:")
		if err != nil {
			return types.DiskInfo{}, err
		}
	}

	return types.DiskInfo{
		Total:     diskStat.Total,
		Used:      diskStat.Used,
		Available: diskStat.Free,
		Usage:     diskStat.UsedPercent,
	}, nil
}

// getNetworkInfo 获取网络信息
func (l *MonitorStatusLogic) getNetworkInfo() (types.NetworkInfo, error) {
	// 获取网络IO统计
	ioCounters, err := net.IOCountersWithContext(l.ctx, false)
	if err != nil {
		return types.NetworkInfo{}, err
	}

	if len(ioCounters) == 0 {
		return types.NetworkInfo{}, nil
	}

	// 汇总所有网络接口的统计信息
	var bytesSent, bytesRecv, packetsSent, packetsRecv uint64
	for _, counter := range ioCounters {
		bytesSent += counter.BytesSent
		bytesRecv += counter.BytesRecv
		packetsSent += counter.PacketsSent
		packetsRecv += counter.PacketsRecv
	}

	return types.NetworkInfo{
		BytesSent:   bytesSent,
		BytesRecv:   bytesRecv,
		PacketsSent: packetsSent,
		PacketsRecv: packetsRecv,
	}, nil
}
