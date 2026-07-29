# Клавиатуры и форматирование

В этом разделе описывается работа с клавиатурами (inline-кнопками), форматированием текста и встроенной системой пагинации в Yandex Bot API.

## Клавиатуры (Suggest Buttons)

API Яндекса поддерживает клавиатуры, которые отображаются под сообщением пользователя. Вы можете создавать их как одномерный список (в один ряд) или в виде двумерной сетки (в несколько рядов).

Для этого используются функции:
- `yabotapi.NewSuggestButtons(persist bool, buttons ...InlineSuggestButton)` — для создания простого ряда кнопок.
- `yabotapi.NewSuggestButtonsGrid(persist bool, rows ...[]InlineSuggestButton)` — для создания многоуровневой сетки кнопок.

### Флаг `Persist` (Поведение клавиатуры)
Первым аргументом в функции создания клавиатуры передается флаг `persist`, который отвечает за ее долговечность:
- `false` (Одноразовая): Клавиатура исчезает после того, как пользователь нажимает на любую кнопку или отправляет текстовое сообщение в чат. Это стандартное поведение для опросов или быстрых ответов.
- `true` (Постоянная): Клавиатура остается прикрепленной к полю ввода и не исчезает после взаимодействия. Пользователь может нажимать на кнопки многократно.

### Директивы и удобные хелперы кнопок

Для максимального упрощения создания кнопок в библиотеке предусмотрены готовые функции-конструкторы:

- **`yabotapi.NewSimpleActionButton(title, actionName string)`** — создает простую кнопку действия без нагрузки `payload`.
- **`yabotapi.NewActionButton(title, actionName string, payload any)`** — создает кнопку действия с контекстной полезной нагрузкой `payload`.
- **`yabotapi.NewURLButton(title, uri string)`** — создает кнопку-ссылку, открывающую внешнюю URL (`NewOpenURIDirective`).
- **`yabotapi.NewTextButton(title, text string)`** — создает кнопку быстрой отправки текста (`NewSendMessageDirective`).

Вы также можете конструировать кастомные директивы напрямую:
- `yabotapi.NewOpenURIDirective(uri string)`
- `yabotapi.NewSendMessageDirective(text string, payload any)`
- `yabotapi.NewServerActionDirective(name string, payload any)`
- `yabotapi.NewSetElementsStateDirective(ids []string, state string, timeout int)`

### Динамическое построение клавиатур (`KeyboardBuilder`)

Для удобной сборки кнопочных сеток в циклах или динамических меню используется `KeyboardBuilder`:

- `.Columns(cols int)` — задает максимальное число кнопок в одном ряду.
- `.AddSimpleButton(title, actionName)` — добавляет кнопку клика.
- `.AddButton(title, actionName, payload)` — добавляет кнопку с payload.
- `.AddURLButton(title, uri)` — добавляет кнопку-ссылку.
- `.AddTextButton(title, text)` — добавляет текстовую кнопку.
- `.AddRawButton(btn)` — добавляет произвольно сконструированную `InlineSuggestButton`.
- `.AddRow(buttons...)` — добавляет готовый ряд кнопок.
- `.Build() *SuggestButtons` — результирует готовую к отправке клавиатуру.

```go
kb := yabotapi.NewKeyboardBuilder(true).Columns(2) // 2 кнопки в каждом ряду
for _, product := range products {
    kb.AddButton(product.Name, "buy_product", product.ID)
}
keyboard := kb.Build()
```

### Пример создания клавиатуры
```go
package main

import (
	yabotapi "github.com/go-yandex-bot-api/yandex-bot-api"
	"github.com/go-yandex-bot-api/yandex-bot-api/types"
	"github.com/go-yandex-bot-api/yandex-bot-api/pkg/router"
)

func sendKeyboard(c *router.Context) error {
	keyboard := yabotapi.NewSuggestButtonsGrid(
		true, // persist = true (клавиатура не исчезнет после нажатия)
		[]types.InlineSuggestButton{
			yabotapi.NewSimpleActionButton("Поздороваться", "btn_hi"),
			yabotapi.NewSimpleActionButton("Попрощаться", "btn_bye"),
		},
		[]types.InlineSuggestButton{
			yabotapi.NewURLButton("Открыть Яндекс", "https://ya.ru"),
			yabotapi.NewTextButton("Отправить контакты", "Покажи контакты"),
		},
	)

	return c.ReplyWithKeyboard("Выберите действие:", keyboard)
}
}
```

## Форматирование текста

Яндекс Мессенджер использует свой диалект Markdown для разметки сообщений. Чтобы не запоминать синтаксис и случайно не забыть экранировать спецсимволы, библиотека предоставляет готовый пакет `pkg/format`.

