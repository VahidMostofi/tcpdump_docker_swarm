package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var (
	handle         *pcap.Handle
	err            error
	deploymentInfo *DeploymentInfo
)

func Parse(_deploymentInfo *DeploymentInfo) {
	deploymentInfo = _deploymentInfo
	pcapFile := FSBase + "/" + deploymentInfo.DefaultNetworkID[:12]

	// Open file instead of device
	handle, err = pcap.OpenOffline(pcapFile)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// c := 0
	// Loop through packets in file
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	//strconv.FormatInt(ts, 10), srcName, dstName, reqType, traceID
	fmt.Println("timestamp,", "src,", "dst,", "req,", "traceID")
	for packet := range packetSource.Packets() {
		printPacketInfo(packet)
		// c++
		// if c == 21 {
		// 	break
		// }
	}
}

func parsePayload(payload []byte) (string, string) {

	// 	buf := bufio.NewReader(bytes.NewReader(payload))
	// 	req, err := http.ReadRequest(buf)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Println(req)
	// }
	var reqType string
	var traceID string
	p := string(payload)
	if strings.HasPrefix(strings.Split(p, "\n")[0], "GET /books") {
		reqType = "get_books"
	} else if strings.HasPrefix(strings.Split(p, "\n")[0], "PUT /books") {
		reqType = "edit_books"
	} else if strings.HasPrefix(strings.Split(p, "\n")[0], "POST") {
		reqType = "auth_login"
	}

	for _, row := range strings.Split(p, "\n") {
		if strings.HasPrefix(row, "uber-trace-id:") {
			traceID = strings.Trim(strings.Join(strings.Split(row, ":")[1:], ":"), " \t")
		}
	}
	// if len(reqType) < 2 {

	// 	panic("couldn't find reqType " + p)
	// }
	if len(traceID) < 2 {
		fmt.Println(string(payload))
		panic("couldn't find traceID")
	}
	return reqType, traceID
	// fmt.Println(strings.Split(p, "\n")[6])
}
func printPacketInfo(packet gopacket.Packet) {
	// Let's see if the packet is an ethernet packet
	// ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
	// if ethernetLayer != nil {
	//   fmt.Println("Ethernet layer detected.")
	//   ethernetPacket, _ := ethernetLayer.(*layers.Ethernet)
	//   fmt.Println("Source MAC: ", ethernetPacket.SrcMAC)
	//   fmt.Println("Destination MAC: ", ethernetPacket.DstMAC)
	//   // Ethernet type is typically IPv4 but could be ARP or other
	//   fmt.Println("Ethernet type: ", ethernetPacket.EthernetType)
	//   fmt.Println()
	// }

	// Let's see if the packet is IP (even though the ether type told
	//us)

	srcName := ""
	dstName := ""
	ignore := false
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		// fmt.Println("IPv4 layer detected.")
		ip, _ := ipLayer.(*layers.IPv4)

		srcName = ip.SrcIP.String()

		if val, ok := deploymentInfo.DNS[srcName]; ok {
			srcName = val
			// fmt.Println(srcName)
		} else {
			ignore = true
		}

		dstName = ip.DstIP.String()
		if val, ok := deploymentInfo.DNS[dstName]; ok {
			dstName = val
		} else {
			ignore = true
		}
	}

	if srcName == "db_service" || srcName == "db.1" || srcName == "jaeger.1" || srcName == "jaeger_service" {
		ignore = true
	}
	if dstName == "db_service" || dstName == "db.1" || dstName == "jaeger.1" || dstName == "jaeger_service" {
		ignore = true
	}

	// Let's see if the packet is TCP
	// tcpLayer := packet.Layer(layers.LayerTypeTCP)
	// var secNumber uint32 = 0
	// var srcPort string
	// var dstPort string
	// var isAck bool = false
	// if tcpLayer != nil {
	//   fmt.Println("TCP layer detected.")
	// tcp, _ := tcpLayer.(*layers.TCP)

	//   // TCP layer variables:
	// SrcPort, DstPort, Seq, Ack, DataOffset, Window, Checksum,
	//   //Urgent
	//   // Bool flags: FIN, SYN, RST, PSH, ACK, URG, ECE, CWR, NS
	//   fmt.Printf("From port %d to %d\n", tcp.SrcPort, tcp.DstPort)
	// isAck = tcp.ACK
	// secNumber = tcp.Seq
	// isAck = tcp.FIN
	// srcPort = tcp.SrcPort.String()
	// dstPort = tcp.DstPort.String()
	//   fmt.Println()
	// }

	// Iterate over all layers, printing out each layer type
	// fmt.Println("All packet layers:")
	// for _, layer := range packet.Layers() {
	//   fmt.Println("- ", layer.LayerType())
	// }

	// When iterating through packet.Layers() above,
	// if it lists Payload layer then that is the same as
	// this applicationLayer. applicationLayer contains the payload
	applicationLayer := packet.ApplicationLayer()
	if applicationLayer != nil {
		// fmt.Println("Application layer/Payload found.")

		// Search for a string inside the payload
		if strings.Contains(string(applicationLayer.Payload()), "HTTP") && !ignore {
			// fmt.Println("HTTP found!")
			ts := packet.Metadata().Timestamp.UnixNano()
			// fmt.Println(strconv.FormatInt(ts, 10))
			// fmt.Println(secNumber)
			// if isAck {
			// fmt.Println("is FIN")
			// }

			// fmt.Printf("%s\n", applicationLayer.Payload())
			reqType, traceID := parsePayload(applicationLayer.Payload())
			fmt.Printf("%s,%s,%s,%s,%s\n", strconv.FormatInt(ts, 10), srcName, dstName, reqType, traceID)

		}
	}

	// Check for errors
	if err := packet.ErrorLayer(); err != nil {
		fmt.Println("Error decoding some part of the packet:", err)
	}
}
