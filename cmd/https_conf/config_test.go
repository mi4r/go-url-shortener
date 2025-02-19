package httpsconf

import (
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMakeSigChan(t *testing.T) {
	sigCh := MakeSigChan()
	assert.NotNil(t, sigCh, "Канал сигналов не должен быть nil")

	// Отправляем сигнал в канал после короткой задержки
	go func() {
		time.Sleep(100 * time.Millisecond)
		sigCh <- syscall.SIGTERM
	}()

	select {
	case sig := <-sigCh:
		assert.Equal(t, syscall.SIGTERM, sig, "Полученный сигнал должен быть SIGTERM")
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Таймаут: сигнал не был получен")
	}
}
