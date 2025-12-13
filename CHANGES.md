# Изменения в IPXpress - Минималистичная расширяемая библиотека

## Что сделано

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

1. Добавить больше встроенных процессоров (watermark, overlay)
2. Добавить middleware для метрик и мониторинга
3. Документировать все ProcessorFunc примеры
4. Добавить unit-тесты для расширяемости
