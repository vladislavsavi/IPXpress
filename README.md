# IPXpress

IPXpress — высокопроизводительное ядро обработки изображений на Go (MVP).

Особенности (MVP):

- Загрузка изображения из байтов/файла
- Изменение размера с сохранением пропорций
- Кодирование в JPEG/PNG и получение результата в виде байтов

Структура проекта:

- `pkg/ipxpress` — основная библиотека и chainable API
- `cmd/ipxpress` — маленький CLI-пример

Быстрый старт:

Сборка CLI:
```bash
go build ./cmd/ipxpress
```

Пример использования (CLI):
```bash
./ipxpress -in input.jpg -out out.jpg -w 800 -h 0 -format jpeg -quality 85
```

Пример использования как библиотека:

```go
proc := ipxpress.New().FromBytes(imgBytes).Resize(800, 0)
if err := proc.Err(); err != nil { /* handle */ }
out, err := proc.ToBytes("jpeg", 85)
```

Тесты:
```bash
go test ./pkg/ipxpress
```
