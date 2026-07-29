# Работа с чатами и опросами

В этом разделе описано, как использовать Yandex Bot API для управления чатами (создание, получение списка участников, обновление ролей) и создания опросов, а также получения их результатов.

## Управление чатами

Все методы для работы с чатами доступны через сервис `bot.Chats`.

### Создание чата или канала

Для создания чата используется метод `CreateChat`. Обратите внимание на важную особенность Yandex API: поля `Channel` и `Public` имеют тип `bool` без тега `omitempty`. Это значит, что вы можете явно передавать `false` для создания нужного типа чата.

**Ключевые комбинации `Channel` и `Public`:**
- `Channel: false, Public: false` — Создает закрытую группу.
- `Channel: false, Public: true` — Создает открытую группу (доступна по ссылке).
- `Channel: true, Public: false` — Создает закрытый канал.
- `Channel: true, Public: true` — Создает открытый канал.

**Внимание:** В зависимости от значения `Channel`, для добавления пользователей при создании необходимо использовать разные поля:
- Если `Channel: false`, используйте поле `Members`.
- Если `Channel: true`, используйте поле `Subscribers`.

Пример создания закрытой группы:

```go
package main

import (
	"context"
	"log"

	"github.com/go-yandex-bot-api/yandex-bot-api/api/chats"
	yabotapi "github.com/go-yandex-bot-api/yandex-bot-api"
)

func main() {
	bot, _ := yabotapi.NewBot("YOUR_TOKEN_HERE")
	ctx := context.Background()

	req := chats.CreateChatRequest{
		Name:        "Новая закрытая группа",
		Description: "Описание группы",
		Channel:     false, // Создаем группу, а не канал
		Public:      false, // Закрытая
		Members: []chats.User{
			{Login: "user_login_1"},
		},
	}

	resp, err := bot.Chats.CreateChat(ctx, req)
	if err != nil {
		log.Fatalf("Ошибка создания чата: %v", err)
	}

	log.Printf("Чат создан! ID: %s", resp.ChatID)
}
```

### Получение списка чатов бота (`GetChats`)

Для получения списка всех чатов и каналов, в которых состоит бот, используется метод `GetChats`:

```go
chats, err := bot.Chats.GetChats(ctx, chats.GetChatsRequest{}.WithLimit(50))
if err != nil {
	log.Fatalf("Ошибка получения чатов: %v", err)
}

for _, c := range chats {
	log.Printf("Чат: %s, ID: %s, Канал: %v", c.Name, c.ID, c.Channel)
}
```

### Получение списка участников и администраторов

Для получения участников чата используйте метод `GetMembers`. Можно отфильтровать пользователей по роли (например, получить только администраторов).

```go
req := chats.GetMembersRequest{
	ChatID: "chat_id_here",
	Role:   "admin", // Фильтрация по роли. Оставьте пустым для получения всех.
}.WithLimit(100) // Используем Builder метод для опциональных полей

members, err := bot.Chats.GetMembers(ctx, req)
if err != nil {
	log.Fatal(err)
}

for _, member := range members {
	log.Printf("Участник: %s, Роль: %s", member.Login, member.Role)
}
```

### Добавление и удаление участников

Для обновления участников, администраторов и подписчиков используется метод `UpdateMembers`.

> **Важно:** По умолчанию сервер Яндекса отправляет уведомления в чат о добавлении/удалении пользователей. Чтобы сделать это "тихо" (без системных сообщений), необходимо явно использовать метод-билдер `WithSendNotifications(false)`.

```go
req := chats.UpdateMembersRequest{
	ChatID: "chat_id_here",
	Members: []chats.User{
		{Login: "new_member_login"},
	},
}.WithSendNotifications(false) // Тихое добавление

err := bot.Chats.UpdateMembers(ctx, req)
if err != nil {
	log.Fatal(err)
}
```

---

## Работа с опросами (Polls)

API опросов доступно через `bot.Polls`. Опросы можно отправлять как в группы, так и в личные сообщения пользователям.

### Создание опроса

При создании опроса необходимо указать вопрос (`Title`) и список вариантов ответа (`Answers`). В качестве получателя нужно передать либо `ChatID` (для отправки в группу), либо `Login` (для личного сообщения).

```go
package main

import (
	"context"
	"log"

	yabotapi "github.com/go-yandex-bot-api/yandex-bot-api"
	"github.com/go-yandex-bot-api/yandex-bot-api/api/polls"
)

func main() {
	bot, _ := yabotapi.NewBot("YOUR_TOKEN_HERE")
	
	// Отправка опроса в чат
	req := polls.CreateRequest{
		ChatID:  "chat_id_here",
		Title:   "Какой ваш любимый язык программирования?",
		Answers: []string{"Go", "Python", "Rust", "Java"},
	}

	resp, err := bot.Polls.Create(context.Background(), req)
	if err != nil {
		log.Fatal("Ошибка отправки опроса:", err)
	}
	
	log.Printf("Опрос создан! Message ID: %v", resp.MessageID)
}
```

### Получение результатов опроса

Если вам нужно узнать количество голосов и общую статистику по ответам, используйте метод `GetResults`, передав ID сообщения опроса:

```go
req := polls.GetResultsRequest{
	ChatID:    "chat_id_here",
	MessageID: 123456789,
}

resp, err := bot.Polls.GetResults(context.Background(), req)
if err != nil {
	log.Fatal(err)
}

log.Printf("Всего голосов: %d", resp.VotedCount)
for answerText, votes := range resp.Answers {
	log.Printf("Ответ '%s': %d голосов", answerText, votes)
}
```

### Получение списка проголосовавших (Voters)

Для получения поименного списка пользователей, проголосовавших за конкретный вариант ответа, используется метод `GetVoters`. Этот метод поддерживает пагинацию с помощью курсоров.

В запросе необходимо указать `AnswerID` — индекс ответа (начиная с 0, согласно порядку вариантов при создании опроса).

```go
// Указываем индекс ответа (например, 0 для первого варианта "Go")
answerID := 0

req := polls.GetVotersRequest{
	ChatID:    "chat_id_here",
	MessageID: 123456789,
	AnswerID:  &answerID,
}.WithLimit(50)

for {
	resp, err := bot.Polls.GetVoters(context.Background(), req)
	if err != nil {
		log.Fatal(err)
	}

	for _, vote := range resp.Votes {
		log.Printf("Пользователь %s проголосовал! (Таймстемп: %d)", vote.User.Login, vote.Timestamp)
	}

	// Получаем следующий курсор. Если пустой - значит дошли до конца
	nextCursor := resp.NextCursor()
	if nextCursor == "" {
		break
	}
	
	// Устанавливаем курсор для следующего запроса
	req.Cursor = nextCursor
}
```

### Обработка появления опросов (`r.HandlePoll`) и подчисление голосов

При создании нового опроса в чате Яндекс Мессенджер присылает объект `Poll` в `Update`. Роутер предоставляет удобный метод подписки:

```go
r.HandlePoll(func(c *router.Context) error {
	poll := c.Update.Poll
	log.Printf("В чате появился новый опрос: '%s' с вариантами %v", poll.Title, poll.Answers)
	return nil
})
```

> **Важно:** Сервер Яндекс Мессенджера не стримит Push-уведомления через Long-Polling (`getUpdates`) на каждый клик ответа в опросе. Для получения актуального количества голосов и списков проголосовавших пользователей используйте вызовы `bot.Polls.GetResults` и `bot.Polls.GetVoters`.

