package bouteitech

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"io"
	"log"
	"net"
	"net/http"
)

func HandleConn(c net.Conn, signer *Signer) {
	// CONNECT メソッドによるトンネルの確立
	defer func() {
		_ = c.Close()
	}()
	br := bufio.NewReader(c)
	req, err := http.ReadRequest(br)
	if err != nil {
		log.Println("read request", err)
		return
	}
	if req.Method != http.MethodConnect {
		http.Error(
			&flushWriter{c}, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed,
		)
		return
	}
	_, err = io.WriteString(c, "HTTP/1.1 200 Connection established\r\n\r\n")
	if err != nil {
		log.Println("write response", err)
		return
	}
	log.Printf("[start] %s from %s", req.Host, remoteIP(req.Host))

	// クライアント(下流)との通信
	host, port, err := net.SplitHostPort(req.Host)
	if err != nil {
		log.Println("split host port", err)
		return
	}
	log.Printf("host: %s, port: %s", host, port)
	cert, err := signer.GetDomainCertificate(host)
	if err != nil {
		log.Println("get domain certificate", err)
		return
	}
	srvTLS := tls.Server(c, &tls.Config{
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return cert, nil
		},
		NextProtos: []string{"http/1.1"},
		MinVersion: tls.VersionTLS12,
	})
	if err := srvTLS.Handshake(); err != nil {
		log.Println("tls handshake", err)
		return
	}
	defer func() {
		_ = srvTLS.Close()
	}()

	// サーバー(上流)との通信
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		log.Println("load system cert pool", err)
		return
	}
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	upConn, err := tls.Dial("tcp", req.Host, &tls.Config{
		ServerName: host,
		RootCAs:    rootCAs,
		NextProtos: []string{"http/1.1"},
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		log.Println("dial upstream", err)
		return
	}
	defer func() {
		_ = upConn.Close()
	}()

	// 下流と上流の中継
	var reqBytes, resBytes int64
	done := make(chan struct{}, 2)

	go func() {
		defer func() { done <- struct{}{} }()
		tee := io.TeeReader(srvTLS, countWriter{&resBytes})
		if _, err := io.Copy(upConn, tee); err != nil {
			log.Printf("C->S copy error: %v", err)
			return
		}
		log.Println("C->S copy finished")
		_ = upConn.CloseWrite()
	}()

	go func() {
		defer func() { done <- struct{}{} }()
		tee := io.TeeReader(upConn, countWriter{&reqBytes})
		if _, err := io.Copy(srvTLS, tee); err != nil {
			log.Printf("S->C copy error: %v", err)
			return
		}
		log.Println("S->C copy finished")
		_ = srvTLS.CloseWrite()
	}()

	<-done
	<-done

	log.Printf(
		"[end] %s from %s - sent: %d bytes, received: %d bytes",
		req.Host, remoteIP(req.Host), reqBytes, resBytes,
	)
}

func remoteIP(host string) net.IP {
	// ポート番号が含まれている場合（例: "1.2.3.4:443"）を考慮して分割
	h, _, err := net.SplitHostPort(host)
	if err != nil {
		// ポートがない場合はそのままの文字列を使用
		h = host
	}

	// 文字列が IPv4/IPv6 アドレスとして妥当かパースする
	return net.ParseIP(h)
}
