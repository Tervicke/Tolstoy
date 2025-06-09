package broker

//All the code for persistance storage
import (
	"encoding/binary"
	"io"
	"os"

	pb "github.com/Tervicke/Tolstoy/internal/proto"
	"google.golang.org/protobuf/proto"
)

//append the record to the filename
func AppendRecord(filename string , record *pb.Record) ( int64 , error){
	//import from a binary buffer 
	//read and write a sample .stream file and write a .index file
	/*
	now := time.Now()
	sample := pb.Record{
		Timestamp:now.Unix(),
		Payload:[]byte("3rd message this is a long message"),
	}
	*/
	f , err := os.OpenFile(filename , os.O_APPEND | os.O_CREATE | os.O_WRONLY , 0644) 
	if err != nil {
		return -1 , err 
	} 
	defer f.Close()
	d , err := proto.Marshal(record)
	if err != nil {
		return -1 , err
	} 
	buf := make([]byte , 8)
	binary.BigEndian.PutUint64(buf , uint64(len(d)))
	pos , err := f.Seek(0 , io.SeekEnd)
	_ , err = f.Write(buf)
	if err != nil {
		return -1 , err
	}
	_ , err = f.Write(d)
	if err != nil {
		return -1 , err
	} 
	_ , err = f.Seek(0 , io.SeekEnd)
	if err != nil {
		return -1 , err 
	}
	return pos , nil 
}

//read a record at a particular position
func readFromRecord(filename string , pos int64 , ) ( *pb.Record , error){
	f , err := os.OpenFile(filename , os.O_CREATE | os.O_APPEND | os.O_RDWR , 0644 )
	if err != nil {
		return nil , err
	}
	_,err = f.Seek(pos, io.SeekStart)
	if err != nil {
		return nil,err
	}
	sizebuf := make([]byte , 8)
	_ , err = f.Read(sizebuf)
	if err != nil {
		return nil , err
	}
	size := binary.BigEndian.Uint64(sizebuf)
	databuf := make([]byte,size) 
	_ , err = f.Read(databuf)
	if err != nil {
		return nil , err
	}
	record := &pb.Record{} 
	err = proto.Unmarshal(databuf,record)
	if err != nil {
		return nil , err
	}
	return record,nil
}

//get the last offset a particular file

func getLastOffset(indexFileName string) (int64 , error) {
	indexFile , err := os.OpenFile(indexFileName, os.O_CREATE | os.O_RDONLY , 0644 )
	if err != nil {
		return -1 , err 
	}
	curOffsetBuf := make([]byte , 8)
	indexFile.Seek(0, io.SeekStart)
	indexFile.Read(curOffsetBuf)
	var cur_offset uint64 = uint64(binary.BigEndian.Uint64(curOffsetBuf))
	return int64(cur_offset), nil
}

func updateOffset(indexFileName string , offset int64) (error){
	indexFile , err := os.OpenFile(indexFileName, os.O_CREATE | os.O_RDWR , 0644 )
	if err != nil {
		return err 
	}
	offsetBuf := make([]byte , 8)
	binary.BigEndian.PutUint64(offsetBuf, uint64(offset))
	//seek and write at the start
	indexFile.Seek(0,io.SeekStart)
	_ , err = indexFile.Write(offsetBuf)
	if err!= nil {
		return err
	}
	return nil
}

func appendIndex(indexFileName string , pos int64) (error){
	indexFile , err := os.OpenFile(indexFileName, os.O_CREATE | os.O_RDWR, 0644 )
	if err != nil {
		return err
	}
	_ , err = indexFile.Seek(0,io.SeekEnd)
	posbuf := make([]byte, 8)
	indexFile.Seek(0,io.SeekEnd)
	binary.BigEndian.PutUint64(posbuf , uint64(pos))
	_ , err = indexFile.Write(posbuf)
	if err != nil {
		return err
	}
	return nil
}

func getOffsetPos(filename string , offset int64)(int64 , error){
	var posInIndex int64 = 8 + (offset * 8)
	indexFile , err := os.OpenFile(filename, os.O_CREATE | os.O_RDWR , 0644 )
	if err != nil {
		return -1 , err
	}
	posbuf := make([]byte , 8)
	indexFile.Seek(posInIndex,io.SeekStart)
	indexFile.Read(posbuf)
	pos := binary.BigEndian.Uint64(posbuf)
	return int64(pos) , nil
}
