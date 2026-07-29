# Ядро и получение обновлений

В этом разделе описывается процесс инициализации экземпляра бота и два основных способа получения обновлений от серверов Яндекс Мессенджера: **Short-Polling** и **Webhooks**.

## Инициализация бота

Точкой входа в библиотеку является функция `yabotapi.NewBot(token, options...)`. Она возвращает инициализированный объект бота и ошибку, если токен пустой.

Конструктор поддерживает паттерн опций (`Option`), позволяя гибко настраивать внутренний HTTP-клиент `core.Client`:

- `yabotapi.WithDebug(true)` — включает логирование всех входящих и исходящих HTTP-запросов (полезно для отладки). **Внимание:** логгер автоматически скрывает токен авторизации (`***REDACTED***`), чтобы предотвратить его утечку, и предотвращает `OOM` при загрузке больших файлов, логируя тело только для `application/json`.
- `yabotapi.WithClient(client)` — позволяет передать собственный `http.Client`.
- `yabotapi.WithMaxRetries(retries)` — настраивает количество автоматических повторных попыток при сетевых ошибках или HTTP 429 (по умолчанию 3).
- `yabotapi.WithAPIURL(url)` — позволяет переопределить базовый URL API Яндекса (по умолчанию `https://botapi.messenger.yandex.net/bot/v1/`).
- `yabotapi.WithErrorHandlingConfig(config)` — позволяет переопределить логику того, на какие статусы ответов нужно делать retry (по умолчанию `429` и `>= 500`).

```go
bot, err := yabotapi.NewBot("YOUR_TOKEN_HERE", 
	yabotapi.WithDebug(true),
	yabotapi.WithMaxRetries(5),
)
if err != nil {
	log.Fatal("Не удалось создать бота:", err)
}
```

---

## Получение обновлений

Вы можете получать обновления от Яндекса двумя способами. Библиотека возвращает `<-chan types.Update`, из которого удобно вычитывать события в цикле.

### 1. Short-Polling (Пуллинг)

Идеально подходит для локальной разработки или в условиях, когда у вашего сервера нет публичного белого IP и SSL-сертификата. 
В библиотеке реализован **надежный Short-Polling** с использованием Jitter-backoff (от 500 мс до 5 секунд) при отсутствии обновлений или сетевых сбоях, чтобы избежать "spin-loop" и лишней нагрузки на CPU и сеть (поскольку серверы Яндекса сразу закрывают соединение, если новых событий нет).

```go
// Конфиг пуллинга, начиная со смещения 0
config := yabotapi.NewUpdateConfig(0)

updatesChannel, err := bot.Updates.GetUpdatesChannel(ctx, config)
if err != nil {
	log.Fatal("Не удалось запустить пуллинг:", err)
}

// Чтение канала (блокирующая операция)
for update := range updatesChannel {
	// Обработка update
}
```
*Обратите внимание: Если для бота ранее был установлен Webhook, метод `GetUpdatesChannel` вернет ошибку. Сначала нужно удалить вебхук.*

### 2. Webhooks (Вебхуки)

Рекомендуемый подход для production-окружений. В этом режиме Яндекс сам отправляет HTTP POST запросы на ваш сервер при наступлении событий.

Библиотека предоставляет встроенный `http.HandlerFunc`, который безопасно читает JSON, предотвращает переполнение памяти (используя `http.MaxBytesReader`) и отдает `503 Service Unavailable`, если ваш канал обработки переполнен (чтобы Яндекс повторил отправку позже).

```go
webhookURL := "https://your-domain.com/webhook"

// 1. Сообщаем Яндексу наш URL
if err := bot.Webhooks.SetWebhook(ctx, webhookURL); err != nil {
	log.Fatal("Ошибка установки вебхука:", err)
}

// 2. Получаем канал и HTTP-обработчик от библиотеки
updatesChannel, handler := bot.Webhooks.ListenForWebhook(ctx)

// 3. Регистрируем обработчик в стандартном роутере
http.HandleFunc("/webhook", handler)

// 4. Запускаем сервер
go func() {
	srv := &http.Server{Addr: ":8080"}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}()

// 5. Обрабатываем обновления из канала, точно так же как и в Short-Polling
for update := range updatesChannel {
	// Обработка update
}
```
