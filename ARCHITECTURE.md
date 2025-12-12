# Архитектура IPXpress

## Обзор

IPXpress — это высокопроизводительный HTTP-сервис для обработки изображений на Go с использованием libvips. Проект спроектирован с учетом масштабируемости, производительности и читаемости кода.

## Структура проекта

```
IPXpress/
├── cmd/
│   └── ipxpress/           # Точка входа приложения
│       └── main.go         # HTTP сервер с инициализацией libvips
├── pkg/
│   └── ipxpress/           # Основной пакет библиотеки
│       ├── cache.go        # Система кеширования
│       ├── config.go       # Конфигурация сервиса
│       ├── fetcher.go      # Загрузка изображений по URL
│       ├── format.go       # Работа с форматами изображений
│       ├── ipxpress.go     # Ядро обработки изображений (Processor)
│       ├── params.go       # Парсинг параметров запроса
│       ├── server.go       # HTTP обработчик запросов
│       ├── *_test.go       # Тесты
│       └── ...
├── static/                 # Статические файлы (если есть)
├── go.mod                  # Go модуль
├── Dockerfile              # Docker образ
└── README.md               # Документация проекта
```

## Компоненты системы

### 1. **Processor** (`ipxpress.go`)

Ядро системы обработки изображений. Использует chainable API для последовательного применения трансформаций.

**Основные возможности:**
- Загрузка изображений из байтов или io.Reader
- Изменение размера с сохранением пропорций (Lanczos3)
- Кодирование в различные форматы (JPEG, PNG, GIF, WebP)
- Автоматическое определение формата исходного изображения
- Управление памятью через libvips

**Пример использования:**
```go
proc := ipxpress.New().
    FromBytes(imageData).
    Resize(800, 600)

if err := proc.Err(); err != nil {
    // обработка ошибки
}

output, err := proc.ToBytes(ipxpress.FormatJPEG, 85)
proc.Close() // освобождение памяти
```

### 2. **Format** (`format.go`)

Модуль для работы с форматами изображений.

**Возможности:**
- Типизированные константы форматов (FormatJPEG, FormatPNG, FormatGIF, FormatWebP)
- Автоматическое определение формата по магическим байтам
- Получение MIME типа для HTTP заголовков
- Валидация форматов

**Примеры:**
```go
// Определение формата
format := ipxpress.DetectFormat(imageData)

// Парсинг строки формата
format := ipxpress.ParseFormat("jpeg") // returns FormatJPEG

// MIME тип
contentType := format.ContentType() // "image/jpeg"
```

### 3. **Cache** (`cache.go`)

Система кеширования ответов с TTL (Time To Live).

**Архитектура:**
- Интерфейс `Cache` для различных реализаций
- Реализация `InMemoryCache` с sync.RWMutex
- Автоматическая очистка устаревших записей
- Кеширование как успешных ответов, так и ошибок

**Структура записи:**
```go
type CacheEntry struct {
    ContentType string    // MIME тип ответа
    Data        []byte    // Данные изображения
    StatusCode  int       // HTTP статус
    ErrorMsg    string    // Сообщение об ошибке (если есть)
    Timestamp   time.Time // Время создания записи
}
```

### 4. **Fetcher** (`fetcher.go`)

Модуль загрузки изображений по URL.

**Возможности:**
- HTTP/HTTPS поддержка
- Connection pooling для высокой производительности
- Валидация URL
- Настраиваемые timeouts
- User-Agent для обхода базовых ограничений

**Конфигурация HTTP клиента:**
```go
- Timeout: 20 секунд
- MaxIdleConns: 500
- MaxIdleConnsPerHost: 100
- MaxConnsPerHost: 256
- DialTimeout: 5 секунд
- KeepAlive: 30 секунд
```

### 5. **Params** (`params.go`)

Парсинг и валидация параметров HTTP запроса.

**Структура:**
```go
type ProcessingParams struct {
    URL     string  // URL изображения
    Width   int     // Максимальная ширина
    Height  int     // Максимальная высота
    Quality int     // Качество (1-100)
    Format  Format  // Формат вывода
}
```

**Логика обработки:**
- Автоматическая валидация параметров
- Значение quality по умолчанию: 85
- Определение необходимости обработки
- Выбор формата вывода (original или указанный)

### 6. **Server** (`server.go`)

HTTP обработчик запросов к сервису.

**Архитектура Handler:**
```go
type Handler struct {
    cache           Cache           // Система кеширования
    fetcher         *Fetcher        // Загрузчик изображений
    config          *Config         // Конфигурация
    processingLimit chan struct{}   // Семафор для ограничения конкурентности
}
```

**Поток обработки запроса:**
1. Парсинг параметров запроса
2. Проверка кеша (быстрый путь)
3. Загрузка изображения (параллельно, I/O bound)
4. Обработка изображения (с ограничением конкурентности, CPU bound)
5. Кеширование результата
6. Отправка ответа клиенту

**Оптимизации:**
- Двухстадийная обработка: сначала I/O (параллельно), затем CPU (с семафором)
- Кеширование ошибок для предотвращения повторных обращений
- Connection pooling для исходящих HTTP запросов
- Оптимизированные HTTP заголовки (Cache-Control)

### 7. **Config** (`config.go`)

Конфигурация сервиса.

