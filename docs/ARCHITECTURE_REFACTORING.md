# Рефакторинг архитектуры: разделение слоев

## Проблема

Изначально пакет `services` содержал смешанную логику:
- Бизнес-логику аутентификации (`AuthService`)
- Адаптеры для внешних сервисов (`RabbitMQAdapter`)
- Интерфейсы для разных доменов

Это нарушало принципы Clean Architecture и Single Responsibility Principle.

## Решение

Разделили `services` на логические слои:

### 1. `services/` - Application Layer (бизнес-логика)
```
services/
├── interfaces.go          # Интерфейсы бизнес-логики
├── auth_service.go        # Сервис аутентификации
├── auth_service_test.go   # Тесты сервиса
└── mocks/                 # Моки для тестирования
```

**Ответственности:**
- Бизнес-логика аутентификации
- Валидация данных
- Генерация JWT токенов
- Координация между репозиториями и messaging

### 2. `messaging/` - Infrastructure Layer (внешние сервисы)
```
messaging/
├── interfaces.go              # Интерфейсы messaging
├── rabbitmq_adapter.go        # Адаптер RabbitMQ
├── rabbitmq_adapter_test.go   # Тесты адаптера
└── mocks/                     # Моки для тестирования
```

**Ответственности:**
- Адаптеры для брокеров сообщений
- Публикация событий
- Управление соединениями

### 3. `repositories/` - Infrastructure Layer (база данных)
```
repositories/
├── interfaces.go          # Интерфейсы репозиториев
├── user_repository.go     # Репозиторий пользователей
├── gorm_adapter.go        # Адаптер GORM
└── mocks/                 # Моки для тестирования
```

**Ответственности:**
- Доступ к данным
- Адаптеры для ORM
- CRUD операции

## Преимущества новой архитектуры

### 1. Разделение ответственностей
- Каждый пакет отвечает за свою область
- Легче понимать и поддерживать код
- Проще тестировать отдельные компоненты

### 2. Следование принципам Clean Architecture
```
┌─────────────────────────────────────┐
│           Application               │
│  ┌─────────────────────────────┐    │
│  │         services/           │    │
│  │    - AuthService            │    │
│  │    - Business Logic         │    │
│  └─────────────────────────────┘    │
└─────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────┐
│           Infrastructure            │
│  ┌─────────────┐  ┌─────────────┐   │
│  │ messaging/  │  │repositories/│   │
│  │ - RabbitMQ  │  │ - GORM      │   │
│  │ - Events    │  │ - Database  │   │
│  └─────────────┘  └─────────────┘   │
└─────────────────────────────────────┘
```

### 3. Улучшенная тестируемость
- Каждый слой можно тестировать независимо
- Моки генерируются автоматически
- 100% покрытие тестами без внешних зависимостей

### 4. Легкость расширения
- Добавление новых брокеров сообщений (Kafka, Redis)
- Замена ORM (GORM → SQLx)
- Добавление новых сервисов

## Правила именования интерфейсов

Унифицировали именование интерфейсов с префиксом `I`:

```go
// Services
type IAuthService interface { ... }

// Messaging
type IMessageBroker interface { ... }
type IRabbitMQConnector interface { ... }
type IRabbitMQPublisher interface { ... }

// Repositories
type IUserRepository interface { ... }
type IDatabase interface { ... }
```

## Миграция

### Изменения в коде

1. **AuthService** теперь использует `messaging.IMessageBroker`
2. **RabbitMQAdapter** перемещен в `messaging/`
3. **Интерфейсы** разделены по доменам
4. **Тесты** обновлены для новой структуры

### Обновление импортов

```go
// Было
import "github.com/Koshsky/subs-service/auth-service/internal/services"

// Стало
import (
    "github.com/Koshsky/subs-service/auth-service/internal/services"
    "github.com/Koshsky/subs-service/auth-service/internal/messaging"
)
```

## Генерация моков

Теперь моки генерируются для каждого слоя отдельно:

```bash
# Моки для services
go generate ./internal/services/

# Моки для messaging
go generate ./internal/messaging/

# Моки для repositories
go generate ./internal/repositories/
```

## Заключение

Новая архитектура обеспечивает:
- ✅ Четкое разделение ответственностей
- ✅ Следование принципам Clean Architecture
- ✅ Улучшенную тестируемость
- ✅ Легкость расширения и поддержки
- ✅ Единообразное именование

Это создает прочную основу для дальнейшего развития приложения.
