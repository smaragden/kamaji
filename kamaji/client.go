package kamaji

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

type Client struct {
	net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
}

func NewClient(conn net.Conn) *Client {
	c := new(Client)
	c.Conn = conn
	c.reader = bufio.NewReader(c.Conn)
	c.writer = bufio.NewWriter(c.Conn)
	return c
}

// Takes a byte slice and prefixes it with the length of the incoming data.
func (c *Client) SendMessage(message []byte) (int, error) {
	defer c.writer.Flush()
	var message_size uint16 = uint16(len(message))
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, message_size)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
		return 0, err
	}
	var message_buffer []byte
	message_buffer = buf.Bytes()
	message_buffer = append(message_buffer[:], message[:]...)

	n, err := c.writer.Write(message_buffer)
	if err != nil {
		//log.WithFields(log.Fields{"module": "clientmanager"}).Error(err)
		return 0, err
	}
	return n, nil
}

// Reads data from socket and extract the message.
func (c *Client) ReadMessage() ([]byte, error) {
	// Get the size of the message
	message_size_bin := make([]byte, 2)
	n, err := c.reader.Read(message_size_bin)
	if err != nil {
		fmt.Println("client read:", err)
		return nil, err
	}
	buf := bytes.NewReader(message_size_bin)
	var message_size uint16
	err = binary.Read(buf, binary.LittleEndian, &message_size)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
		return nil, err
	}

	// Read the message
	var message_data []byte
	for bytes_read := uint16(0); bytes_read < message_size; {
		tmp := make([]byte, message_size - bytes_read)
		n, err = c.reader.Read(tmp)
		message_data = append(message_data[:], tmp[:]...)
		if err != nil {
			fmt.Println("client read:", err)
			return nil, err
		}
		bytes_read += uint16(n)
	}
	return message_data, nil
}
