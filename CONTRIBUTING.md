# Руководство разработчика IPXpress

## Содержание

1. [Начало работы](#начало-работы)
2. [Структура кода](#структура-кода)
3. [Разработка новых функций](#разработка-новых-функций)
4. [Тестирование](#тестирование)
5. [Оптимизация производительности](#оптимизация-производительности)
6. [Отладка](#отладка)
7. [Деплой](#деплой)

## Начало работы

### Требования

- Go 1.24+
- libvips 8.15+
- Git

### Установка libvips

#### Ubuntu/Debian

```bash
sudo apt-get update
sudo apt-get install libvips-dev
```

#### macOS

```bash
brew install vips
```

### Клонирование и запуск

```bash
# Клонировать репозиторий
git clone https://github.com/vladislavsavi/IPXpress.git
cd IPXpress

# Установить зависимости
go mod download

# Запустить тесты
go test ./pkg/ipxpress/... -v

# Собрать
go build ./cmd/ipxpress

# Запустить
./ipxpress -addr :8080
```

## Структура кода

### Основные компоненты

```
pkg/ipxpress/
├── cache.go        # Система кеширования (интерфейс + реализация)
├── config.go       # Конфигурация сервиса
├── fetcher.go      # HTTP клиент для загрузки изображений
├── format.go       # Работа с форматами изображений
├── ipxpress.go     # Ядро обработки (Processor)
├── params.go       # Парсинг параметров HTTP запроса
├── server.go       # HTTP handler
└── *_test.go       # Тесты
```

### Принципы организации кода

1. **Разделение ответственности:** Каждый файл отвечает за свою область
2. **Интерфейсы:** Используются для абстракций (Cache, etc.)
3. **Chainable API:** Processor использует fluent interface
4. **Иммутабельность:** Конфигурация создается один раз
5. **Явное управление ресурсами:** Close() для освобождения памяти

## Разработка новых функций

### Добавление нового формата изображения

Пример: добавим поддержку AVIF.

#### 1. Обновить `format.go`

```go
const (
    // ... существующие форматы
    FormatAVIF Format = "avif"
)

func (f Format) ContentType() string {
    switch f {
    // ... существующие случаи
    case FormatAVIF:
        return "image/avif"
    // ...
    }
}

func (f Format) IsValid() bool {
    switch f {
    // ... существующие случаи
    case FormatAVIF:
        return true
    // ...
    }
}

func DetectFormat(data []byte) Format {
    // ... существующие проверки
    
    // AVIF: проверка ftypavif
    if len(data) >= 12 && 
       data[4] == 0x66 && data[5] == 0x74 && 
       data[6] == 0x79 && data[7] == 0x70 {
        return FormatAVIF
    }
    
    return ""
}
```

#### 2. Обновить `ipxpress.go`

```go
func (p *Processor) ToBytes(format Format, quality int) ([]byte, error) {
    // ... существующие проверки
    
    switch format {
    // ... существующие случаи
    
    case FormatAVIF:
        // Проверить доступность AVIF в libvips
        params := vips.NewAvifExportParams()
        params.Quality = quality
        params.Speed = 5 // 0-9, меньше = лучше качество
        buf, _, err := p.img.ExportAvif(params)
        if err != nil {
            return nil, fmt.Errorf("failed to encode AVIF: %w", err)
        }
        return buf, nil
    
    // ...
    }
}
```

#### 3. Добавить тесты

```go
// ipxpress_test.go
func TestProcessorAVIF(t *testing.T) {
    // Создать тестовое изображение
    img := createTestImage(100, 50)
    
    proc := New().FromBytes(img).Resize(50, 0)
    
    out, err := proc.ToBytes(FormatAVIF, 85)
    if err != nil {
        t.Fatalf("encode AVIF: %v", err)
    }
    
    // Проверить формат
    format := DetectFormat(out)
    if format != FormatAVIF {
        t.Errorf("expected AVIF, got %s", format)
    }
}
```

### Добавление новой трансформации

Пример: добавим операцию crop (обрезка).

#### 1. Добавить метод в `Processor`

```go
// ipxpress.go

// Crop crops the image to the specified rectangle.
func (p *Processor) Crop(x, y, width, height int) *Processor {
    if p.err != nil {
        return p
    }
    if p.img == nil {
        p.err = errors.New("no image loaded")
        return p
    }
    
    // Валидация
    if x < 0 || y < 0 || width <= 0 || height <= 0 {
        p.err = errors.New("invalid crop dimensions")
        return p
    }
    
    // Обрезка через libvips
    if err := p.img.ExtractArea(x, y, width, height); err != nil {
        p.err = fmt.Errorf("failed to crop: %w", err)
        return p
    }
    
    return p
}
```

#### 2. Обновить параметры

```go
// params.go

type ProcessingParams struct {
    URL     string
    Width   int
    Height  int
    Quality int
    Format  Format
    // Новые параметры crop
    CropX      int
    CropY      int
    CropWidth  int
    CropHeight int
}

func ParseProcessingParams(r *http.Request) *ProcessingParams {
    q := r.URL.Query()
    
    params := &ProcessingParams{
        // ... существующие поля
        CropX:      parseInt(q.Get("crop_x")),
        CropY:      parseInt(q.Get("crop_y")),
        CropWidth:  parseInt(q.Get("crop_w")),
        CropHeight: parseInt(q.Get("crop_h")),
    }
    
    return params
}
```

#### 3. Использовать в server.go

```go
// server.go

func (h *Handler) processImage(imageData []byte, params *ProcessingParams) *CacheEntry {
    proc := New().FromBytes(imageData)
    
    // Применить crop если указан
    if params.CropWidth > 0 && params.CropHeight > 0 {
        proc = proc.Crop(params.CropX, params.CropY, 
                        params.CropWidth, params.CropHeight)
    }
    
    // Применить resize
    if params.Width > 0 || params.Height > 0 {
        proc = proc.Resize(params.Width, params.Height)
    }
    
    // ... остальная логика
}
```

### Добавление middleware

Пример: добавим логирование запросов.

```go
// middleware.go (новый файл)

package ipxpress

import (
    "log"
    "net/http"
    "time"
)

// LoggingMiddleware логирует все HTTP запросы.
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Wrapper для захвата статус кода
        wrapped := &statusWriter{ResponseWriter: w, status: 200}
        
        next.ServeHTTP(wrapped, r)
        
        duration := time.Since(start)
        log.Printf("[%s] %s %s - %d (%v)",
            r.Method, r.URL.Path, r.RemoteAddr,
            wrapped.status, duration)
    })
}

type statusWriter struct {
    http.ResponseWriter
    status int
}

func (w *statusWriter) WriteHeader(status int) {
    w.status = status
    w.ResponseWriter.WriteHeader(status)
}

// MetricsMiddleware собирает метрики производительности.
func MetricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Инкремент счетчика запросов
        // Замер latency
        // Отправка в систему мониторинга
        
        next.ServeHTTP(w, r)
    })
}
```

Использование в `main.go`:

```go
handler := ipxpress.NewHandler(config)

// Обернуть в middleware
wrappedHandler := ipxpress.LoggingMiddleware(handler)
wrappedHandler = ipxpress.MetricsMiddleware(wrappedHandler)

mux.Handle("/ipx/", http.StripPrefix("/ipx/", wrappedHandler))
```

## Тестирование

### Запуск тестов

```bash
# Все тесты
go test ./pkg/ipxpress/... -v

# С покрытием
go test ./pkg/ipxpress/... -cover -coverprofile=coverage.out

# Просмотр покрытия
go tool cover -html=coverage.out
```

### Написание тестов

#### Unit тест для Processor

```go
func TestProcessorResize(t *testing.T) {
    // Создать тестовое изображение
    img := createTestRGBA(200, 100)
    
    proc := New().FromBytes(img)
    
    // Применить resize
    proc = proc.Resize(100, 0)
    
    // Проверить отсутствие ошибок
    if err := proc.Err(); err != nil {
        t.Fatalf("resize failed: %v", err)
    }
    
    // Закодировать и проверить размер
    out, err := proc.ToBytes(FormatPNG, 85)
    proc.Close()
    
    if err != nil {
        t.Fatalf("encode failed: %v", err)
    }
    
    // Декодировать результат
    decoded, _, err := image.Decode(bytes.NewReader(out))
    if err != nil {
        t.Fatalf("decode result: %v", err)
    }
    
    bounds := decoded.Bounds()
    if bounds.Dx() != 100 || bounds.Dy() != 50 {
        t.Errorf("unexpected size: %dx%d", bounds.Dx(), bounds.Dy())
    }
}
```

#### Integration тест

```go
func TestServerIntegration(t *testing.T) {
    // Создать тестовый image server
    imgServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Вернуть тестовое изображение
        w.Header().Set("Content-Type", "image/jpeg")
        jpeg.Encode(w, createTestImage(1000, 500), &jpeg.Options{Quality: 90})
    }))
    defer imgServer.Close()
    
    // Создать IPXpress handler
    handler := NewHandler(DefaultConfig())
    server := httptest.NewServer(handler)
    defer server.Close()
    
    // Сделать запрос
    resp, err := http.Get(server.URL + "?url=" + 
        url.QueryEscape(imgServer.URL+"/test.jpg") + "&w=500")
    
    if err != nil {
        t.Fatalf("request failed: %v", err)
    }
    defer resp.Body.Close()
    
    // Проверить статус
    if resp.StatusCode != http.StatusOK {
        t.Errorf("expected 200, got %d", resp.StatusCode)
    }
    
    // Проверить Content-Type
    ct := resp.Header.Get("Content-Type")
    if !strings.HasPrefix(ct, "image/") {
        t.Errorf("invalid content-type: %s", ct)
    }
}
```

#### Benchmark тест

```go
func BenchmarkProcessorResize(b *testing.B) {
    img := createTestRGBA(2000, 1000)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        proc := New().FromBytes(img).Resize(800, 400)
        out, _ := proc.ToBytes(FormatJPEG, 85)
        proc.Close()
        _ = out
    }
}

func BenchmarkServerHandler(b *testing.B) {
    handler := NewHandler(DefaultConfig())
    
    // Создать тестовый запрос
    req := httptest.NewRequest("GET", "/ipx/?url=...&w=800", nil)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        w := httptest.NewRecorder()
        handler.ServeHTTP(w, req)
    }
}
```

## Оптимизация производительности

### Профилирование

#### CPU профилирование

```bash
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof
```

В интерактивном режиме:
```
(pprof) top10      # Топ 10 функций
(pprof) list ProcessImage  # Детали функции
(pprof) web        # Визуализация
```

#### Memory профилирование

```bash
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

### Советы по оптимизации

1. **Используйте pprof во время нагрузочного тестирования:**

```go
import _ "net/http/pprof"

func main() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // ... основной код
}
```

2. **Минимизируйте аллокации:**
   - Переиспользуйте буферы
   - Избегайте string concatenation в горячих путях
   - Используйте sync.Pool для временных объектов

3. **Оптимизация libvips:**
   - Настройте ConcurrencyLevel под ваше железо
   - Экспериментируйте с MaxCacheMem
   - Используйте векторные операции (автоматически)

## Отладка

### Логирование

#### Включить VIPS logs

```go
// main.go
vips.LoggingSettings(func(domain string, level vips.LogLevel, msg string) {
    log.Printf("[%s:%s] %s", domain, level, msg)
}, vips.LogLevelInfo)
```

#### Structured logging

```go
import "github.com/rs/zerolog/log"

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    log.Info().
        Str("url", r.URL.String()).
        Str("remote", r.RemoteAddr).
        Msg("processing request")
    
    // ... обработка
    
    log.Info().
        Str("url", r.URL.String()).
        Int("status", statusCode).
        Dur("duration", duration).
        Msg("request completed")
}
```

### Debugging с Delve

```bash
# Установить delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Запустить с отладчиком
dlv debug ./cmd/ipxpress

# В delve:
(dlv) break server.go:123
(dlv) continue
```

## Деплой

### Docker

```dockerfile
# Dockerfile
FROM golang:1.24-alpine AS builder

# Установить зависимости для сборки
RUN apk add --no-cache vips-dev build-base

WORKDIR /app
COPY go.* ./
RUN go mod download

COPY . .
RUN go build -o ipxpress ./cmd/ipxpress

FROM alpine:latest
RUN apk add --no-cache vips

COPY --from=builder /app/ipxpress /usr/local/bin/

EXPOSE 8080
CMD ["ipxpress", "-addr", ":8080"]
```

Сборка и запуск:

```bash
docker build -t ipxpress:latest .
docker run -p 8080:8080 ipxpress:latest
```

### Kubernetes

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ipxpress
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ipxpress
  template:
    metadata:
      labels:
        app: ipxpress
    spec:
      containers:
      - name: ipxpress
        image: ipxpress:latest
        ports:
        - containerPort: 8080
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "2000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: ipxpress-service
spec:
  selector:
    app: ipxpress
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

### Systemd (Linux)

```ini
# /etc/systemd/system/ipxpress.service
[Unit]
Description=IPXpress Image Processing Service
After=network.target

[Service]
Type=simple
User=ipxpress
WorkingDirectory=/opt/ipxpress
ExecStart=/opt/ipxpress/ipxpress -addr :8080
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

Управление:

```bash
sudo systemctl enable ipxpress
sudo systemctl start ipxpress
sudo systemctl status ipxpress
```

## Best Practices

1. **Всегда вызывайте Close()** после использования Processor
2. **Используйте context.Context** для отмены длительных операций
3. **Валидируйте входные данные** перед обработкой
4. **Логируйте ошибки** с достаточным контекстом
5. **Пишите тесты** для новых функций
6. **Документируйте API** в комментариях к коду
7. **Используйте семантическое версионирование** (SemVer)

## Дополнительные ресурсы

- [libvips документация](https://www.libvips.org/API/current/)
- [govips примеры](https://github.com/davidbyttow/govips)
- [Go testing best practices](https://go.dev/doc/tutorial/add-a-test)
