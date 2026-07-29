# Yandex Messenger Bot API for Go

[![CI](https://github.com/go-yandex-bot-api/yandex-bot-api/actions/workflows/ci.yml/badge.svg)](https://github.com/go-yandex-bot-api/yandex-bot-api/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/go-yandex-bot-api/yandex-bot-api.svg)](https://pkg.go.dev/github.com/go-yandex-bot-api/yandex-bot-api)
[![Go Version](https://img.shields.io/github/go-mod/go-version/go-yandex-bot-api/yandex-bot-api)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

`yandex-bot-api` — библиотека для разработки ботов в Яндекс Мессенджере на языке Go.

Предоставляет удобное взаимодействие с API Яндекс Мессенджера, роутер, пагинацию и управление состояниями (анкетами).

[Read in English](README_EN.md)

## Возможности

- **Полное покрытие API Яндекса**: Отправляйте простой текст, прикрепляйте файлы, загружайте галереи изображений, создавайте интерактивные опросы и сложные инлайн-клавиатуры.
- **Встроенный Роутер**: Элегантная и строгая маршрутизация входящих сообщений (`HandleCommand`, `HandleText`, `HandleButton`). Позволяет отказаться от гигантских нечитаемых конструкций `switch-case`.
- **Машина состояний (FSM)**: Удобное управление сессиями пользователей. Создавайте сложные многошаговые диалоги (например, поэтапные анкеты или корзины покупок), безопасно сохраняя данные прямо в контекст пользователя через `Context.FSM()`.
- **Потоковая передача данных**: Поддержка интерфейса `io.Reader` для загрузки файлов без предварительного сохранения на жесткий диск. Идеально для мгновенной отправки динамически сгенерированных отчетов или графики.
- **Middlewares (Промежуточные слои)**: Возможность внедрять глобальные перехватчики на уровне роутера. Отлично подходит для добавления логгеров, авторизации или систем сбора метрик (Prometheus/Grafana).
- **Два режима получения данных**: Библиотека из коробки поддерживает как Long Polling (отлично для локальной разработки), так и Webhook (идеально для масштабируемых Production-систем).

## Установка

Убедитесь, что у вас установлен Go версии 1.23 или выше, и выполните:

```bash
go get -u github.com/go-yandex-bot-api/yandex-bot-api
```

## Быстрый старт

Ниже представлен код простого эхо-бота. Обратите внимание на то, как объект `router.Context` берет на себя всю рутину: он сам понимает, кто прислал запрос, и отправляет ответ в нужный чат без необходимости вручную оперировать ID чатов.

```go
package main

import (
	"context"
	"log"

	yabotapi "github.com/go-yandex-bot-api/yandex-bot-api"
	"github.com/go-yandex-bot-api/yandex-bot-api/pkg/router"
)

func main() {
	// 1. Инициализация HTTP-клиента бота с вашим токеном
	bot, err := yabotapi.NewBot("YOUR_YANDEX_TOKEN")
	if err != nil {
		log.Fatal(err)
	}

	// 2. Запуск фонового получения обновлений (Long Polling)
	updates, err := bot.Updates.GetUpdatesChannel(context.Background(), yabotapi.NewUpdateConfig(0))
	if err != nil {
		log.Fatal(err)
	}

	// 3. Создание роутера для обработки сообщений
	r := router.NewRouter(bot)

	// Обработка конкретной команды (например, /start)
	r.HandleCommand("start", func(c *router.Context) error {
		return c.Reply("Привет! Я модульный бот для Яндекс Мессенджера.")
	})

	// Обработка любого произвольного текста
	r.HandleText(func(c *router.Context) error {
		return c.Reply("Эхо: " + c.Update.Text)
	})

	// 4. Запуск блокирующего цикла маршрутизации
	log.Println("Бот успешно запущен и готов к работе")
	r.Start(context.Background(), updates)
}
```

## Архитектура проекта

Кодовая база спроектирована по принципам чистой архитектуры и разделена на строгие, независимые пакеты:

```text
github.com/go-yandex-bot-api/yandex-bot-api/
├── api/             # Пакеты, реализующие все публичные методы API Яндекса
├── core/            # Ядро библиотеки: конфигурация HTTP-клиента и отправка запросов
├── pkg/             # Инструменты для разработчика (SDK Utilities):
│   ├── fsm/         # Интерфейсы и In-Memory хранилище конечных автоматов
│   └── router/      # Маршрутизатор (Router), структура Context и абстракции Middlewares
├── types/           # Глобальные структуры данных (Update, Message, Keyboard и др.)
├── examples/        # Эталонные примеры реализации различных ботов
├── aliases.go       # Удобные алиасы типов для чистоты импортов в ваших проектах
└── yabotapi.go      # Главный фасад библиотеки и конструкторы
```

## Доступные сервисы

Для удобства использования, все методы API разбиты на предметные сервисы. После инициализации бота вы можете обращаться к ним напрямую через поля структуры `bot`:

* `bot.Updates` — Сервис для получения входящих событий (Polling).
* `bot.Webhooks` — Сервис для установки, удаления и обработки Webhook-запросов от серверов Яндекса.
* `bot.Messages` — Основной сервис для отправки текста, форматированного Markdown, и инлайн-клавиатур.
* `bot.Files` — Сервис для работы с медиа (загрузка и скачивание аудио, видео, изображений).
* `bot.Polls` — Сервис для создания голосований и опросов.
* `bot.Info` — Сервис для получения информации об участниках и свойствах групповых чатов.
* `bot.Users` — Сервис для работы с профилями пользователей.

## Тонкая настройка бота

### 1. Настройка HTTP-клиента
При создании экземпляра бота вы можете передать дополнительные опции с помощью паттерна Functional Options. Это позволяет тонко настроить сетевое взаимодействие:

```go
bot, err := yabotapi.NewBot(
    "TOKEN",
    yabotapi.WithDebug(true), // Включение детального дампа HTTP-запросов (только для отладки!)
)
if err != nil {
    log.Fatal(err)
}
```

### 2. Режимы работы: Polling vs Webhook

Библиотека предоставляет оба стандарта получения данных. Выбор зависит от архитектуры вашего проекта:

**Long Polling (Подходит для локальной разработки)**  
Бот сам отправляет запросы к серверам Яндекса, чтобы узнать, есть ли новые сообщения.
```go
cfg := yabotapi.UpdateConfig{
    Offset: 0,
    Limit:  100, // Пакетная загрузка: до 100 сообщений за один сетевой запрос
}
updates, _ := bot.Updates.GetUpdatesChannel(context.Background(), cfg)
```

**Webhook (Стандарт для Production)**  
Яндекс Мессенджер сам отправляет HTTP-запросы на ваш сервер в момент, когда пользовать пишет боту. Это экономит ресурсы и снижает задержку (latency).
```go
// 1. Сообщаем Яндексу адрес вашего сервера
bot.Webhooks.Set(context.Background(), webhooks.SetRequest{
    URL: "https://api.your-domain.com/yandex/webhook",
})

// 2. Поднимаем стандартный Go HTTP-хэндлер для обработки входящих POST-запросов
http.HandleFunc("/yandex/webhook", bot.Webhooks.ListenForWebhook(updatesChan))
```

## Готовые примеры реализации
В директории [`/examples`](./examples) находятся готовые, компилируемые примеры ботов для решения реальных задач:
- **`07_fsm_questionnaire`** — Бот-анкета с пошаговым сохранением состояний пользователя.
- **`08_pagination`** — Использование инлайн-клавиатур и сложных JSON-пейлоадов для создания страниц.
- **`09_project_structure`** — Пример того, как правильно разделять код обработчиков по разным файлам (Dependency Injection паттерн).
