package main

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/heroku/log-iss/Godeps/_workspace/src/github.com/bmizerany/lpx"
)

const (
	logplexDefaultHost = "host" // https://github.com/heroku/logplex/blob/master/src/logplex_http_drain.erl#L443
)

func fix(r io.Reader, remoteAddr string, logplexDrainToken string) ([]byte, error) {
	nilVal := []byte(`- `)

	var messageWriter bytes.Buffer
	var messageLenWriter bytes.Buffer

	readCopy := new(bytes.Buffer)

	lp := lpx.NewReader(bufio.NewReader(io.TeeReader(r, readCopy)))
	for lp.Next() {
		header := lp.Header()

		// LEN SP PRI VERSION SP TIMESTAMP SP HOSTNAME SP APP-NAME SP PROCID SP MSGID SP STRUCTURED-DATA MSG
		messageWriter.Write(header.PrivalVersion)
		messageWriter.WriteString(" ")
		messageWriter.Write(header.Time)
		messageWriter.WriteString(" ")
		if string(header.Hostname) == logplexDefaultHost && logplexDrainToken != "" {
			messageWriter.WriteString(logplexDrainToken)
		} else {
			messageWriter.Write(header.Hostname)
		}
		messageWriter.WriteString(" ")
		messageWriter.Write(header.Name)
		messageWriter.WriteString(" ")
		messageWriter.Write(header.Procid)
		messageWriter.WriteString(" ")
		messageWriter.Write(header.Msgid)
		messageWriter.WriteString(" [origin ip=\"")
		messageWriter.WriteString(remoteAddr)
		messageWriter.WriteString("\"]")

		b := lp.Bytes()
		if len(b) >= 2 && bytes.Equal(b[0:2], nilVal) {
			messageWriter.Write(b[1:])
		} else if len(b) > 0 {
			if b[0] != '[' {
				messageWriter.WriteString(" ")
			}
			messageWriter.Write(b)
		}

		messageLenWriter.WriteString(strconv.Itoa(messageWriter.Len()))
		messageLenWriter.WriteString(" ")
		messageWriter.WriteTo(&messageLenWriter)
	}

	if lp.Err() != nil {
		return nil, lp.Err()
	}

	fullMessage, err := ioutil.ReadAll(&messageLenWriter)
	if err != nil {
		return nil, err
	}

	return fullMessage, nil
}
