# Изменения в IPXpress

## v0.2.0 (2025-12-15)

**Расширение функциональности для использования любых функций libvips:**

- **Прямой доступ к ImageRef**: метод `ImageRef()` для получения прямого доступа к `vips.ImageRef` и использования любых функций libvips.
- **ApplyFunc метод**: применение пользовательских функций обработки с автоматической обработкой ошибок и поддержкой цепочки операций.
- **VipsOperationBuilder (fluent API)**: удобный интерфейс для цепочки операций с методами `Blur()`, `Sharpen()`, `Modulate()`, `Median()`, `Flatten()`, `Invert()`.
- **CustomOperation тип**: для создания переиспользуемых пользовательских операций.
- **Встроенные операции**: `GaussianBlurOperation()`, `EdgeDetectionOperation()`, `SepiaOperation()`, `BrightnessOperation()`, `SaturationOperation()`, `ContrastOperation()`.
- **Документация**: новый файл `CUSTOM_OPERATIONS.md` с полными примерами.
- **Юнит-тесты**: полное покрытие в `extensions_test.go`.

Примеры использования:
```go
// Прямой доступ: img := proc.ImageRef()
// ApplyFunc: proc.ApplyFunc(func(img *vips.ImageRef) error { ... })
// Builder: builder.Blur(2.0).Sharpen(1.5, 0.5, 1.0)
```

## v0.1.0 (2025-12-14)

Основные улучшения и фиксы:

- **Дефолтная конфигурация**: `NewDefaultConfig()` и автоприменение при `NewHandler(nil)`. Библиотечным клиентам не нужно настраивать вручную.
- **Исправление внутреннего кеша**: проставление `Timestamp` в `InMemoryCache.Set` — стабильный TTL, повторные запросы бьют в кеш.
- **HTTP кеширование**: конфигурируемые заголовки `Cache-Control` (`ClientMaxAge`, `SMaxAge`) и поддержка `ETag`/`If-None-Match` (включено по умолчанию).
- **Короткие алиасы параметров**: совместимость с ipx v2 (`w,h,f,q,s,b,pos`), плюс парсинг `s=WxH`.
- **Заголовки контента**: корректный `Content-Type` и принудительный `Content-Disposition: inline` для отображения, без скачивания.
- **Тюнинг энкодеров**: WebP `ReductionEffort=4` (быстро), AVIF `Speed=6` — баланс скорости и сжатия; отсутствие непредвиденных даунгрейдов формата.

Стабильность:

- Все тесты проходят: `go test ./...` OK.
- Сравнение с внешним ipx показало одинаковые размеры выходных файлов, локально быстрее.

Конфигурация (новые поля):

- `ClientMaxAge` — `Cache-Control: max-age` (секунды), по умолчанию `604800`.
- `SMaxAge` — `Cache-Control: s-maxage` для CDN, по умолчанию `0` (выключено).
- `EnableETag` — включение `ETag` и `304 Not Modified`, по умолчанию `true`.

Пример:

```go
cfg := ipxpress.NewDefaultConfig()
cfg.ClientMaxAge = 3600 // 1 час
cfg.SMaxAge = 3600      // 1 час для CDN
cfg.CacheTTL = 10 * time.Minute
cfg.EnableETag = true
handler := ipxpress.NewHandler(cfg)
```

Документация:

- Обновлены `README.md`, `README.library.md`, `API.md` — разделы про кеширование и дефолтные настройки.

### 1. Расширяемая архитектура обработчика


**Добавлены новые типы:**
- `ProcessorFunc` - функция для кастомной обработки изображений
- `MiddlewareFunc` - функция для добавления middleware

**Новые методы Handler:**
```go
handler.UseProcessor(processorFunc)  // Добавить кастомный процессор
handler.UseMiddleware(middleware)     // Добавить middleware
```

### 2. Встроенные процессоры (`pkg/ipxpress/examples.go`)

- `AutoOrientProcessor()` - автоматическая ориентация по EXIF
- `StripMetadataProcessor()` - удаление метаданных для приватности
- `CompressionOptimizer()` - оптимизация сжатия для веб

### 3. Встроенные middleware

- `CORSMiddleware(origins)` - CORS заголовки
- `LoggingMiddleware(logger)` - логирование запросов
- `RateLimitMiddleware(maxRequests)` - ограничение частоты запросов
- `AuthMiddleware(tokens)` - аутентификация по токенам

### 4. Документация

**Новые файлы:**
- `LIBRARY_USAGE.md` - полная документация по использованию библиотеки
- `README.library.md` - краткий README для библиотеки
- `examples/library_usage/main.go` - рабочий пример

## Примеры использования

### Простая интеграция

```go
handler := ipxpress.NewHandler(nil)
http.Handle("/img/", http.StripPrefix("/img/", handler))
```

### С кастомными процессорами

```go
handler := ipxpress.NewHandler(nil)
handler.UseProcessor(ipxpress.AutoOrientProcessor())
handler.UseProcessor(ipxpress.StripMetadataProcessor())

// Свой процессор
customProc := func(proc *ipxpress.Processor, params *ipxpress.ProcessingParams) *ipxpress.Processor {
    return proc.Sharpen(1.5, 1.0, 2.0)
}
handler.UseProcessor(customProc)
```

### С middleware

```go
handler := ipxpress.NewHandler(nil)
handler.UseMiddleware(ipxpress.CORSMiddleware([]string{"*"}))
handler.UseMiddleware(ipxpress.AuthMiddleware([]string{"secret-token"}))
```

### Несколько обработчиков

```go
// Публичный с ограничениями
publicHandler := ipxpress.NewHandler(config1)
publicHandler.UseMiddleware(ipxpress.RateLimitMiddleware(100))

// Приватный с авторизацией
privateHandler := ipxpress.NewHandler(config2)
privateHandler.UseMiddleware(ipxpress.AuthMiddleware(tokens))

http.Handle("/public/img/", http.StripPrefix("/public/img/", publicHandler))
http.Handle("/private/img/", http.StripPrefix("/private/img/", privateHandler))
```

## Архитектурные улучшения

### До:
- Жесткая структура без возможности расширения
- Только встроенные трансформации
- Нет поддержки middleware
- Один обработчик на весь сервер

### После:
- Гибкая архитектура с ProcessorFunc и MiddlewareFunc
- Можно добавлять кастомные процессоры в pipeline
- Поддержка middleware для HTTP уровня
- Можно создавать несколько обработчиков с разными настройками
- Легко интегрируется в существующие проекты

## Обратная совместимость

✅ Все существующие функции работают без изменений  
✅ Старый код продолжит работать  
✅ Новые функции полностью опциональны

## Использование в других проектах

```bash
go get github.com/vladislavsavi/ipxpress/pkg/ipxpress
```

```go
import "github.com/vladislavsavi/ipxpress/pkg/ipxpress"

handler := ipxpress.NewHandler(nil)
// Добавляем в свой роутер
```

## Тестирование

```bash
# Компиляция
go build ./...

# Запуск примера
go run examples/library_usage/main.go

# Тест API
curl "http://localhost:8080/img/?url=https://example.com/image.jpg&w=800"
```

## Следующие шаги

1. Добавить `s-maxage` и `ETag` настройки в CLI флаги.
2. Встроить метрики (профилирование, p50/p95 времени обработки).
3. Конфигурируемый стор кэша (например, Redis) для горизонтального масштабирования.
4. Документировать рекомендации по качеству и производительности.
