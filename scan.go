package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/gopacket"

	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	"github.com/hashicorp/go-hclog"
	"github.com/miekg/dns"
)

const timeout = 2 * time.Second

type ScanExecutor struct {
	Logger hclog.Logger
}

func NewScanExecutor(logger hclog.Logger) *ScanExecutor {
	return &ScanExecutor{
		Logger: logger,
	}
}

func (s *ScanExecutor) Execute(args ...interface{}) (string, error) {
	startIP := "192.168.1.1"
	endIP := "192.168.1.255"
	ports := []int{22, 80, 443, 21, 25}
	scanTimeout := 60 * time.Second

	if len(args) >= 2 {
		startIP = fmt.Sprintf("%v", args[0])
		endIP = fmt.Sprintf("%v", args[1])
	}

	if len(args) >= 3 {
		ports = parsePorts(args[2], s.Logger)
	}

	if len(args) >= 4 {
		if timeout, err := strconv.Atoi(fmt.Sprintf("%v", args[3])); err == nil && timeout > 0 {
			scanTimeout = time.Duration(timeout) * time.Second
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), scanTimeout)
	defer cancel()

	results := s.scanNetwork(ctx, startIP, endIP, ports)
	return s.formatResponse(results)
}

func (s *ScanExecutor) scanNetwork(ctx context.Context, startIP, endIP string, ports []int) []map[string]interface{} {
	var wg sync.WaitGroup
	results := []map[string]interface{}{}
	resultLock := &sync.Mutex{}

	start := ipToInt(startIP)
	end := ipToInt(endIP)

	for ip := start; ip <= end; ip++ {
		select {
		case <-ctx.Done():
			s.Logger.Warn("扫描任务已超时")
			return results
		default:
			ipStr := intToIP(ip)
			if isHostAlive(ipStr, ports) {
				wg.Add(1)
				go func(ip string) {
					defer wg.Done()
					s.scanHost(ip, ports, resultLock, &results)
				}(ipStr)
			}
		}
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		s.Logger.Warn("扫描任务已超时")
		return results
	case <-done:
		return results
	}
}

func (s *ScanExecutor) scanHost(ip string, ports []int, lock *sync.Mutex, results *[]map[string]interface{}) {
	for _, port := range ports {
		address := fmt.Sprintf("%s:%d", ip, port)
		conn, err := net.DialTimeout("tcp", address, timeout)
		if err != nil {
			continue
		}

		service := s.identifyService(conn)
		if service == "Unknown" {
			continue
		}
		conn.Close()

		lock.Lock()
		*results = append(*results, map[string]interface{}{
			"ip":      ip,
			"port":    port,
			"service": service,
		})
		lock.Unlock()
	}
}

func (s *ScanExecutor) identifyService(conn net.Conn) string {
	conn.SetReadDeadline(time.Now().Add(timeout))
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return "Unknown"
	}

	banner := string(buffer[:n])
	s.Logger.Log(hclog.Trace, "scan-banner", "banner", banner)

	switch {
	case strings.Contains(banner, "SSH"):
		return "SSH Server"
	case strings.Contains(banner, "HTTP"):
		return "Web Server"
	case strings.Contains(banner, "FTP"):
		return "FTP Server"
	case strings.Contains(banner, "SMTP"):
		return "Mail Server"
	default:
		return "Unknown"
	}
}

func (s *ScanExecutor) formatResponse(results []map[string]interface{}) (string, error) {
	bytes, err := json.Marshal(results)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func parsePorts(input interface{}, logger hclog.Logger) []int {
	var ports []int

	switch v := input.(type) {
	case string:
		for _, part := range strings.Split(v, ",") {
			part = strings.TrimSpace(part)

			// 处理端口范围
			if strings.Contains(part, "-") {
				rangeParts := strings.Split(part, "-")
				if len(rangeParts) != 2 {
					logger.Error("无效的端口范围格式", "input", part)
					continue
				}

				start, err1 := strconv.Atoi(rangeParts[0])
				end, err2 := strconv.Atoi(rangeParts[1])
				if err1 != nil || err2 != nil || start > end || !isValidPort(start) || !isValidPort(end) {
					logger.Error("无效的端口范围", "start", start, "end", end)
					continue
				}

				for p := start; p <= end; p++ {
					ports = append(ports, p)
				}
			} else {
				if port, err := strconv.Atoi(part); err == nil && isValidPort(port) {
					ports = append(ports, port)
				}
			}
		}
	case []interface{}:
		for _, item := range v {
			switch num := item.(type) {
			case int:
				if isValidPort(num) {
					ports = append(ports, num)
				}
			case float64:
				if port := int(num); num == float64(port) && isValidPort(port) {
					ports = append(ports, port)
				}
			}
		}
	case int:
		if isValidPort(v) {
			ports = append(ports, v)
		}
	case float64:
		if port := int(v); v == float64(port) && isValidPort(port) {
			ports = append(ports, port)
		}
	}

	if len(ports) == 0 {
		logger.Error("无效的端口参数", "input", input)
	}
	return ports
}

func isValidPort(port int) bool {
	return port > 0 && port <= 65535
}

// 将IP转换为整数
func ipToInt(ip string) uint32 {
	parts := strings.Split(ip, ".")
	var result uint32
	for i, part := range parts {
		num := atoi(part)
		result += uint32(num) << uint(24-8*i)
	}
	return result
}

// 将整数转换为IP
func intToIP(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		ip>>24&0xff,
		ip>>16&0xff,
		ip>>8&0xff,
		ip&0xff)
}

