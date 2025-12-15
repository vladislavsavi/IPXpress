# IPXpress - Минималистичная расширяемая библиотека обработки изображений

IPXpress - это быстрая и гибкая библиотека для обработки изображений на Go, построенная на libvips.

## Ключевые возможности

✨ **Минималистичный API** - простое использование в любом проекте  
🔌 **Полная расширяемость** - используйте любые функции libvips  
🚀 **Высокая производительность** - использует libvips для быстрой обработки  
🔄 **Кеширование** - встроенное кеширование результатов  
🎯 **Гибкость** - используйте как библиотеку или готовый сервер  
⚙️ **Прямой доступ** - полный доступ к ImageRef для любых операций  

## Быстрый старт

### Как библиотека

```go
import "github.com/vladislavsavi/ipxpress/pkg/ipxpress"

func main() {
    // Самый простой способ: дефолтная конфигурация
    handler := ipxpress.NewHandler(nil)
    http.Handle("/img/", http.StripPrefix("/img/", handler))
    http.ListenAndServe(":8080", nil)
}
```

### Как самостоятельный сервер

```bash
go build -o ipxpress ./cmd/ipxpress
./ipxpress -addr :8080
```

## Использование

### Базовая интеграция

```go
// Создать обработчик с настройками по умолчанию
handler := ipxpress.NewHandler(nil)

// Или с кастомной конфигурацией
config := &ipxpress.Config{
    ProcessingLimit: 10,
    CacheTTL:        5 * time.Minute,
}
handler := ipxpress.NewHandler(config)

// Явный способ получить дефолтный конфиг
cfg := ipxpress.NewDefaultConfig()
handler2 := ipxpress.NewHandler(cfg)

// Добавить в ваш роутер
http.Handle("/images/", http.StripPrefix("/images/", handler))
```

### Расширение функциональности

#### Добавление кастомных процессоров

```go
handler := ipxpress.NewHandler(nil)

// Автоматически ориентировать изображения по EXIF
handler.UseProcessor(ipxpress.AutoOrientProcessor())

// Удалять метаданные для приватности
handler.UseProcessor(ipxpress.StripMetadataProcessor())

// Оптимизировать для веб-доставки
handler.UseProcessor(ipxpress.CompressionOptimizer())
```

#### Создание своего процессора

```go
customProcessor := func(proc *ipxpress.Processor, params *ipxpress.ProcessingParams) *ipxpress.Processor {
    // Ваша логика обработки
    return proc.Sharpen(1.5, 1.0, 2.0)
}

handler.UseProcessor(customProcessor)
```

#### Добавление middleware

```go
// CORS
handler.UseMiddleware(ipxpress.CORSMiddleware([]string{"*"}))

// Аутентификация
handler.UseMiddleware(ipxpress.AuthMiddleware([]string{"secret-token"}))

// Логирование
logger := func(format string, args ...interface{}) {
    log.Printf(format, args...)
}
handler.UseMiddleware(ipxpress.LoggingMiddleware(logger))

// Свой middleware
customMiddleware := func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Ваша логика
        next.ServeHTTP(w, r)
    })
}
handler.UseMiddleware(customMiddleware)
```

### Несколько обработчиков

```go
// Публичный обработчик с ограничениями
publicHandler := ipxpress.NewHandler(&ipxpress.Config{
    ProcessingLimit: 5,
})
publicHandler.UseMiddleware(ipxpress.RateLimitMiddleware(100))

// Приватный обработчик с авторизацией
privateHandler := ipxpress.NewHandler(&ipxpress.Config{
    ProcessingLimit: 20,
})
privateHandler.UseMiddleware(ipxpress.AuthMiddleware([]string{"admin-token"}))

// Монтируем оба
http.Handle("/public/img/", http.StripPrefix("/public/img/", publicHandler))
http.Handle("/private/img/", http.StripPrefix("/private/img/", privateHandler))
```

## API

```
GET /ipx/?url=https://example.com/image.jpg&w=800&h=600&quality=85&format=webp
```