```go
import "github.com/go-yandex-bot-api/yandex-bot-api/pkg/format"
```

### Доступные функции:
- `format.Escape(text string)` — Экранирование специальных символов Markdown (`*`, `_`, `~`, `+`, `` ` ``, `[`, `]`, `\`), чтобы они выводились как обычный текст.
- `format.Bold(text string)` — **Жирный** текст (`**text**`).
- `format.Italic(text string)` — __Курсив__ (`__text__`).
- `format.Strikethrough(text string)` — ~~Зачеркнутый~~ текст (`~~text~~`).
- `format.Underline(text string)` — ++Подчеркнутый++ текст (`++text++`).
- `format.Code(text string)` — Строчный фрагмент кода.
- `format.CodeBlock(text, language string)` — Многострочный блок кода с поддержкой подсветки синтаксиса.
- `format.Link(text, url string)` — [Кликабельная ссылка](url).

### Пример использования
```go
func sendFormattedText(c *router.Context) error {
	text := format.Bold("Внимание!") + "\n" +
		"Пожалуйста, ознакомьтесь с нашей документацией на " + format.Link("GitHub", "https://github.com") + "\n\n" +
		"Пример использования:\n" +
		format.CodeBlock("fmt.Println(\"Hello World\")", "go") + "\n" +
		"Пользователь с ником " + format.Escape("_user*name_") + " подключился."
		
	return c.Reply(text)
}
```

## Пагинация

Для упрощения создания страниц с результатами и удобной навигации по ним, в библиотеке предусмотрен пакет `pkg/pagination`.

```go
import "github.com/go-yandex-bot-api/yandex-bot-api/pkg/pagination"
```

Пакет содержит две основные функции:
1. `pagination.PaginateSlice(totalItems, page, limit int) (start, end, totalPages)` — Помогает безопасно вычислять границы среза (`slice[start:end]`) в зависимости от текущей страницы и лимита элементов на страницу. Защищает от выхода за границы массива.
2. `pagination.NewPaginationRow(currentPage, totalPages int, actionName string) []InlineSuggestButton` — Генерирует готовый ряд кнопок: `⬅️ Назад`, `Текущая / Всего`, `Вперед ➡️`. При нажатии на кнопки отправляется серверное действие (`ServerAction`) с названием `actionName` и payload-ом `pagination.Payload{Page: ...}`.

### Полный пример пагинации
```go
package main

import (
	"fmt"
	"log"

	yabotapi "github.com/go-yandex-bot-api/yandex-bot-api"
	"github.com/go-yandex-bot-api/yandex-bot-api/pkg/router"
	"github.com/go-yandex-bot-api/yandex-bot-api/pkg/pagination"
)

var dataItems = []string{"Яблоко", "Банан", "Апельсин", "Груша", "Киви", "Манго", "Персик", "Слива"}
const ItemsPerPage = 3

func sendItemsPage(c *router.Context, page int) error {
	// 1. Вычисляем безопасные границы для слайса
	start, end, totalPages := pagination.PaginateSlice(len(dataItems), page, ItemsPerPage)

	// 2. Формируем текст сообщения
	text := fmt.Sprintf("📄 Страница %d из %d:\n\n", page, totalPages)
	for i := start; i < end; i++ {
		text += fmt.Sprintf("- %s\n", dataItems[i])
	}

	// 3. Создаем ряд кнопок навигации. 
	// При нажатии будет вызван хендлер кнопки с именем "items_page"
	navRow := pagination.NewPaginationRow(page, totalPages, "items_page")

	// 4. Генерируем клавиатуру. Устанавливаем persist=true, чтобы она не исчезала.
	keyboard := yabotapi.NewSuggestButtonsGrid(true, navRow)

	return c.ReplyWithKeyboard(text, keyboard)
}

func main() {
	bot, _ := yabotapi.NewBot("TOKEN")
	r := router.NewRouter(bot)

	// Пользователь вызывает команду /start
	r.HandleCommand("start", func(c *router.Context) error {
		return sendItemsPage(c, 1)
	})

	// Хендлер для кнопок пагинации (имя совпадает с actionName из NewPaginationRow)
	r.HandleButton("items_page", func(c *router.Context) error {
		var payload pagination.Payload
		
		// Автоматически парсим JSON, прикрепленный к кнопке
		if err := c.BindPayload(&payload); err != nil {
			return err
		}
		
		// Отправляем запрошенную страницу
		return sendItemsPage(c, payload.Page)
	})

	// ... Запуск роутера
}
```
