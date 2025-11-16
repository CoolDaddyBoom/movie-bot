КРОК 1 - Розумові питання:

1. Мій бот повинен звертатись до АПІ телеграма, кожну секунду і питати чи є нові повідомлення (getUpdates)

2. Це звернення повинно бути постійним, тобто, повинен крутитись цикл:

повтор:
    запитати "є нові повідомлення?"
    якщо є - обробити
    почекати 1 секунду
    йти на повтор

3. Потрібно розпарсити відповідь від телеграма, і взяти з неї ІД чату та текст, що відправив користувач

4. Зробити якісь операції (збереження фільму в БД/ діставання з БД + видалення)

5. Відповісти користовачу - запит до іншого API Telegram: `sendMessage`
- ID чату (щоб знати КОМУ відправити)
- Текст повідомлення

КРОК 2 - Логіка: 

ГОЛОВНИЙ ЦИКЛ:
    повтор завжди:
        1. Запитати Telegram: "дай нові повідомлення"
        2. Для кожного повідомлення:
            - Подивитись що користувач написав
            - Якщо "/start" → відповісти "Привіт"
            - Якщо "/random" → дати випадковий фільм
            - Якщо "/help" → показати команди
            - Інакше → зберегти як назву фільму
        3. Почекати 1 секунду


movie-bot/
├── main.go                    ← точка входу (створює все, запускає)
├── go.mod                     ← залежності
│
├── telegram/                  ← все про Telegram
│   ├── client.go             ← HTTP запити (GetUpdates, SendMessage)
│   ├── types.go              ← структури (Response, Update, Message)
│   └── processor.go          ← обробка команд (Process)
│
├── storage/                   ← все про БД
│   ├── storage.go            ← інтерфейс Storage
│   └── sqlite/
│       └── sqlite.go         ← реалізація SQLite
│
└── consumer/
    └── consumer.go           ← головний цикл (EventConsumer)

    1. Користувач пише "Інтерстеллар"
         ↓
2. Telegram зберігає це повідомлення
         ↓
3. EventConsumer запитує: "Дай updates"
         ↓
4. TelegramClient робить HTTP запит
         ↓
5. Telegram повертає JSON:
   {"result": [{"update_id": 123, "message": {"text": "Інтерстеллар", "chat": {"id": 456}}}]}
         ↓
6. TelegramClient парсить JSON → Update struct
   Update {
       UpdateID: 123
       Message: {Text: "Інтерстеллар", Chat: {ID: 456}}
   }
         ↓
7. EventConsumer передає Update в Processor
         ↓
8. Processor дивиться:
   - Це не команда
   - Значить це назва фільму
   - Треба зберегти
         ↓
9. Processor викликає Storage.AddMovie("Інтерстеллар", username)
         ↓
10. Storage робить SQL запит:
    INSERT INTO movies (title, username) VALUES ("Інтерстеллар", "vasya")
         ↓
11. Processor викликає TelegramClient.SendMessage(456, "Збережено ✓")
         ↓
12. TelegramClient робить HTTP запит до Telegram
         ↓
13. Користувач бачить "Збережено ✓"

main.go
  ↓ створює
  ├─ TelegramClient
  ├─ Storage (SQLite)
  ├─ Processor (передає йому Client і Storage)
  └─ EventConsumer (передає йому Client і Processor)
     ↓ запускає
     EventConsumer.Start() ← безкінечний цикл
```

### Залежності:
```
EventConsumer
  ↓ використовує
  ├─ TelegramClient (щоб отримувати updates)
  └─ Processor (щоб обробляти updates)

Processor
  ↓ використовує
  ├─ TelegramClient (щоб відправляти повідомлення)
  └─ Storage (щоб зберігати/діставати фільми)

TelegramClient
  └─ нікого не використовує (тільки http пакет)

Storage
  └─ нікого не використовує (тільки database/sql)