package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompressWriter(t *testing.T) {
	rr := httptest.NewRecorder()
	cw := newCompressWriter(rr)

	// Установка заголовка
	cw.WriteHeader(http.StatusOK)

	if contentEncoding := rr.Header().Get("Content-Encoding"); contentEncoding != "gzip" {
		t.Errorf("Expected Content-Encoding to be gzip, got %s", contentEncoding)
	}

	// Запись данных
	data := []byte("test data")
	_, err := cw.Write(data)
	if err != nil {
		t.Fatalf("Unexpected error writing data: %v", err)
	}

	// Закрытие writer'а
	err = cw.Close()
	if err != nil {
		t.Fatalf("Unexpected error closing compressWriter: %v", err)
	}

	// Проверка, что данные сжаты
	gzReader, err := gzip.NewReader(rr.Body)
	if err != nil {
		t.Fatalf("Unexpected error creating gzip reader: %v", err)
	}
	defer gzReader.Close()

	uncompressedData := &bytes.Buffer{}
	_, err = uncompressedData.ReadFrom(gzReader)
	if err != nil {
		t.Fatalf("Unexpected error reading uncompressed data: %v", err)
	}

	if uncompressedData.String() != string(data) {
		t.Errorf("Expected uncompressed data to be %s, got %s", data, uncompressedData.String())
	}
}

func TestCompressReader(t *testing.T) {
	// Подготовка сжатых данных
	data := []byte("test data")
	var compressedData bytes.Buffer
	gzWriter := gzip.NewWriter(&compressedData)
	_, err := gzWriter.Write(data)
	if err != nil {
		t.Fatalf("Unexpected error writing data: %v", err)
	}
	gzWriter.Close()

	// Создание compressReader
	reader := io.NopCloser(bytes.NewReader(compressedData.Bytes()))
	cr, err := newCompressReader(reader)
	if err != nil {
		t.Fatalf("Unexpected error creating compressReader: %v", err)
	}

	// Чтение декомпрессированных данных
	uncompressedData := &bytes.Buffer{}
	_, err = io.Copy(uncompressedData, cr)
	if err != nil {
		t.Fatalf("Unexpected error reading uncompressed data: %v", err)
	}
	cr.Close()

	if uncompressedData.String() != string(data) {
		t.Errorf("Expected uncompressed data to be %s, got %s", data, uncompressedData.String())
	}
}

func TestCompressMiddlewareWithGzipRequest(t *testing.T) {
	data := []byte("test data")
	var compressedData bytes.Buffer
	gzWriter := gzip.NewWriter(&compressedData)
	_, err := gzWriter.Write(data)
	if err != nil {
		t.Fatalf("Unexpected error writing data: %v", err)
	}
	gzWriter.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Unexpected error reading request body: %v", err)
		}
		defer r.Body.Close()

		if string(body) != string(data) {
			t.Errorf("Expected request body to be %s, got %s", data, body)
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(compressedData.Bytes()))
	req.Header.Set("Content-Encoding", "gzip")
	rr := httptest.NewRecorder()

	middleware := CompressMiddleware(handler)
	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}
}
