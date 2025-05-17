package broker

import (
	pb "Tolstoy/proto"
	"log"
	"net"
	"encoding/binary"

	"google.golang.org/protobuf/proto"
)

//
func AckTypeFor(t pb.Type) pb.Type {
	switch t {
	case pb.Type_CONN_REQUEST:
		return pb.Type_ACK_CONN_REQUEST
	case pb.Type_DIS_CONN_REQUEST:
		return pb.Type_ACK_DIS_CONN_REQUEST
	case pb.Type_PUBLISH:
		return pb.Type_ACK_PUBLISH
	case pb.Type_SUBSCRIBE:
		return pb.Type_ACK_SUBSCRIBE
	case pb.Type_UNSUBSCRIBE:
		return pb.Type_ACK_UNSUBSCRIBE
	default:
		return pb.Type_UNKNOWN
	}
}

func ackPacket(packetConn net.Conn, packet *pb.Packet) {
	ackType := AckTypeFor(packet.Type)
	ackPacket := packet
	ackPacket.Type = ackType
	writePacket(packetConn , ackPacket)
	log.Printf("Acknowledged a %s packet",packet.Type)
}

func writePacket(packetConn net.Conn , packet *pb.Packet) (error) {
	data , err := proto.Marshal(packet)
	if err != nil {
		return err 
	}
	size := make([]byte , 4)
	binary.BigEndian.PutUint32(size , uint32(len(data)))
	_,err = packetConn.Write(size)

	if err != nil {
		return err
	}

	_,err = packetConn.Write(data)

	if err != nil {
		return err
	}

	return nil
}