// 字符串转整数
func atoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

// 多层级主机存活检测策略
func isHostAlive(ip string, ports []int) bool {

	// TCP检测（最高优先级）
	if checkTCP(ip, ports) {
		return true
	}
	// UDP检测（第二优先级）
	if checkUDP(ip, ports) {
		return true
	}
	// 第三优先级
	if checkHTTP(ip) {
		return true
	}
	//  HTTP检测（最后尝试）
	return checkARP(ip)
}

func checkARP(ip string) bool {
	// 获取网络接口列表
	ifaces, err := net.Interfaces()
	if err != nil {
		return false
	}

	// 使用channel来实现快速失败机制
	resultChan := make(chan bool, len(ifaces))
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// 并发处理每个网卡
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		go func(iface net.Interface) {
			// 创建ARP请求包
			handle, err := pcap.OpenLive(iface.Name, 65536, true, pcap.BlockForever)
			if err != nil {
				resultChan <- false
				return
			}
			defer handle.Close()

			// 设置过滤器
			if err := handle.SetBPFFilter(fmt.Sprintf("arp and host %s", ip)); err != nil {
				resultChan <- false
				return
			}

			// 发送ARP请求
			srcIP := getInterfaceIP(&iface)
			if srcIP == nil {
				resultChan <- false
				return
			}

			eth := layers.Ethernet{
				SrcMAC:       iface.HardwareAddr,
				DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				EthernetType: layers.EthernetTypeARP,
			}
			arp := layers.ARP{
				AddrType:          layers.LinkTypeEthernet,
				Protocol:          layers.EthernetTypeIPv4,
				HwAddressSize:     6,
				ProtAddressSize:   4,
				Operation:         layers.ARPRequest,
				SourceHwAddress:   []byte(iface.HardwareAddr),
				SourceProtAddress: []byte(srcIP.To4()),
				DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
				DstProtAddress:    []byte(net.ParseIP(ip).To4()),
			}

			buffer := gopacket.NewSerializeBuffer()
			opts := gopacket.SerializeOptions{
				FixLengths:       true,
				ComputeChecksums: true,
			}

			if err = gopacket.SerializeLayers(buffer, opts, &eth, &arp); err != nil {
				resultChan <- false
				return
			}

			if err := handle.WritePacketData(buffer.Bytes()); err != nil {
				resultChan <- false
				return
			}

			// 使用context控制超时
			for {
				select {
				case <-ctx.Done():
					resultChan <- false
					return
				default:
					data, _, err := handle.ReadPacketData()
					if err != nil {
						continue
					}

					packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)
					if arpLayer := packet.Layer(layers.LayerTypeARP); arpLayer != nil {
						arpResp, _ := arpLayer.(*layers.ARP)
						if arpResp.Operation == layers.ARPReply &&
							net.IP(arpResp.SourceProtAddress).Equal(net.ParseIP(ip)) {
							resultChan <- true
							return
						}
					}
				}
			}
		}(iface)
	}

	// 等待任意一个goroutine返回true或者所有goroutine都返回false
	for i := 0; i < len(ifaces); i++ {
		select {
		case <-ctx.Done():
			return false
		case result := <-resultChan:
			if result {
				return true
			}
		}
	}
	return false
}

func getInterfaceIP(iface *net.Interface) net.IP {
	addrs, err := iface.Addrs()
	if err != nil {
		return nil
	}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip == nil || ip.IsLoopback() {
			continue
		}
		ip = ip.To4()
		if ip != nil {
			return ip
		}
	}
	return nil
}

func checkTCP(ip string, ports []int) bool {
	for _, port := range ports {
		address := fmt.Sprintf("%s:%d", ip, port)
		conn, err := net.DialTimeout("tcp", address, timeout)
		if err == nil {
			conn.Close()
			return true
		}
	}
	return false
}

func checkUDP(ip string, ports []int) bool {
	for _, port := range ports {
		address := fmt.Sprintf("%s:%d", ip, port)
		conn, err := net.DialTimeout("udp", address, timeout)
		if err != nil {
			continue
		}
		defer conn.Close()

		// 发送DNS查询作为示例UDP探测
		conn.SetDeadline(time.Now().Add(timeout))
		if _, err := conn.Write(createDNSQuery()); err != nil {
			continue
		}

		buffer := make([]byte, 1024)
		_, err = conn.Read(buffer)
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok {
				if syscallErr, ok := opErr.Err.(*os.SyscallError); ok {
					if syscallErr.Err == syscall.ECONNREFUSED {
						return true
					}
				}
			}
		} else {
			return true
		}
	}
	return false
}

func createDNSQuery() []byte {
	msg := &dns.Msg{
		MsgHdr: dns.MsgHdr{Id: dns.Id(), RecursionDesired: true},
		Question: []dns.Question{{
			Name:   ".",
			Qtype:  dns.TypeA,
			Qclass: dns.ClassINET,
		}},
	}
	pack, _ := msg.Pack()
	return pack
}

func checkHTTP(ip string) bool {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Head(fmt.Sprintf("http://%s", ip))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}
