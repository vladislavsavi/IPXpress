# API Документация IPXpress

## Обзор

IPXpress предоставляет REST API для обработки изображений в реальном времени. Сервис загружает изображения по URL, применяет трансформации и возвращает результат.

## Базовый URL

```
http://localhost:8080/ipx/
```

## Endpoint

### GET /ipx/

Обрабатывает изображение с заданными параметрами.

#### Параметры запроса

| Параметр | Тип | Обязательный | По умолчанию | Описание |
|----------|-----|--------------|--------------|----------|
| `url` | string | **Да** | - | URL изображения для обработки (HTTP/HTTPS) |
| `w` | integer | Нет | - | Максимальная ширина в пикселях |
| `h` | integer | Нет | - | Максимальная высота в пикселях |
| `quality` | integer | Нет | 85 | Качество сжатия для JPEG/WebP (1-100) |
| `format` | string | Нет | original | Формат вывода: `jpeg`, `png`, `gif`, `webp` |

#### Заголовки ответа

- `Content-Type`: MIME тип изображения (`image/jpeg`, `image/png`, и т.д.)
- `Content-Length`: Размер изображения в байтах
- `Cache-Control`: Директивы кеширования

#### Коды ответа

| Код | Описание |
|-----|----------|
| 200 | Успешная обработка изображения |
| 400 | Неверные параметры запроса |
| 500 | Внутренняя ошибка сервера |

## Примеры использования

### 1. Базовое изменение размера

Изменить размер изображения с сохранением пропорций:

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/photo.jpg&w=800" \
  -o resized.jpg
```

**Поведение:**
- Ширина будет 800px
- Высота вычисляется автоматически для сохранения пропорций
- Формат остается JPEG

### 2. Изменение размера с обеими размерностями

Вписать изображение в прямоугольник 1000x600:

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/photo.jpg&w=1000&h=600" \
  -o fitted.jpg
```

**Поведение:**
- Изображение масштабируется так, чтобы поместиться в 1000x600
- Пропорции сохраняются
- Итоговый размер может быть меньше указанного

### 3. Конвертация формата

Конвертировать JPEG в WebP:

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/photo.jpg&format=webp" \
  -o photo.webp
```

### 4. Конвертация в PNG без сжатия

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/photo.jpg&format=png" \
  -o photo.png
```

**Примечание:** PNG не поддерживает параметр `quality`, он игнорируется.

### 5. Изменение размера с контролем качества

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/photo.jpg&w=1200&quality=95" \
  -o high-quality.jpg
```

**Рекомендации по качеству:**
- `70-80`: Хорошее качество, меньший размер файла
- `85` (default): Оптимальный баланс
- `90-100`: Высокое качество, больший размер файла

### 6. Оптимизация для веба (WebP + качество)

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/large.jpg&w=1200&format=webp&quality=80" \
  -o optimized.webp
```

### 7. Создание превью

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/image.jpg&w=200&h=200&quality=75" \
  -o thumbnail.jpg
```

### 8. Получение оригинального изображения

Если не указаны параметры трансформации, возвращается оригинал:

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/photo.jpg" \
  -o original.jpg
```

## Поведение resize

### Только ширина (w)

```bash
?url=https://example.com/1000x500.jpg&w=500
# Результат: 500x250
```

Высота масштабируется пропорционально.

### Только высота (h)

```bash
?url=https://example.com/1000x500.jpg&h=100
# Результат: 200x100
```

Ширина масштабируется пропорционально.

### Ширина и высота (w + h)

```bash
?url=https://example.com/1000x500.jpg&w=600&h=400
# Результат: 600x300 (вписывается в 600x400)
```

Изображение масштабируется так, чтобы поместиться в заданный прямоугольник, сохраняя пропорции.

## Поддерживаемые форматы

### Входные форматы

- JPEG / JPG
- PNG (включая прозрачность)
- GIF (статичные)
- WebP

### Выходные форматы

| Формат | Значение | Качество | Прозрачность | Примечания |
|--------|----------|----------|--------------|------------|
| JPEG | `jpeg` или `jpg` | ✅ | ❌ | Лучшее сжатие для фото |
| PNG | `png` | ❌ | ✅ | Без потерь, для графики |
| GIF | `gif` | ❌ | ✅ | Ограниченная палитра |
| WebP | `webp` | ✅ | ✅ | Современный формат |

