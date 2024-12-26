package httpRedirectToHttpsTest

import (
	"testing"
	"net"
	"time"
	"crypto/tls"
	"crypto/ecdsa"
	"math/big"
	"crypto/x509"
	"crypto/x509/pkix"
	"bytes"
	"encoding/pem"
	"crypto/rand"
	"crypto/elliptic"
	"net/http"
	"io"
	"golang.org/x/net/http2"
	"github.com/bronze1man/httpRedirectToHttps"
)

func TestExample(t *testing.T) {
	s,err:=net.Listen("tcp",":4002")
	if err!=nil{
		panic(err)
	}
	defer s.Close()
	cert:= _MustGenSelfSignedCert()
	tlsConfig:=&tls.Config{
		GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error){
			return cert,nil
		},
		NextProtos: []string{"h2"},
	}
	ln2 := httpRedirectToHttps.NewListener(httpRedirectToHttps.NewListenerRequest{
		Ln: s,
		Cnf: tlsConfig,
	})
	server:=http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hello world from https"))
		}),
	}
	server.Handler = httpRedirectToHttps.NewHandler(server.Handler)
	go server.Serve(ln2)
	defer server.Close()

	func(){
		c:=http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		resp,err:=c.Get("http://127.0.0.1:4002/abc")
		if err!=nil{
			panic(err)
		}
		defer resp.Body.Close()
		//fmt.Println(resp.StatusCode,resp.Header.Get("Location"))
		ok(resp.StatusCode==307)
		ok(resp.Header.Get("Location")=="https://127.0.0.1:4002/abc")
	}()
	func(){
		c:=http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: &http2.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
					//NextProtos: []string{"h2"},
				},
			},
		}
		resp,err:=c.Get("https://127.0.0.1:4002/abc")
		if err!=nil{
			panic(err)
		}
		defer resp.Body.Close()
		//fmt.Println(resp.StatusCode,resp.Proto)
		ok(resp.StatusCode==200)
		b,err:=io.ReadAll(resp.Body)
		ok(bytes.Equal(b,[]byte("hello world from https")))
		ok(resp.Proto=="HTTP/2.0")
	}()
}
func ok(b bool){
	if b==false{
		panic("fail")
	}
}

func _MustGenSelfSignedCert() (*tls.Certificate) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		panic(err)
	}
	startTime := time.Now()
	notBefore := startTime.Truncate(time.Hour*24).Add(-24*time.Hour*365)
	notAfter := startTime.Truncate(time.Hour*24).Add(24*time.Hour*365)
	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{
			CommonName: "self-signed certificate",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	var (
		certBuf       = &bytes.Buffer{}
		privateKeyBuf = &bytes.Buffer{}
	)
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, privateKey.Public(), privateKey)
	if err != nil {
		panic(err)
	}
	err = pem.Encode(certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	if err != nil {
		panic(err)
	}
	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		panic(err)
	}
	err = pem.Encode(privateKeyBuf, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privateKeyBytes})
	if err != nil {
		panic(err)
	}
	tlsCert, err := tls.X509KeyPair(certBuf.Bytes(), privateKeyBuf.Bytes())
	if err != nil {
		panic(err)
	}
	return &tlsCert
}