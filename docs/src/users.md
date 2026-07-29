# Пользователи и Информация о Боте (`bot.Users`)

Сервис `bot.Users` позволяет получать сведения о самом боте, а также генерировать ссылки (диплинки) на диалоги с пользователями.

## Получение сведений о боте (`GetMe` / `GetSelf`)

Метод `GetMe` (или его алиас `GetSelf`) возвращает информацию о текущем боте (имя, логин):

```go
botInfo, err := bot.Users.GetMe(ctx)
if err != nil {
    log.Fatalf("Ошибка получения данных бота: %v", err)
}

fmt.Printf("Имя бота: %s, Логин: %s\n", botInfo.Name, botInfo.Login)
```

## Получение диплинка пользователя (`GetUserLink`)

Метод `GetUserLink` возвращает прямые ссылки на открытый чат или звонок с пользователем по его логину:

```go
linkResp, err := bot.Users.GetUserLink(ctx, yabotapi.GetUserLinkRequest{
    Login: "john_doe",
})
if err != nil {
    log.Printf("Ошибка получения ссылки на пользователя: %v", err)
}

fmt.Printf("Ссылка на чат: %s\n", linkResp.ChatURL)
```
