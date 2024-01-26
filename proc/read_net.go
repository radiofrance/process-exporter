package proc

// This is the code that reads the /proc/net/* files, comming from the
// prometheus/procfs package.
// For a reason I don't understand, the prometheus/procfs package does not
// expose newNetIPSocket, so I had to copy it here.
//
// The following code is under the Apache License 2.0, and was a bit modified
// to remove unused code
//
// See https://github.com/prometheus/procfs/blob/9fdfbe8443465c0da56ca44ff5b1d16250e6a396/net_ip_socket.go
// for the original code.

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	// readLimit is used by io.LimitReader while reading the content of the
	// /proc/net/udp{,6} files. The number of lines inside such a file is dynamic
	// as each line represents a single used socket.
	// In theory, the number of available sockets is 65535 (2^16 - 1) per IP.
	// With e.g. 150 Byte per line and the maximum number of 65535,
	// the reader needs to handle 150 Byte * 65535 =~ 10 MB for a single IP.
	readLimit = 4294967296 // Byte -> 4 GiB
)

// This contains generic data structures for both udp and tcp sockets.
type (
	// NetIPSocket represents the contents of /proc/net/{t,u}dp{,6} file without the header.
	NetIPSocket []NetIPSocketLine

	// NetIPSocketLine represents the fields parsed from a single line
	// in /proc/net/{t,u}dp{,6}. Fields which are not used by IPSocket are skipped.
	// Drops is non-nil for udp{,6}, but nil for tcp{,6}.
	// For the proc file format details, see https://linux.die.net/man/5/proc.
	NetIPSocketLine struct {
		Sl uint64
		// LocalAddr net.IP
		// LocalPort uint64
		// RemAddr   net.IP
		// RemPort   uint64
		St      uint64
		TxQueue uint64
		RxQueue uint64
		UID     uint64
		Inode   uint64
		Drops   *uint64
	}
)

var (
	ErrFileParse = errors.New("Error Parsing File")
	ErrFileRead  = errors.New("Error Reading File")
)

func NewNetIPSocket(file string) (NetIPSocket, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var netIPSocket NetIPSocket
	isUDP := strings.Contains(file, "udp")

	lr := io.LimitReader(f, readLimit)
	s := bufio.NewScanner(lr)
	s.Scan() // skip first line with headers
	for s.Scan() {
		fields := strings.Fields(s.Text())
		line, err := parseNetIPSocketLine(fields, isUDP)
		if err != nil {
			return nil, err
		}
		netIPSocket = append(netIPSocket, *line)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return netIPSocket, nil
}

// parseNetIPSocketLine parses a single line, represented by a list of fields.
func parseNetIPSocketLine(fields []string, isUDP bool) (*NetIPSocketLine, error) {
	line := &NetIPSocketLine{}
	if len(fields) < 10 {
		return nil, fmt.Errorf(
			"%w: Less than 10 columns found %q",
			ErrFileParse,
			strings.Join(fields, " "),
		)
	}
	var err error // parse error

	// sl
	s := strings.Split(fields[0], ":")
	if len(s) != 2 {
		return nil, fmt.Errorf("%w: Unable to parse sl field in line %q", ErrFileParse, fields[0])
	}

	if line.Sl, err = strconv.ParseUint(s[0], 0, 64); err != nil {
		return nil, fmt.Errorf("%s: Unable to parse sl field in %q: %w", ErrFileParse, line.Sl, err)
	}

	// st
	if line.St, err = strconv.ParseUint(fields[3], 16, 64); err != nil {
		return nil, fmt.Errorf("%s: Cannot parse st value in %q: %w", ErrFileParse, line.St, err)
	}

	// tx_queue and rx_queue
	q := strings.Split(fields[4], ":")
	if len(q) != 2 {
		return nil, fmt.Errorf(
			"%w: Missing colon for tx/rx queues in socket line %q",
			ErrFileParse,
			fields[4],
		)
	}
	if line.TxQueue, err = strconv.ParseUint(q[0], 16, 64); err != nil {
		return nil, fmt.Errorf("%s: Cannot parse tx_queue value in %q: %w", ErrFileParse, line.TxQueue, err)
	}
	if line.RxQueue, err = strconv.ParseUint(q[1], 16, 64); err != nil {
		return nil, fmt.Errorf("%s: Cannot parse trx_queue value in %q: %w", ErrFileParse, line.RxQueue, err)
	}

	// uid
	if line.UID, err = strconv.ParseUint(fields[7], 0, 64); err != nil {
		return nil, fmt.Errorf("%s: Cannot parse UID value in %q: %w", ErrFileParse, line.UID, err)
	}

	// inode
	if line.Inode, err = strconv.ParseUint(fields[9], 0, 64); err != nil {
		return nil, fmt.Errorf("%s: Cannot parse inode value in %q: %w", ErrFileParse, line.Inode, err)
	}

	// drops
	if isUDP {
		drops, err := strconv.ParseUint(fields[12], 0, 64)
		if err != nil {
			return nil, fmt.Errorf("%s: Cannot parse drops value in %q: %w", ErrFileParse, drops, err)
		}
		line.Drops = &drops
	}

	return line, nil
}