```go
type Config struct {
    CacheTTL        time.Duration // TTL кеша (30 сек)
    ProcessingLimit int           // Макс. одновременных обработок (256)
    CleanupInterval time.Duration // Интервал очистки кеша (30 сек)
}
```

## Потоки данных

### Обработка запроса

```
HTTP Request
    ↓
[ParseParams] → ProcessingParams
    ↓
[Cache Check] → Hit? → [Write Response]
    ↓ Miss
[Fetch Image] → imageData (parallel, no semaphore)
    ↓
[Acquire Semaphore] → limit concurrent processing
    ↓
[Process Image] → Processor chain
    ↓
[Encode Output] → output bytes
    ↓
[Release Semaphore]
    ↓
[Cache Result]
    ↓
[Write Response]
```

### Обработка изображения (Processor)

```
Image Bytes
    ↓
[Detect Format] → Format
    ↓
[Decode with libvips] → vips.ImageRef
    ↓
[Resize (optional)] → transformed ImageRef
    ↓
[Encode to Format] → output bytes
    ↓
[Close/Free Memory]
    ↓
Output Bytes
```

## Конкурентность и производительность

### Стратегия обработки

1. **Фаза I/O (без ограничений):**
   - Загрузка изображений по HTTP
   - Проверка кеша
   - Параллельная обработка множества запросов

2. **Фаза CPU (с семафором):**
   - Обработка изображений через libvips
   - Ограничение: 256 одновременных операций
   - Предотвращение перегрузки памяти

### Кеширование

- **Ключ кеша:** MD5(url|width|height|quality|format)
- **TTL:** 30 секунд (настраивается)
- **Очистка:** Периодическая (каждые 30 сек)
- **Хранение:** In-memory (быстрый доступ)

### Управление памятью

- Немедленное освобождение после обработки (`proc.Close()`)
- libvips настройки:
  - MaxCacheMem: 2048 MB
  - MaxCacheSize: 5000 изображений
  - ConcurrencyLevel: 0 (все ядра CPU)

## API

### HTTP Endpoint

**URL:** `/ipx/`

**Параметры запроса:**

| Параметр | Тип | Обязательный | Описание |
|----------|-----|--------------|----------|
| `url` | string | ✅ | URL изображения (HTTP/HTTPS) |
| `w` | int | ❌ | Максимальная ширина в пикселях |
| `h` | int | ❌ | Максимальная высота в пикселях |
| `quality` | int | ❌ | Качество сжатия (1-100, default: 85) |
| `format` | string | ❌ | Формат вывода (jpeg, png, gif, webp) |

**Примеры:**

```bash
# Изменение размера
GET /ipx/?url=https://example.com/image.jpg&w=800&h=600

# Конвертация в WebP
GET /ipx/?url=https://example.com/image.jpg&format=webp&quality=90

# Только изменение размера (сохранение формата)
GET /ipx/?url=https://example.com/image.png&w=500
```

## Расширение системы

### Добавление нового формата

1. Добавить константу в `format.go`:
```go
const FormatAVIF Format = "avif"
```

2. Обновить `ContentType()`:
```go
case FormatAVIF:
    return "image/avif"
```

3. Добавить обработку в `Processor.ToBytes()` (`ipxpress.go`)

### Замена системы кеширования

Реализовать интерфейс `Cache`:
```go
type RedisCache struct {
    client *redis.Client
}

func (c *RedisCache) Get(key string) (*CacheEntry, bool) { ... }
func (c *RedisCache) Set(key string, entry *CacheEntry) { ... }
func (c *RedisCache) Cleanup() { ... }
```

Использовать в Handler:
```go
handler := &Handler{
    cache: NewRedisCache(...),
    // ...
}
```

## Тестирование

Запуск тестов:
```bash
go test ./pkg/ipxpress/... -v
```

Покрытие кода:
```bash
go test ./pkg/ipxpress/... -cover
```

## Мониторинг и логирование

### Текущие логи
- libvips logs (уровень WARNING+)
- HTTP запросы (через стандартный log)

### Рекомендации для продакшена
- Добавить structured logging (zap, zerolog)
- Метрики Prometheus (latency, cache hit rate, error rate)
- Distributed tracing (OpenTelemetry)
- Health check endpoint (`/health`)

## Производительность

### Целевые показатели
- **Throughput:** 3000+ req/sec
- **Latency:** <50ms (для кешированных), <200ms (с обработкой)
- **Concurrency:** 256 одновременных обработок

### Оптимизации
- Connection pooling (500 idle connections)
- Response caching (30 sec TTL)
- Efficient memory management (immediate cleanup)
- Vector operations в libvips (SIMD)

## Безопасность

### Текущие меры
- Валидация URL (только HTTP/HTTPS)
- Таймауты на все операции
- Ограничение конкурентности (DoS защита)

### Рекомендации
- Rate limiting по IP
- Белый список доменов для URL
- Максимальный размер файла
- Аутентификация/авторизация

## Развертывание

### Docker
```bash
docker build -t ipxpress .
docker run -p 8080:8080 ipxpress
```

### Нативная сборка
```bash
go build -o ipxpress ./cmd/ipxpress
./ipxpress -addr :8080
```

## Зависимости

- **libvips:** Быстрая библиотека обработки изображений
- **govips:** Go биндинги для libvips
- Стандартная библиотека Go

## Лицензия

См. файл LICENSE в корне проекта.