**Параметры:**
- `url` (обязательный) - URL изображения для обработки
- `w` - максимальная ширина
- `h` - максимальная высота  
- `quality` - качество (1-100, по умолчанию 85)
- `format` - формат вывода (jpeg, png, gif, webp)

Полная документация API: [API.md](API.md)

## Прямая обработка изображений

Можно использовать без HTTP:

```go
proc := ipxpress.New().
    FromBytes(imageData).
    Resize(800, 600).
    Sharpen(1.0, 1.0, 2.0)

output, err := proc.ToBytes(ipxpress.FormatJPEG, 85)
proc.Close()
```

### Использование любых функций libvips

IPXpress предоставляет полный доступ к любым функциям libvips через несколько механизмов:

#### 1. Прямой доступ к ImageRef

```go
proc := ipxpress.New().FromBytes(imageData)

// Получить прямой доступ к vips.ImageRef
img := proc.ImageRef()
if img != nil {
    img.Blur(2.0)
    img.Sharpen(1.5, 0.5, 1.0)
    img.Modulate(1.1, 1.2, 0)
}

output, _ := proc.ToBytes(ipxpress.FormatJPEG, 85)
```

#### 2. ApplyFunc для пользовательских операций

```go
proc := ipxpress.New().
    FromBytes(imageData).
    ApplyFunc(func(img *vips.ImageRef) error {
        if err := img.Blur(2.0); err != nil {
            return err
        }
        return img.Sharpen(1.5, 0.5, 1.0)
    })

output, _ := proc.ToBytes(ipxpress.FormatJPEG, 85)
```

#### 3. VipsOperationBuilder для цепочки операций

```go
proc := ipxpress.New().FromBytes(imageData)

builder := ipxpress.NewVipsOperationBuilder(proc)
err := builder.
    Blur(2.0).
    Sharpen(1.5, 0.5, 1.0).
    Modulate(1.1, 1.2, 0).
    Error()

output, _ := proc.ToBytes(ipxpress.FormatJPEG, 85)
```

Подробно: [CUSTOM_OPERATIONS.md](CUSTOM_OPERATIONS.md)

## Документация

- [LIBRARY_USAGE.md](LIBRARY_USAGE.md) - Детальная документация по использованию библиотеки
- [API.md](API.md) - Полное описание API
- [ARCHITECTURE.md](ARCHITECTURE.md) - Архитектура проекта
- [CUSTOM_OPERATIONS.md](CUSTOM_OPERATIONS.md) - Расширение возможностей и использование любых функций libvips

## Примеры использования

### Интеграция с Chi Router

```go
r := chi.NewRouter()
r.Get("/", homeHandler)

imgHandler := ipxpress.NewHandler(nil)
r.Mount("/img", imgHandler)
```

### Интеграция с Gorilla Mux

```go
r := mux.NewRouter()
r.HandleFunc("/", homeHandler)

imgHandler := ipxpress.NewHandler(nil)
r.PathPrefix("/img/").Handler(http.StripPrefix("/img/", imgHandler))
```

### Production setup

```go
// Библиотека автоматически инициализирует vips
// Вам не нужно вызывать vips.Startup() или vips.Shutdown()

// Базовый вариант без настроек
handler := ipxpress.NewHandler(nil)

// Кастомные настройки при необходимости
config := ipxpress.NewDefaultConfig()
config.ProcessingLimit = 10
config.CacheTTL = 30 * time.Minute

handler = ipxpress.NewHandler(config)
handler.UseProcessor(ipxpress.AutoOrientProcessor())
handler.UseProcessor(ipxpress.CompressionOptimizer())
handler.UseMiddleware(ipxpress.CORSMiddleware([]string{"*"}))
```

## Требования

- Go 1.21+
- libvips 8.12+

**Примечание:** Библиотека автоматически инициализирует libvips при первом использовании. Вам не нужно вручную вызывать `vips.Startup()` или `vips.Shutdown()`.

## Установка libvips

### Ubuntu/Debian
```bash
apt-get install libvips-dev
```

### macOS
```bash
brew install vips
```

## Лицензия

MIT License - см. [LICENSE](LICENSE)

