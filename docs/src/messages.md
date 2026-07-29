# Работа с сообщениями

В библиотеке **yandex-bot-api** работа с сообщениями реализована через сервис `Messages`, доступный у экземпляра бота (`bot.Messages`). Однако, при использовании встроенного роутера (`pkg/router`), удобнее всего пользоваться методами контекста `*router.Context` (например, `c.Reply()`), так как они автоматически определяют, куда (какому пользователю или в какой чат) отправлять ответ, а также учитывают контекст тредов.

---

## 1. Отправка текстовых сообщений

### Использование `bot.Messages.SendText` (Низкоуровневый подход)
Метод `SendText` позволяет гибко настраивать отправляемое сообщение. Вы можете отправить сообщение как в определенный чат (по `chat_id`), так и конкретному пользователю (по `login`). 

```go
package main

import (
	"context"
	"log"

	yabotapi "github.com/go-yandex-bot-api/yandex-bot-api"
	"github.com/go-yandex-bot-api/yandex-bot-api/api/messages"
	"github.com/go-yandex-bot-api/yandex-bot-api/types"
)

func main() {
	bot, _ := yabotapi.NewBot("YOUR_TOKEN_HERE")
	ctx := context.Background()

	// Отправка сообщения в чат по его ID
	resp, err := bot.Messages.SendText(ctx, messages.SendTextRequest{
		ChatID: "0/0/chat_id_here",
		Text:   "Привет из Yandex Bot API!",
		// Опциональные параметры:
		// Important: true,
		// DisableNotification: true,
	})
	if err != nil {
		log.Println("Ошибка отправки:", err)
		return
	}
	
	log.Println("Сообщение отправлено, ID:", resp.MessageID)
}
```

> **Важно**: Согласно специфике Яндекса, вы должны указать либо `ChatID`, либо `Login` получателя.

### Использование `router.Context.Reply` (Удобный подход)
Если вы используете роутер, контекст автоматически подставит нужный `ChatID` или `Login`, чтобы ответить туда же, откуда пришло сообщение.

```go
r.HandleCommand("start", func(c *router.Context) error {
	// Отправит текст в тот же чат/пользователю
	return c.Reply("Привет! Я бот. 👋")
})
```

Для отправки клавиатур используется `c.ReplyWithKeyboard`:
```go
r.HandleButton("btn_hi", func(c *router.Context) error {
	// Отправка сообщения вместе с клавиатурой
	return c.ReplyWithKeyboard("Выберите действие:", keyboard)
})
```

---

## 2. Системные сообщения

Системные сообщения обычно выглядят как сервисные уведомления в чате (как правило, центрированные). Для них используется метод `SendSystemMessage`.

```go
resp, err := bot.Messages.SendSystemMessage(ctx, messages.SendSystemMessageRequest{
	ChatID: "0/0/chat_id_here",
	Text:   "Пользователь Иван присоединился к чату",
})
```

---

## 3. Отправка стикеров

Стикеры отправляются с помощью метода `SendSticker`. Для этого нужно знать ID стикерпака (`StickerSetID`) и ID самого стикера (`StickerID`).

```go
resp, err := bot.Messages.SendSticker(ctx, messages.SendStickerRequest{
	ChatID:       "0/0/chat_id_here",
	StickerSetID: "yandex_stickers_id",
	StickerID:    "sticker_123",
	// Можно передать ReplyMessageID, чтобы стикер был отправлен как реплай
})
```

---

## 4. Индикатор набора текста

Если ваш бот выполняет долгую операцию, рекомендуется отправить индикатор набора текста (typing), чтобы пользователь понимал, что бот "думает".

```go
err := bot.Messages.SendTyping(ctx, messages.SendTypingRequest{
	ChatID: "0/0/chat_id_here",
})
// После этого выполняем долгую работу и затем присылаем ответ
```

---

## 5. Управление сообщениями (Удаление, Закрепление, Открепление)

API предоставляет возможности для управления уже отправленными сообщениями. Все эти методы принимают ID чата/пользователя и `MessageID` целевого сообщения.

### Удаление сообщения
```go
err := bot.Messages.Delete(ctx, messages.DeleteMessageRequest{
	ChatID:    "0/0/chat_id_here",
	MessageID: 123456789,
})
```

### Закрепление сообщения (Pin)
```go
err := bot.Messages.Pin(ctx, messages.PinMessageRequest{
	ChatID:    "0/0/chat_id_here",
	MessageID: 123456789,
})
```

### Открепление сообщения (Unpin)
```go
err := bot.Messages.Unpin(ctx, messages.UnpinMessageRequest{
	ChatID:    "0/0/chat_id_here",
	MessageID: 123456789,
})
```

---

## Резюме
- Используйте методы `router.Context` (`c.Reply`, `c.ReplyWithKeyboard`) для быстрого ответа на входящие обновления.
- Используйте методы сервиса `bot.Messages` (`SendText`, `SendSticker`, `Delete` и т.д.) для фоновой отправки, отложенных рассылок или более тонкого управления параметрами сообщения.
