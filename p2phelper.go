package main

import (
	"bytes"
	"github.com/gogf/greuse"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
)

type ProtocolType string

func String(protocolType ProtocolType) string {
	return string(protocolType)
}

const (
	IPV6        ProtocolType = "ipv6"
	IPV4        ProtocolType = "ipv4"
	STUN_FC     ProtocolType = "fc"
	STUN_RC     ProtocolType = "rc"
	STUN_PRC    ProtocolType = "prc"
	STUN_SYM    ProtocolType = "sym"
	UNAVALIABLE ProtocolType = "null"
)

type P2PType struct {
	protocol      ProtocolType
	conn          net.Conn
	ip            string
	inner_port    uint16
	external_port uint16
}

func P2PHelper() P2PType {
	resp, err := http.Get("http://v6.ip.zxinc.org/getip")
	if err == nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			return P2PType{protocol: IPV6, ip: string(body), inner_port: 60000, external_port: 60000}
		}
	} else {
		//local_addr,_:=net.ResolveAddr("tcp4","0.0.0.0:0")
		//remote_addr,_:=net.ResolveTCPAddr("tcp4","tran.host:60001")
		conn, err := greuse.Dial("tcp4", "0.0.0.0:60000", "tran.host:60001")

		if err == nil {
			buf := make([]byte, 100)
			conn.Read(buf)
			cnt := bytes.IndexByte(buf, '\n')

			if cnt != -1 {
				println(string(buf[:cnt]))
				addr := strings.Split(string(buf[:cnt]), ":")
				port, _ := strconv.Atoi(addr[1])
				return P2PType{conn: conn, protocol: IPV4, ip: addr[0], inner_port: 60000, external_port: uint16(port)}
			}
		}

		//udpaddr,_:=net.ResolveUDPAddr("udp4", "0.0.0.0:60000")
		//udpconn,_:=net.ListenUDP("udp4",udpaddr)
		//defer udpconn.Close()
		//stuncli:=stun.NewClientWithConnection(udpconn)
		//stuncli.SetServerAddr("stun.voipbuster.com:3478")
		//nat, host, err := stuncli.Discover()
		//
		//if(err==nil){
		//	//host_local,err:=stuncli.Keepalive()
		//	//if(err==nil){
		//		switch nat {
		//		case stun.NATFull:
		//			return P2PType{protocol:STUN_FC,ip:host.IP(),external_port: host.Port()}
		//		case stun.NATPortRestricted:
		//			return P2PType{protocol:STUN_PRC,ip:host.IP(),external_port: host.Port()}
		//		case stun.NATSymmetric:
		//			return P2PType{protocol:STUN_SYM,ip:host.IP(),external_port: host.Port()}
		//		case stun.NATRestricted:
		//			return P2PType{protocol:STUN_RC,ip:host.IP(),external_port: host.Port()}
		//		case stun.NATNone:
		//			return P2PType{protocol:IPV4,ip:host.IP(),external_port: host.Port()}
		//		default:
		//			return P2PType{protocol: UNAVALIABLE}
		//		}
		//	//}
		//
		//}
	}
	return P2PType{protocol: UNAVALIABLE}
}
