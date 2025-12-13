# IPXpress

IPXpress — высокопроизводительный сервис обработки изображений на Go с поддержкой множества форматов.

## Особенности

- ✅ Загрузка изображений по URL (HTTP/HTTPS)
- ✅ Изменение размера с сохранением пропорций (Lanczos фильтр)
- ✅ Поддержка форматов: **JPEG, PNG, GIF, WebP**
- ✅ Контроль качества сжатия (1-100)
- ✅ REST API сервис

## Поддерживаемые форматы

| Формат | Декодирование | Кодирование | Качество |
|--------|---|---|---|
| JPEG | ✅ | ✅ | ✅ |
| PNG | ✅ | ✅ | ❌ |
| GIF | ✅ | ✅ | ❌ |
| WebP | ✅ | ✅ | ✅ |
| AVIF | ✅ | ✅ | ✅ |

## Структура проекта

```
.
├── cmd/
│   └── ipxpress/          # HTTP сервер
├── pkg/ipxpress/          # Основная библиотека
│   ├── cache.go           # Система кеширования
│   ├── config.go          # Конфигурация
│   ├── fetcher.go         # Загрузка изображений
│   ├── format.go          # Форматы изображений
│   ├── ipxpress.go        # Image Processor
│   ├── params.go          # Параметры запроса
│   ├── server.go          # HTTP handler
│   └── *_test.go          # Тесты
├── ARCHITECTURE.md        # Архитектура проекта
├── API.md                 # API документация
├── CONTRIBUTING.md        # Руководство для разработчиков
└── README.md              # Этот файл
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

#### В формате AVIF (современный формат с лучшим сжатием)

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&w=1200&format=avif&quality=80" -o result.avif
```

#### Применение размытия

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&blur=5.0" -o blurred.jpg
```

#### Увеличение резкости

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&sharpen=1.5_1_2" -o sharp.jpg
```

#### Поворот на 90 градусов

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&rotate=90" -o rotated.jpg
```

#### Отражение (flip/flop)

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&flip=true" -o flipped.jpg
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&flop=true" -o flopped.jpg
```

#### Преобразование в ч/б

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&grayscale=true" -o grayscale.jpg
```

#### Вырезать область (crop)

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&extract=100_100_400_400" -o cropped.jpg
```

#### Комбинирование эффектов

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&w=800&grayscale=true&sharpen=1.0&quality=90&format=webp" -o processed.webp
```

#### С выбором алгоритма ресэмплинга

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&w=200&kernel=lanczos3" -o resized.jpg
```

#### Разрешить увеличение (upscale)

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/small.jpg&w=2000&enlarge=true" -o enlarged.jpg
```

## Параметры API

| Параметр | Описание | Тип | Обязательный |
|----------|---------|-----|---|
| `url` | URL изображения | string | ✅ |
| `w` | Максимальная ширина в пикселях | int | ❌ |
| `h` | Максимальная высота в пикселях | int | ❌ |
| `quality` | Качество сжатия (1-100) | int | ❌ |
| `format` | Формат вывода (jpeg, png, gif, webp, avif) | string | ❌ |

### Параметры изменения размера

| Параметр | Описание | Примеры |
|----------|---------|---------|
| `fit` | Режим масштабирования | contain, cover, fill, inside, outside |
| `position` | Позиционирование при кропе | center, top, bottom, left, right, entropy, attention |
| `kernel` | Алгоритм ресэмплинга | nearest, cubic, mitchell, lanczos2, lanczos3 |
| `enlarge` | Разрешить увеличение | true, false |

### Операции обработки

| Параметр | Описание | Формат значения |
|----------|---------|-----------------|
| `blur` | Размытие по Гауссу | sigma (float, например 5.0) |
| `sharpen` | Увеличение резкости | sigma_flat_jagged (например "1.5_1_2") |
| `rotate` | Поворот изображения | 0, 90, 180, 270 (градусы) |
| `flip` | Отразить вертикально | true |
| `flop` | Отразить горизонтально | true |
| `grayscale` | Преобразовать в ч/б | true |

### Кадрирование и расширение

| Параметр | Описание | Формат значения |
|----------|---------|-----------------|
| `extract` | Вырезать область | left_top_width_height (например "10_10_200_200") |
| `extend` | Добавить границы | top_right_bottom_left (например "10_10_10_10") |

### Цветовые операции

| Параметр | Описание | Формат значения |
|----------|---------|-----------------|
| `background` | Цвет фона | hex без # (например "ffffff" или "fff") |
| `negate` | Инвертировать цвета | true |
| `normalize` | Нормализация | true |
| `gamma` | Гамма коррекция | float (например 2.2) |
| `modulate` | Модуляция HSB | brightness_saturation_hue (например "1.2_0.8_90") |
| `flatten` | Удалить альфа канал | true |

**Поведение resize:**
- Если указана только ширина (`w`) — высота масштабируется пропорционально
- Если указана только высота (`h`) — ширина масштабируется пропорционально
- Если указаны обе — изображение масштабируется в наибольший размер, который поместится в прямоугольник

## Документация

- **[API.md](API.md)** - Полная документация API с примерами
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Архитектура и внутреннее устройство
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Руководство для разработчиков

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

- `github.com/davidbyttow/govips/v2` — Go binding для libvips (обработка изображений с нативной поддержкой JPEG, PNG, GIF, WebP, AVIF)

**Примечание:** Требуется установленная библиотека libvips. См. [инструкции по установке](https://github.com/davidbyttow/govips#prerequisites).

Библиотека автоматически инициализирует libvips при первом использовании, поэтому вам не нужно вручную вызывать `vips.Startup()` или `vips.Shutdown()`.

## Лицензия

MIT

