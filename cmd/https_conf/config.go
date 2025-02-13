// httpsconf - пакет конфигурации https подключения
package httpsconf

import (
	"os"
	"os/signal"
	"syscall"
)

var (
	// CertFile - сертификационный файл для https подключения
	CertFile = "cert.pem"
	// KeyFile - ключ для https подключения
	KeyFile = "key.pem"
)

// MakeSigChan создает канал, принимающий сигналы
func MakeSigChan() chan os.Signal {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	return sigCh
}
