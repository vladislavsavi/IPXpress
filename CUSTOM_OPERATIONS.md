# Расширение IPXpress: Использование любых функций libvips

## Обзор

IPXpress предоставляет несколько способов использовать любые функции из libvips, без ограничений встроенными методами библиотеки. Это позволяет вам применять произвольные трансформации изображений.

## Методы расширения

### 1. Direct ImageRef Access - `ImageRef()`

Получите прямой доступ к базовому объекту `vips.ImageRef` для использования любых функций libvips.

```go
processor := ipxpress.New()
processor.FromBytes(imageData)

// Получить доступ к ImageRef и использовать любые функции libvips
imgRef := processor.ImageRef()
if imgRef != nil {
    // Используйте любые методы libvips напрямую
    imgRef.Blur(2.5)
    imgRef.Sharpen(1.5, 0.5, 1.0)
    imgRef.Modulate(1.1, 1.2, 0)
}

result, err := processor.ToBytes(ipxpress.FormatJPEG, 85)
```

### 2. ApplyFunc - Функции обратного вызова

Используйте метод `ApplyFunc` для применения пользовательских функций обработки с автоматическим управлением ошибками.

```go
processor := ipxpress.New()
processor.FromBytes(imageData)

// Применить пользовательскую функцию
processor.ApplyFunc(func(img *vips.ImageRef) error {
    // Можно использовать любые функции libvips
    if err := img.Blur(2.0); err != nil {
        return err
    }
    return img.Sharpen(1.5, 0.5, 1.0)
})

if processor.Err() != nil {
    log.Fatal(processor.Err())
}

result, err := processor.ToBytes(ipxpress.FormatWebP, 80)
```

### 3. VipsOperationBuilder - Fluent API

Построить цепочку операций с удобным интерфейсом и обработкой ошибок:

```go
processor := ipxpress.New()
processor.FromBytes(imageData)

builder := ipxpress.NewVipsOperationBuilder(processor)
err := builder.
    Blur(2.0).
    Sharpen(1.5, 0.5, 1.0).
    Modulate(1.1, 1.2, 0).
    Median(3).
    Error()

if err != nil {
    log.Fatal(err)
}

result, err := processor.ToBytes(ipxpress.FormatJPEG, 85)
```

### 4. CustomOperation - Пользовательские операции как процессоры

Создавайте переиспользуемые пользовательские операции:

```go
// Определить пользовательскую операцию
applySepiaEffect := func(p *ipxpress.Processor, params *ipxpress.ProcessingParams) error {
    img := p.ImageRef()
    if img == nil {
        return errors.New("no image loaded")
    }
    
    // Применить эффект сепия
    if err := img.Modulate(1.0, 0.0, 0); err != nil {
        return err
    }
    sepiaColor := &vips.Color{R: 255, G: 200, B: 124}
    return img.Tint(sepiaColor)
}

// Использовать операцию
processor := ipxpress.New()
processor.FromBytes(imageData)
processor.ApplyCustom(applySepiaEffect, nil)
```

### 5. Процессоры в обработчике

Добавляйте пользовательские операции в pipeline обработки запросов:

```go
config := &ipxpress.Config{ProcessingLimit: 10}
handler := ipxpress.NewHandler(config)

// Добавить пользовательский обработчик с любой операцией libvips
handler.UseProcessor(func(p *ipxpress.Processor, params *ipxpress.ProcessingParams) *ipxpress.Processor {
    return p.ApplyFunc(func(img *vips.ImageRef) error {
        // Применить любую операцию libvips
        return img.Blur(1.5)
    })
})

// Или используя встроенные операции
handler.UseProcessor(func(p *ipxpress.Processor, params *ipxpress.ProcessingParams) *ipxpress.Processor {
    return p.ApplyFunc(func(img *vips.ImageRef) error {
        // Применить сложную операцию
        if err := img.Sharpen(2.0, 0.5, 1.0); err != nil {
            return err
        }
        return img.Modulate(1.05, 1.1, 0)
    })
})

mux := http.NewServeMux()
mux.Handle("/img/", http.StripPrefix("/img/", handler))
http.ListenAndServe(":8080", mux)
```

## Примеры использования

### Пример 1: Применение фильтра размытия с резкостью

```go
processor := ipxpress.New()
processor.FromBytes(imageData).
    ApplyFunc(func(img *vips.ImageRef) error {
        if err := img.Blur(0.5); err != nil {
            return err
        }
        return img.Sharpen(2.0, 0.5, 1.0)
    })

result, _ := processor.ToBytes(ipxpress.FormatJPEG, 90)
```

### Пример 2: Создание эффекта сепия

