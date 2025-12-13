# Автоматическая инициализация libvips

## Изменения

Библиотека IPXpress теперь автоматически инициализирует libvips при первом использовании. Пользователям больше не нужно вручную вызывать `vips.Startup()` и `vips.Shutdown()`.

## Что было изменено

### 1. Автоматическая инициализация

Добавлена функция `initVips()` с использованием `sync.Once`, которая автоматически вызывается при:
- Создании нового обработчика через `NewHandler()`
- Создании нового процессора через `New()`

### 2. Настройка по умолчанию

При автоматической инициализации используются следующие настройки:
```go
vips.Config{
    ConcurrencyLevel: 0,    // Использовать все доступные ядра CPU
    MaxCacheMem:      2048, // 2GB кеш памяти
    MaxCacheSize:     5000, // До 5000 файлов в кеше
}
```

### 3. Кастомная настройка (опционально)

Для production окружений с высокими нагрузками добавлена функция `InitVipsWithConfig()`:

```go
ipxpress.InitVipsWithConfig(&vips.Config{
    ConcurrencyLevel: 0,
    MaxCacheMem:      4096,
    MaxCacheSize:     10000,
    MaxCacheFiles:    0,
}, vips.LogLevelWarning)
```

**Важно:** Эту функцию нужно вызвать **до** создания любых обработчиков или процессоров.

## Примеры использования

### Простое использование (рекомендуется)

```go
package main

import (
    "net/http"
    "github.com/vladislavsavi/ipxpress/pkg/ipxpress"
)

func main() {
    // vips инициализируется автоматически
    handler := ipxpress.NewHandler(nil)
    http.Handle("/img/", http.StripPrefix("/img/", handler))
    http.ListenAndServe(":8080", nil)
}
```

### С кастомными настройками vips

```go
package main

import (
    "net/http"
    "github.com/davidbyttow/govips/v2/vips"
    "github.com/vladislavsavi/ipxpress/pkg/ipxpress"
)

func main() {
    // Опционально: настройка vips перед использованием
    ipxpress.InitVipsWithConfig(&vips.Config{
        ConcurrencyLevel: 0,
        MaxCacheMem:      4096,
        MaxCacheSize:     10000,
    }, vips.LogLevelWarning)
    
    handler := ipxpress.NewHandler(nil)
    http.Handle("/img/", http.StripPrefix("/img/", handler))
    http.ListenAndServe(":8080", nil)
}
```

### Прямая обработка изображений

```go
func processImage(data []byte) ([]byte, error) {
    // vips инициализируется автоматически при первом вызове
    proc := ipxpress.New().
        FromBytes(data).
        Resize(800, 600)
    
    if err := proc.Err(); err != nil {
        return nil, err
    }
    
    result, err := proc.ToBytes(ipxpress.FormatJpeg, 85)
    proc.Close()
    
    return result, err
}
```

## Обратная совместимость

Изменения полностью обратно совместимы. Если вы явно вызываете `vips.Startup()` в своем коде, это будет работать как и раньше - автоматическая инициализация определит, что vips уже запущен, и не будет инициализировать его повторно.

## Обновленные файлы

1. `pkg/ipxpress/ipxpress.go` - добавлена автоматическая инициализация
2. `pkg/ipxpress/server.go` - NewHandler теперь вызывает initVips()
3. `cmd/ipxpress/main.go` - обновлен для использования InitVipsWithConfig()
4. `examples/library_usage/main.go` - удалена ручная инициализация vips
5. `pkg/ipxpress/ipxpress_test.go` - удалена ручная инициализация из тестов
6. Вся документация (README.md, LIBRARY_USAGE.md, и др.) - обновлена

## Преимущества

✅ Проще для новых пользователей - не нужно разбираться с libvips  
✅ Меньше boilerplate кода  
✅ Невозможно забыть инициализировать vips  
✅ Безопасная инициализация с использованием sync.Once  
✅ Возможность кастомной настройки для production  

## Требования

Библиотека libvips по-прежнему должна быть установлена в системе:

**Ubuntu/Debian:**
```bash
apt-get install libvips-dev
```

**macOS:**
```bash
brew install vips
```

См. [govips prerequisites](https://github.com/davidbyttow/govips#prerequisites) для других платформ.