## Производительность и кеширование

### Кеширование

Сервис кеширует обработанные изображения на **30 секунд**. Повторные запросы с теми же параметрами обрабатываются мгновенно.

**Ключ кеша:**
```
MD5(url + width + height + quality + format)
```

### Заголовки Cache-Control

```
Cache-Control: public, max-age=604800  # Обработанные изображения (7 дней)
Cache-Control: public, max-age=31536000 # Оригинальные изображения (1 год)
```

### Рекомендации

1. **CDN:** Разместите IPXpress за CDN для лучшей производительности
2. **URL стабильность:** Используйте стабильные URL изображений
3. **Пакетная обработка:** Отправляйте запросы параллельно

## Ограничения

### Текущие ограничения

- Максимум 256 одновременных обработок
- Timeout на загрузку: 20 секунд
- Timeout на подключение: 5 секунд
- Только HTTP/HTTPS URL

### Рекомендуемые практики

1. **Размер изображений:**
   - Входные: до 20-30 MP
   - Выходные: разумные размеры (до 4000px по большей стороне)

2. **Rate limiting:**
   - Рекомендуется ограничить количество запросов на клиента
   - Используйте nginx/haproxy для rate limiting

3. **Мониторинг:**
   - Отслеживайте latency и error rate
   - Настройте алерты на 5xx ошибки

## Интеграция

### JavaScript / Fetch API

```javascript
const imageUrl = encodeURIComponent('https://example.com/photo.jpg');
const apiUrl = `http://localhost:8080/ipx/?url=${imageUrl}&w=800&format=webp`;

fetch(apiUrl)
  .then(response => response.blob())
  .then(blob => {
    const img = document.createElement('img');
    img.src = URL.createObjectURL(blob);
    document.body.appendChild(img);
  });
```

### Python / requests

```python
import requests

params = {
    'url': 'https://example.com/photo.jpg',
    'w': 800,
    'format': 'webp',
    'quality': 85
}

response = requests.get('http://localhost:8080/ipx/', params=params)

with open('output.webp', 'wb') as f:
    f.write(response.content)
```

### Go

```go
package main

import (
    "io"
    "net/http"
    "net/url"
    "os"
)

func main() {
    params := url.Values{}
    params.Add("url", "https://example.com/photo.jpg")
    params.Add("w", "800")
    params.Add("format", "webp")
    
    apiURL := "http://localhost:8080/ipx/?" + params.Encode()
    
    resp, err := http.Get(apiURL)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    
    out, err := os.Create("output.webp")
    if err != nil {
        panic(err)
    }
    defer out.Close()
    
    io.Copy(out, resp.Body)
}
```

### HTML (прямое использование)

```html
<img src="http://localhost:8080/ipx/?url=https://example.com/photo.jpg&w=400&format=webp" 
     alt="Processed image">
```

## Обработка ошибок

### Примеры ошибок

#### Отсутствует URL

```bash
curl "http://localhost:8080/ipx/"
# HTTP 400: missing image URL
```

#### Неверный URL

```bash
curl "http://localhost:8080/ipx/?url=not-a-valid-url"
# HTTP 400: invalid image URL: ...
```

#### Недоступное изображение

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/404.jpg"
# HTTP 400: image fetch failed with status 404
```

#### Ошибка обработки

```bash
curl "http://localhost:8080/ipx/?url=https://example.com/corrupted.jpg&w=800"
# HTTP 500: processing: failed to decode image
```

### Обработка в коде

```javascript
fetch(apiUrl)
  .then(response => {
    if (!response.ok) {
      return response.text().then(text => {
        throw new Error(`Server error: ${text}`);
      });
    }
    return response.blob();
  })
  .catch(error => {
    console.error('Image processing failed:', error);
  });
```

## Health Check

### Endpoint

```
GET /health
```

### Пример

```bash
curl http://localhost:8080/health
# OK
```

Используйте для мониторинга доступности сервиса.

## Дополнительные ресурсы

- [README.md](README.md) - Общая информация о проекте
- [ARCHITECTURE.md](ARCHITECTURE.md) - Архитектура и внутреннее устройство
- [Примеры использования](examples/) - Больше примеров интеграции
