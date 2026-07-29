# Введение

`yandex-bot-api` — библиотека для разработки ботов в Яндекс Мессенджере на языке Go. Предоставляет удобное взаимодействие с API Яндекс Мессенджера, роутер, пагинацию и управление состояниями (анкетами).

## Установка

Для установки библиотеки в ваш проект используйте команду `go get`:

```bash
go get github.com/go-yandex-bot-api/yandex-bot-api
```

Требуется версия Go 1.23 или выше.

## Быстрый старт (Эхо-бот)

Ниже представлен минимальный пример эхо-бота, который использует механизм Short-Polling для получения сообщений и отправляет пользователю его же текст в ответ.

```go
package main

import (
	"context"
	"log"

	yabotapi "github.com/go-yandex-bot-api/yandex-bot-api"
)

func main() {
	// 1. Инициализация бота. Рекомендуется включать WithDebug при разработке.
	bot, err := yabotapi.NewBot("YOUR_TOKEN_HERE", yabotapi.WithDebug(true))
	if err != nil {
		log.Fatal("Ошибка инициализации бота:", err)
	}
	
	ctx := context.Background()

	// 2. Запуск Short-Polling
	updatesChannel, err := bot.Updates.GetUpdatesChannel(ctx, yabotapi.NewUpdateConfig(0))
	if err != nil {
		log.Fatal("Ошибка запуска пуллинга:", err)
	}

	log.Println("Бот запущен в режиме Short-Polling... Нажмите Ctrl+C для остановки.")

	// 3. Обработка входящих обновлений в цикле
	for update := range updatesChannel {
		log.Printf("Получено обновление с ID: %d", update.UpdateID)

		// Если это текстовое сообщение (а не системное или callback от кнопки)
		if update.Text != "" {
			// Формируем ответ (эхо)
			reply := yabotapi.NewReply(update, "Вы сказали: "+update.Text)
			
			// Отправляем сообщение обратно
			if _, err := bot.Messages.SendText(ctx, reply); err != nil {
				log.Println("Ошибка отправки ответа:", err)
			}
		}
	}
}
```
