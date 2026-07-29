# Конечные автоматы (FSM)

Конечный автомат (Finite State Machine, FSM) — это механизм, позволяющий боту запоминать контекст общения с пользователем и выстраивать диалог в виде последовательных шагов (например, пошаговый опрос или создание сложной сущности).

## Архитектура FSM

В библиотеке `yandex-bot-api` поддержка FSM интегрирована напрямую в пакет `router`.
Основная идея состоит в том, что каждому пользователю назначается определенное «состояние» (State) в виде строки, а также может сохраняться дополнительная информация (Payload/Data). 
В зависимости от того, в каком состоянии находится пользователь, `Router` будет направлять его сообщения в соответствующий обработчик.

Хранилище состояния абстрагировано через интерфейс `fsm.Storage`:

```go
type Storage interface {
	Set(userID string, state string)
	Get(userID string) string
	Delete(userID string)
	SetData(userID string, key string, value any)
	GetData(userID string, key string) (any, bool)
}
```

Вы можете использовать `fsm.NewMemoryStorage()` для хранения данных в оперативной памяти (с поддержкой TTL для автоматической очистки зависших состояний), или написать свою реализацию поверх Redis / PostgreSQL для распределенных систем.

### Внедрение хранилища

Для активации FSM необходимо создать экземпляр хранилища и передать его в роутер через метод `WithStorage()`:

```go
import "github.com/go-yandex-bot-api/yandex-bot-api/pkg/fsm"
import "time"

// Создаем хранилище в памяти с очисткой неактивных сессий через 30 минут
storage := fsm.NewMemoryStorage(30 * time.Minute)
// При завершении работы приложения рекомендуется вызвать Stop(), чтобы остановить сборщик мусора
defer storage.Stop()

// Передаем хранилище в роутер
r := router.NewRouter(bot).WithStorage(storage)
```

## Использование ContextFSM

В каждом обработчике доступен объект `c.FSM()`, который предоставляет изолированный контекст конечного автомата для конкретного отправителя (связки ChatID:Login).

Доступные методы:
* `c.FSM().SetState(state string)` — устанавливает новое состояние пользователя.
* `c.FSM().GetState() string` — возвращает текущее состояние.
* `c.FSM().SetData(key string, value any)` — сохраняет произвольные данные (ключ-значение).
* `c.FSM().GetData(key string) (any, bool)` — извлекает ранее сохраненные данные.
* `c.FSM().Clear()` — полностью удаляет и состояние, и сохраненные данные пользователя.

## Маршрутизация по состояниям

Роутер предоставляет специальные методы для перехвата сообщений от пользователей в определенных состояниях. В цикле роутинга обработчики состояний проверяются до обработчика обычного текста, но после обработчиков кнопок и команд (команды всегда выполняются приоритетно, что позволяет сбросить FSM в любой момент, написав `/cancel`).

* `HandleState(state string, handler router.HandlerFunc)` — обрабатывает апдейты для точного совпадения состояния.
* `HandleStateRegexp(pattern string, handler router.HandlerFunc) error` — обрабатывает апдейты для состояний, соответствующих регулярному выражению (полезно для динамических состояний, например `item_edit_.*`).

## Комплексный пример: Анкета (Questionnaire)

Ниже представлен полноценный пример создания пошаговой анкеты. Бот собирает имя, возраст и любимый цвет пользователя по шагам, после чего выводит итоговую карточку.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	yabotapi "github.com/go-yandex-bot-api/yandex-bot-api"
	"github.com/go-yandex-bot-api/yandex-bot-api/pkg/fsm"
	"github.com/go-yandex-bot-api/yandex-bot-api/pkg/router"
)

func main() {
	bot, err := yabotapi.NewBot("YOUR_TOKEN_HERE")
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}
	ctx := context.Background()

	updatesChannel, err := bot.Updates.GetUpdatesChannel(ctx, yabotapi.NewUpdateConfig(0))
	if err != nil {
		log.Fatal("Failed to start short-polling:", err)
	}

	// 1. Инициализируем хранилище (память очищается через 30 минут бездействия)
	storage := fsm.NewMemoryStorage(30 * time.Minute)
	defer storage.Stop()

	// 2. Подключаем хранилище к роутеру
	r := router.NewRouter(bot).WithStorage(storage)

	// Шаг 1: Пользователь отправляет команду /start -> спрашиваем имя
	r.HandleCommand("start", func(c *router.Context) error {
		// Переводим пользователя в состояние "step_name"
		c.FSM().SetState("step_name")
		return c.Reply("Привет! Давай познакомимся. Как тебя зовут?")
	})

	// Шаг 2: Обработка имени -> спрашиваем возраст
	r.HandleState("step_name", func(c *router.Context) error {
		name := c.Update.Text

		// Сохраняем имя в полезную нагрузку (payload)
		c.FSM().SetData("name", name)

		// Переводим пользователя на следующий шаг
		c.FSM().SetState("step_age")
		return c.Reply(fmt.Sprintf("Приятно познакомиться, %s! Сколько тебе лет?", name))
	})

	// Шаг 3: Обработка возраста -> спрашиваем любимый цвет
	r.HandleState("step_age", func(c *router.Context) error {
		age := c.Update.Text

		// Сохраняем возраст
		c.FSM().SetData("age", age)
		
		// Переводим пользователя на следующий шаг
		c.FSM().SetState("step_color")
		return c.Reply("Понял. А какой твой любимый цвет?")
	})

	// Шаг 4: Обработка цвета -> выводим результат и очищаем FSM
	r.HandleState("step_color", func(c *router.Context) error {
		color := c.Update.Text

		// Извлекаем ранее сохраненные данные
		var name string
		var age string

		if val, ok := c.FSM().GetData("name"); ok {
			name, _ = val.(string)
		}
		if val, ok := c.FSM().GetData("age"); ok {
			age, _ = val.(string)
		}

		// Отправляем итоговый результат
		summary := fmt.Sprintf("📝 **Твоя анкета**\nИмя: %s\nВозраст: %s\nЦвет: %s\n\nПрофиль сохранен!", name, age, color)
		err := c.Reply(summary)

		// Полностью очищаем состояние и сохраненные данные
		c.FSM().Clear()
		return err
	})

	// Резервный обработчик: если пользователь вне состояния пишет текст
	r.HandleDefault(func(c *router.Context) error {
		if c.Update.Text != "" && !c.Update.IsCommand() {
			return c.Reply("Отправьте /start, чтобы начать заполнение анкеты.")
		}
		return nil
	})

	log.Println("Бот-анкета запущен. Отправьте /start")
	r.Start(ctx, updatesChannel)
}
```
