package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type Packet struct {
	Timestamp   int64
	Source      string
	Destination string
	ReqType     string
	TraceID     string
	DebugID     string
}

var (
	handle         *pcap.Handle
	err            error
	deploymentInfo *DeploymentInfo
)

func Parse(_deploymentInfo *DeploymentInfo) {
	deploymentInfo = _deploymentInfo
	pcapFile := FSBase + "/" + _deploymentInfo.Networks["overlay"].ShortID + "/merged.pcap"

	// Open file instead of device
	handle, err = pcap.OpenOffline(pcapFile)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packets := make([]*Packet, 0)
	for packet := range packetSource.Packets() {
		p := printPacketInfo(packet)
		if p != nil {
			packets = append(packets, p)
		}
	}

	s := "ts,src,dst,req,traceID,debugID\n"
	for _, p := range packets {
		s += fmt.Sprintf("%.02f,%s,%s,%s,%s,%s\n", (float64(p.Timestamp) / 1000000.0), p.Source, p.Destination, p.ReqType, strings.ReplaceAll(p.TraceID, "\n", ""), p.DebugID)
	}

	ioutil.WriteFile(FSBase+_deploymentInfo.Networks["overlay"].ShortID+"/http_packets.csv", []byte(s), 0777)
	fmt.Println(FSBase + _deploymentInfo.Networks["overlay"].ShortID + "/http_packets.csv")

	// convert packets to reqType -> debug_id -> [packets]
	// data := make(map[string]map[string][]*Packet, 3)

	// for _, p := range packets {
	// 	// if the reqType has not seen yet:
	// 	if _, ok := data[p.ReqType]; !ok {
	// 		data[p.ReqType] = make(map[string][]*Packet, 0)
	// 	}

	// 	// if the debugID has not seen yet:
	// 	if _, ok := data[p.ReqType][p.DebugID]; !ok {
	// 		data[p.ReqType][p.DebugID] = make([]*Packet, 0)
	// 	}

	// 	data[p.ReqType][p.DebugID] = append(data[p.ReqType][p.DebugID], p)
	// }

	// stats(data)

}

// func stats(data map[string]map[string][]*Packet) {
// 	for reqType, debugID2Packets := range data {
// 		fmt.Println(reqType)
// 		for debugID, packets := range debugID2Packets {
// 			sort.Slice(packets, func(i, j int) bool {
// 				return packets[i].Timestamp < packets[j].Timestamp
// 			})
// 			fmt.Println(debugID)
// 			for _, p := range packets {
// 				fmt.Println(p.Timestamp)
// 			}
// 		}
// 	}
// }

func parsePayload(payload []byte) (string, string, string) {

	var reqType string
	var traceID string
	var debugID string
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
			traceID = strings.Trim(strings.Join(strings.Split(row, ":")[1:], ":"), " \t\n\r ")
		}

		if strings.HasPrefix(row, "debug_id:") || strings.HasPrefix(row, "Debug_id:") {
			debugID = strings.Trim(strings.Split(row, ": ")[1], "\n\t\r ")
		}
	}
	return reqType, traceID, debugID
}

func printPacketInfo(packet gopacket.Packet) *Packet {
	var p *Packet
	srcName := ""
	dstName := ""
	ignore := false
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		ip, _ := ipLayer.(*layers.IPv4)

		srcName = ip.SrcIP.String()

		if val, ok := deploymentInfo.DNS[srcName]; ok {
			srcName = val
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
	ignores := map[string]string{
		"jaeger_srv_ing": "",
		"jaeger.1_ing":   "",
		"jaeger_srv_def": "",
		"jaeger.1_def":   "",
		"db.1_def":       "",
		"db_srv_def":     "",
		"db_srv_ing":     "",
		"db.1_ing":       "",
	}

	if _, ok := ignores[srcName]; ok {
		ignore = true
	}
	if _, ok := ignores[dstName]; ok {
		ignore = true
	}

	applicationLayer := packet.ApplicationLayer()
	if applicationLayer != nil {
		if strings.Contains(string(applicationLayer.Payload()), "HTTP") && !ignore {

			ts := packet.Metadata().Timestamp.UnixNano()
			reqType, traceID, debugID := parsePayload(applicationLayer.Payload())
			if len(debugID) == 0 {
				return nil
			}
			p = &Packet{
				Timestamp:   ts,
				Source:      srcName,
				Destination: dstName,
				ReqType:     reqType,
				TraceID:     traceID,
				DebugID:     debugID,
			}

			return p
		}
	}

	// if err := packet.ErrorLayer(); err != nil {
	// 	fmt.Println("Error decoding some part of the packet:", err)
	// 	return nil
	// }

	return nil
}