```go
processor := ipxpress.New()
processor.FromBytes(imageData).
    ApplyFunc(func(img *vips.ImageRef) error {
        // Преобразовать в оттенки серого
        if err := img.Modulate(1.0, 0.0, 0); err != nil {
            return err
        }
        // Применить сепия оттенок
        sepiaColor := &vips.Color{R: 255, G: 200, B: 124}
        return img.Tint(sepiaColor)
    })

result, _ := processor.ToBytes(ipxpress.FormatJPEG, 85)
```

### Пример 3: Регулировка контраста и яркости

```go
processor := ipxpress.New()
processor.FromBytes(imageData).
    ApplyFunc(func(img *vips.ImageRef) error {
        // Увеличить яркость на 10% и насыщенность на 20%
        return img.Modulate(1.1, 1.2, 0)
    }).
    ApplyFunc(func(img *vips.ImageRef) error {
        // Увеличить контраст
        return img.Linear([]float64{1.3}, []float64{0})
    })

result, _ := processor.ToBytes(ipxpress.FormatWebP, 80)
```

### Пример 4: Создание миниатюры со специальной обработкой

```go
processor := ipxpress.New()
processor.FromBytes(imageData).
    Resize(200, 200).
    ApplyFunc(func(img *vips.ImageRef) error {
        // Применить фильтр размытия для миниатюры
        if err := img.Blur(1.0); err != nil {
            return err
        }
        // Применить резкость
        return img.Sharpen(1.5, 0.5, 1.0)
    })

result, _ := processor.ToBytes(ipxpress.FormatJPEG, 75)
```

### Пример 5: Использование встроенных предопределенных операций

```go
processor := ipxpress.New()
processor.FromBytes(imageData)

// Используйте встроенные factory функции для создания операций
processor.ApplyCustom(ipxpress.GaussianBlurOperation(2.5), nil).
    ApplyCustom(ipxpress.SaturationOperation(1.2), nil)

result, _ := processor.ToBytes(ipxpress.FormatJPEG, 85)
```

## Доступные встроенные операции

IPXpress предоставляет factory функции для часто используемых операций:

- `GaussianBlurOperation(sigma)` - Размытие Гаусса
- `EdgeDetectionOperation(kernel)` - Обнаружение краев
- `SepiaOperation()` - Эффект сепия
- `BrightnessOperation(brightness)` - Регулировка яркости
- `SaturationOperation(saturation)` - Регулировка насыщенности
- `ContrastOperation(contrast)` - Регулировка контраста

## VipsOperationBuilder - Встроенные методы

Builder предоставляет удобные методы для цепочки операций:

```go
builder := ipxpress.NewVipsOperationBuilder(processor)
err := builder.
    Blur(2.0).                    // Гауссово размытие (GaussianBlur)
    Sharpen(1.5, 0.5, 1.0).      // Резкость
    Modulate(1.1, 1.2, 0).        // Яркость, насыщенность, оттенок
    Median(3).                    // Размытие (медианное/гауссово)
    Error()

if err != nil {
    log.Fatal(err)
}
```

## Полный доступ к libvips

Если вам нужна операция, не включенная в встроенные методы, используйте `ImageRef()` напрямую:

```go
processor := ipxpress.New()
processor.FromBytes(imageData)

img := processor.ImageRef()
if img != nil {
    // Полный доступ ко всем методам vips.ImageRef
    img.Blur(...)
    img.Sharpen(...)
    img.Convolve(...)
    img.Composite(...)
    // и т.д.
}
```

## Обработка ошибок

Все методы поддерживают цепочку и обработку ошибок:

```go
processor := ipxpress.New()
processor.FromBytes(imageData).
    Resize(800, 600).
    ApplyFunc(func(img *vips.ImageRef) error {
        return img.Blur(2.0)
    })

if processor.Err() != nil {
    log.Printf("Ошибка обработки: %v", processor.Err())
    return
}

result, err := processor.ToBytes(ipxpress.FormatJPEG, 85)
if err != nil {
    log.Printf("Ошибка кодирования: %v", err)
}
```

## Производительность

- Все операции работают в памяти благодаря libvips
- Используется кэширование на уровне Handler
- Поддерживаются одновременные запросы (настраивается через `ProcessingLimit`)
- Автоматическое управление памятью при закрытии Processor

## Закрытие ресурсов

```go
processor := ipxpress.New()
processor.FromBytes(imageData).
    Resize(800, 600).
    ApplyFunc(func(img *vips.ImageRef) error {
        return img.Blur(1.5)
    })

result, _ := processor.ToBytes(ipxpress.FormatJPEG, 85)
processor.Close() // Важно: освободить ресурсы
```

