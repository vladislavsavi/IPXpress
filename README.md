# IPXpress

IPXpress — высокопроизводительный сервис обработки изображений на Go с поддержкой множества форматов.

## Особенности

- ✅ Загрузка изображений по URL (HTTP/HTTPS)
- ✅ Изменение размера с сохранением пропорций (Lanczos фильтр)
- ✅ Поддержка форматов: **JPEG, PNG, GIF, WebP**
- ✅ Контроль качества сжатия (1-100)
- ✅ REST API сервис
- ✅ Chainable API для библиотечного использования

## Поддерживаемые форматы

| Формат | Декодирование | Кодирование | Качество |
|--------|---|---|---|
| JPEG | ✅ | ✅ | ✅ |
| PNG | ✅ | ✅ | ❌ |
| GIF | ✅ | ✅ | ❌ |
| WebP | ✅ | ✅ | ✅ |

## Структура проекта

```
.
├── cmd/
│   ├── ipxpress/          # CLI утилита (будущая реализация)
│   └── ipxpress-server/   # REST API сервер
├── pkg/ipxpress/          # Основная библиотека
│   ├── ipxpress.go        # Image Processor
│   ├── server.go          # HTTP handler
│   └── server_test.go     # Тесты
└── README.md
```

## Быстрый старт

### Сборка сервера

```bash
go build ./cmd/ipxpress-server
```

### Запуск сервера

```bash
./ipxpress-server -addr :8080
```

Сервер будет доступен по адресу `http://localhost:8080/ipx/`

### Примеры запросов

#### Базовый запрос с изменением размера

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&w=800&h=600"
```

#### С контролем качества

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&w=1000&h=500&quality=85&format=jpeg"
```

#### В формате WebP

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&w=1000&h=500&quality=100&format=webp" -o result.webp
```

#### В формате PNG

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&format=png" -o result.png
```

## Параметры API

| Параметр | Описание | Тип | Обязательный |
|----------|---------|-----|---|
| `url` | URL изображения | string | ✅ |
| `w` | Максимальная ширина в пикселях | int | ❌ |
| `h` | Максимальная высота в пикселях | int | ❌ |
| `quality` | Качество сжатия (1-100) | int | ❌ |
| `format` | Формат вывода (jpeg, png, gif, webp) | string | ❌ |

**Поведение resize:**
- Если указана только ширина (`w`) — высота масштабируется пропорционально
- Если указана только высота (`h`) — ширина масштабируется пропорционально
- Если указаны обе — изображение масштабируется в наибольший размер, который поместится в прямоугольник

## Использование как библиотека

```go
package main

import (
	"github.com/deadpixel/ipxpress/pkg/ipxpress"
)

func main() {
	// Загрузить изображение из байтов
	proc := ipxpress.New().
		FromBytes(imageBytes).
		Resize(800, 600)
	
	if err := proc.Err(); err != nil {
		panic(err)
	}
	
	// Закодировать в WebP с качеством 85
	output, err := proc.ToBytes("webp", 85)
	if err != nil {
		panic(err)
	}
	
	// Использовать output...
}
```

## Тесты

```bash
go test ./pkg/ipxpress
```

## Зависимости

- `github.com/chai2010/webp` — WebP кодирование/декодирование
- `github.com/disintegration/imaging` — Высококачественное масштабирование (Lanczos)
- `golang.org/x/image` — Поддержка JPEG, PNG, GIF

## Лицензия

MIT
