// package main

// import (
//   "fmt"
//   "log"
//   "strings"

//   "github.com/google/gopacket"
//   "github.com/google/gopacket/layers"
//   "github.com/google/gopacket/pcap"
// )

// var (
//   pcapFile = "../tcpdumps/tcpdump0.pcap"
//   handle   *pcap.Handle
//   err      error
// )
// var dns = map[string]string{
//   "10.1.19.12": "books1",
//   "10.1.19.9": "auth1",
//   "10.1.19.15":  "db",
//   "10.1.19.3":  "gateway1",
//   "10.1.19.13": "books2",
// //   "10.1.19.7":  "jaeger",
//   "10.1.19.10": "auth2",
//   "10.1.19.4":  "gateway2",
//   "10.1.19.5":  "lb",
//   "10.1.19.11": "book_service",
// //   "10.1.19.6":  "jaeger_service",
//   "10.1.19.8": "auth_service",
//   "10.1.19.2":  "gateway_service",
//   "10.1.19.14":  "db_service"}

// func main() {
// 	f1()
// //   // Open file instead of device
// //   handle, err = pcap.OpenOffline(pcapFile)
// //   if err != nil {
// //     log.Fatal(err)
// //   }
// //   defer handle.Close()

// //   c := 0
// //   // Loop through packets in file
// //   packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
// //   for packet := range packetSource.Packets() {
// //     printPacketInfo(packet)
// //     c++
// //     if c == 2000000 {
// //       break
// //     }
// //     // fmt.Println("====")
// //   }
// }

// func printPacketInfo(packet gopacket.Packet) {
//   // Let's see if the packet is an ethernet packet
//   // ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
//   // if ethernetLayer != nil {
//   //   fmt.Println("Ethernet layer detected.")
//   //   ethernetPacket, _ := ethernetLayer.(*layers.Ethernet)
//   //   fmt.Println("Source MAC: ", ethernetPacket.SrcMAC)
//   //   fmt.Println("Destination MAC: ", ethernetPacket.DstMAC)
//   //   // Ethernet type is typically IPv4 but could be ARP or other
//   //   fmt.Println("Ethernet type: ", ethernetPacket.EthernetType)
//   //   fmt.Println()
//   // }

//   // Let's see if the packet is IP (even though the ether type told
//   //us)

//   srcName := ""
//   dstName := ""
//   ignore := false
//   ipLayer := packet.Layer(layers.LayerTypeIPv4)
//   if ipLayer != nil {
//     // fmt.Println("IPv4 layer detected.")
//     ip, _ := ipLayer.(*layers.IPv4)

//     // IP layer variables:
//     // Version (Either 4 or 6)
//     // IHL (IP Header Length in 32-bit words)
//     // TOS, Length, Id, Flags, FragOffset, TTL, Protocol (TCP?),
//     // Checksum, SrcIP, DstIP
//     srcName = ip.SrcIP.String()
//     if val, ok := dns[srcName]; ok {
//       srcName = val
//     }else{
// 		ignore = true
// 	}

//     dstName = ip.DstIP.String()
//     if val, ok := dns[dstName]; ok {
//       dstName = val
//     }else{
// 		ignore = true
// 	}
//   }

//   // Let's see if the packet is TCP
//   tcpLayer := packet.Layer(layers.LayerTypeTCP)
//   var secNumber uint32 = 0
//   var srcPort string
//   var dstPort string

//   if tcpLayer != nil {
//     //   fmt.Println("TCP layer detected.")
//     tcp, _ := tcpLayer.(*layers.TCP)

//     //   // TCP layer variables:
//     // SrcPort, DstPort, Seq, Ack, DataOffset, Window, Checksum,
//     //   //Urgent
//     //   // Bool flags: FIN, SYN, RST, PSH, ACK, URG, ECE, CWR, NS
//     //   fmt.Printf("From port %d to %d\n", tcp.SrcPort, tcp.DstPort)
//     secNumber = tcp.Seq
//     srcPort = tcp.SrcPort.String()
//     dstPort = tcp.DstPort.String()
//     //   fmt.Println()
//   }

//   // Iterate over all layers, printing out each layer type
//   // fmt.Println("All packet layers:")
//   // for _, layer := range packet.Layers() {
//   //   fmt.Println("- ", layer.LayerType())
//   // }

//   // When iterating through packet.Layers() above,
//   // if it lists Payload layer then that is the same as
//   // this applicationLayer. applicationLayer contains the payload
//   applicationLayer := packet.ApplicationLayer()
//   if applicationLayer != nil {
//     // fmt.Println("Application layer/Payload found.")

//     // Search for a string inside the payload
//     if strings.Contains(string(applicationLayer.Payload()), "HTTP") && !ignore{
//       // fmt.Println("HTTP found!")
//       fmt.Println(packet.Metadata().Timestamp)
//       fmt.Println(secNumber)
//       fmt.Printf("From %s:%s to %s:%s\n", srcName, srcPort, dstName, dstPort)
//       fmt.Printf("%s\n", applicationLayer.Payload())
//       fmt.Println("======")
//     }
//   }

//   // Check for errors
//   if err := packet.ErrorLayer(); err != nil {
//     fmt.Println("Error decoding some part of the packet:", err)
//   }
// }