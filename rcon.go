package mcutil

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/mcstatus-io/mcutil/options"
)

var (
	defaultRCONOptions = options.RCON{
		Timeout: time.Second * 5,
	}
)

// RCONClient is a client for interacting with RCON and contains multiple methods
type RCONClient struct {
	conn        net.Conn
	Messages    chan string
	runTrigger  chan bool
	authSuccess bool
	requestID   int32
}

// NewRCON creates a new RCON client from the options parameter
func NewRCON() *RCONClient {
	return &RCONClient{
		conn:        nil,
		Messages:    make(chan string),
		runTrigger:  make(chan bool),
		authSuccess: false,
		requestID:   0,
	}
}

// Dial creates a new connection to the server
func (r *RCONClient) Dial(host string, port uint16, options ...options.RCON) error {
	opts := parseRCONOptions(options...)

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), opts.Timeout)

	if err != nil {
		return err
	}

	if err = conn.SetDeadline(time.Now().Add(opts.Timeout)); err != nil {
		return err
	}

	r.conn = conn

	return nil
}

// Login communicates authentication with the server using the plaintext password
func (r *RCONClient) Login(password string) error {
	if r.conn == nil {
		return ErrNotConnected
	}

	if r.authSuccess {
		return ErrAlreadyLoggedIn
	}

	// Login request packet
	// https://wiki.vg/RCON#3:_Login
	{
		buf := &bytes.Buffer{}

		// Length - int32
		if err := binary.Write(buf, binary.LittleEndian, int32(10+len(password))); err != nil {
			return err
		}

		// Request ID - int32
		if err := binary.Write(buf, binary.LittleEndian, int32(0)); err != nil {
			return err
		}

		// Type - int32
		if err := binary.Write(buf, binary.LittleEndian, int32(3)); err != nil {
			return err
		}

		// Payload - null-terminated string
		if _, err := buf.Write(append([]byte(password), 0x00)); err != nil {
			return err
		}

		// Padding - byte
		if err := buf.WriteByte(0x00); err != nil {
			return err
		}

		if _, err := io.Copy(r.conn, buf); err != nil {
			return err
		}
	}

	// Login response packet
	// https://wiki.vg/RCON#3:_Login
	{
		var packetLength int32

		// Length - int32
		{
			if err := binary.Read(r.conn, binary.LittleEndian, &packetLength); err != nil {
				return err
			}
		}

		// Request ID - int32
		{
			var requestID int32

			if err := binary.Read(r.conn, binary.LittleEndian, &requestID); err != nil {
				return err
			}

			if requestID == -1 {
				return ErrInvalidPassword
			} else if requestID != 0 {
				return fmt.Errorf("rcon: received unexpected request ID (expected=0, received=%d)", requestID)
			}
		}

		// Type - int32
		{
			var packetType int32

			if err := binary.Read(r.conn, binary.LittleEndian, &packetType); err != nil {
				return err
			}

			if packetType != 0x02 {
				return fmt.Errorf("rcon: received unexpected packet type (expected=0x02, received=0x%02X)", packetType)
			}
		}

		// Remaining bytes
		{
			data := make([]byte, packetLength-8)

			if _, err := r.conn.Read(data); err != nil {
				return err
			}
		}
	}

	r.authSuccess = true

	if err := r.conn.SetReadDeadline(time.Time{}); err != nil {
		return err
	}

	go (func() {
		for {
			// TODO figure out EOF issue, and how to not continuously loop with EOF errors when client is open

			err := r.readMessage()

			if err != nil {
				fmt.Println(err)
			}
		}
	})()

	return nil
}

// Run executes the command on the server but does not wait for a response
func (r *RCONClient) Run(command string) error {
	if r.conn == nil {
		return ErrNotConnected
	}

	if !r.authSuccess {
		return ErrNotAuthenticated
	}

	r.requestID++

	// Command packet
	// https://wiki.vg/RCON#2:_Command
	{
		buf := &bytes.Buffer{}

		// Length - int32
		if err := binary.Write(buf, binary.LittleEndian, int32(10+len(command))); err != nil {
			return err
		}

		// Request ID - int32
		if err := binary.Write(buf, binary.LittleEndian, r.requestID); err != nil {
			return err
		}

		// Type - int32
		if err := binary.Write(buf, binary.LittleEndian, int32(2)); err != nil {
			return err
		}

		// Payload - null-terminated string
		if _, err := buf.Write(append([]byte(command), 0x00)); err != nil {
			return err
		}

		if err := buf.WriteByte(0x00); err != nil {
			return err
		}

		if _, err := io.Copy(r.conn, buf); err != nil {
			return err
		}
	}

	return nil
}

// Close closes the connection to the server
func (r *RCONClient) Close() error {
	r.authSuccess = false
	r.requestID = 0

	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			return err
		}
	}

	r.conn = nil

	return nil
}

func (r *RCONClient) readMessage() error {
	// Command response packet
	// https://wiki.vg/RCON#0:_Command_response
	{
		var packetLength int32

		// Length - int32
		{
			// TODO convert to binary.Read() for the rest of the package
			data := make([]byte, 4)

			n, err := r.conn.Read(data)

			if err != nil {
				return err
			}

			if n < 4 {
				return nil
			}

			packetLength = int32(binary.LittleEndian.Uint16(data))
		}

		// Request ID - int32
		{
			data := make([]byte, 4)

			n, err := r.conn.Read(data)

			if err != nil {
				return err
			}

			if n < 4 {
				return nil
			}
		}

		// Type - int32
		{
			var packetType int32

			if err := binary.Read(r.conn, binary.LittleEndian, &packetType); err != nil {
				return err
			}

			if packetType != 2 {
				return fmt.Errorf("rcon: received unexpected packet type (expected=0x00, received=0x%02X)", packetType)
			}
		}

		// Payload - null-terminated string
		{
			data := make([]byte, packetLength-8)

			n, err := r.conn.Read(data)

			if err != nil {
				return err
			}

			if n < 1 {
				return nil
			}

			r.Messages <- string(data)
		}
	}

	return nil
}

func parseRCONOptions(opts ...options.RCON) options.RCON {
	if len(opts) < 1 {
		return defaultRCONOptions
	}

	return opts[0]
}
