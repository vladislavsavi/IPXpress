// Package main демонстрирует использование IPXpress как библиотеки
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/deadpixel/ipxpress/pkg/ipxpress"
)

func main() {
	// Инициализация libvips (обязательно)
	vips.Startup(nil)
	defer vips.Shutdown()

	// Пример 1: Простое изменение размера
	example1()

	// Пример 2: Изменение размера с конвертацией формата
	example2()

	// Пример 3: Batch обработка
	example3()
}

// example1 демонстрирует базовое использование
func example1() {
	fmt.Println("=== Пример 1: Изменение размера ===")

	// Прочитать изображение из файла
	data, err := os.ReadFile("input.jpg")
	if err != nil {
		log.Printf("Ошибка чтения файла: %v", err)
		return
	}

	// Создать процессор и применить трансформации
	proc := ipxpress.New().
		FromBytes(data).
		Resize(800, 600)

	// Проверить ошибки
	if err := proc.Err(); err != nil {
		log.Printf("Ошибка обработки: %v", err)
		return
	}

	// Закодировать в JPEG с качеством 85
	output, err := proc.ToBytes(ipxpress.FormatJPEG, 85)
	proc.Close() // Освободить память
	if err != nil {
		log.Printf("Ошибка кодирования: %v", err)
		return
	}

	// Сохранить результат
	if err := os.WriteFile("output_resized.jpg", output, 0644); err != nil {
		log.Printf("Ошибка записи: %v", err)
		return
	}

	fmt.Printf("Обработано: %d байт -> %d байт\n", len(data), len(output))
}

// example2 демонстрирует конвертацию формата
func example2() {
	fmt.Println("\n=== Пример 2: Конвертация в WebP ===")

	data, err := os.ReadFile("input.jpg")
	if err != nil {
		log.Printf("Ошибка чтения файла: %v", err)
		return
	}

	// Изменить размер и конвертировать в WebP
	proc := ipxpress.New().
		FromBytes(data).
		Resize(1200, 0) // Только ширина, высота автоматически

	if err := proc.Err(); err != nil {
		log.Printf("Ошибка обработки: %v", err)
		return
	}

	// Закодировать в WebP с высоким качеством
	output, err := proc.ToBytes(ipxpress.FormatWebP, 90)
	proc.Close()
	if err != nil {
		log.Printf("Ошибка кодирования: %v", err)
		return
	}

	if err := os.WriteFile("output.webp", output, 0644); err != nil {
		log.Printf("Ошибка записи: %v", err)
		return
	}

	fmt.Printf("Исходный формат: %s\n", proc.OriginalFormat())
	fmt.Printf("Конвертировано: %d байт -> %d байт (%.1f%%)\n",
		len(data), len(output), float64(len(output))/float64(len(data))*100)
}

// example3 демонстрирует batch обработку
func example3() {
	fmt.Println("\n=== Пример 3: Batch обработка ===")

	// Список файлов для обработки
	files := []string{"img1.jpg", "img2.png", "img3.jpg"}

	// Размеры для генерации
	sizes := []struct {
		width  int
		height int
		suffix string
	}{
		{800, 0, "_medium"},
		{400, 0, "_small"},
		{150, 0, "_thumb"},
	}

	for _, filename := range files {
		data, err := os.ReadFile(filename)
		if err != nil {
			log.Printf("Пропуск %s: %v", filename, err)
			continue
		}

		// Создать варианты разных размеров
		for _, size := range sizes {
			proc := ipxpress.New().
				FromBytes(data).
				Resize(size.width, size.height)

			if err := proc.Err(); err != nil {
				log.Printf("Ошибка обработки %s: %v", filename, err)
				continue
			}

			// Сохранить в оригинальном формате
			format := proc.OriginalFormat()
			output, err := proc.ToBytes(format, 85)
			proc.Close()
			if err != nil {
				log.Printf("Ошибка кодирования %s: %v", filename, err)
				continue
			}

			// Сформировать имя выходного файла
			outName := fmt.Sprintf("%s%s.%s",
				filename[:len(filename)-4],
				size.suffix,
				format)

			if err := os.WriteFile(outName, output, 0644); err != nil {
				log.Printf("Ошибка записи %s: %v", outName, err)
				continue
			}

			fmt.Printf("Создано: %s (%d байт)\n", outName, len(output))
		}
	}
}

// example4 демонстрирует использование с io.Reader
func example4() {
	fmt.Println("\n=== Пример 4: Работа с io.Reader ===")

	file, err := os.Open("input.jpg")
	if err != nil {
		log.Printf("Ошибка открытия файла: %v", err)
		return
	}
	defer file.Close()

	// Обработать из Reader
	proc := ipxpress.New().
		FromReader(file).
		Resize(500, 500)

	if err := proc.Err(); err != nil {
		log.Printf("Ошибка обработки: %v", err)
		return
	}

	output, err := proc.ToBytes(ipxpress.FormatPNG, 0)
	proc.Close()
	if err != nil {
		log.Printf("Ошибка кодирования: %v", err)
		return
	}

	if err := os.WriteFile("output.png", output, 0644); err != nil {
		log.Printf("Ошибка записи: %v", err)
		return
	}

	fmt.Println("PNG создан успешно")
}

// example5 демонстрирует обработку ошибок
func example5() {
	fmt.Println("\n=== Пример 5: Обработка ошибок ===")

	data := []byte("invalid image data")

	proc := ipxpress.New().
		FromBytes(data).
		Resize(800, 600)

	// Проверить ошибку после каждой операции
	if err := proc.Err(); err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		// Не нужно вызывать Close() если FromBytes провалился
		return
	}

	output, err := proc.ToBytes(ipxpress.FormatJPEG, 85)
	proc.Close()
	if err != nil {
		fmt.Printf("Ошибка кодирования: %v\n", err)
		return
	}

	fmt.Printf("Успешно: %d байт\n", len(output))
}
