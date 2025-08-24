# Auth Service

Сервис аутентификации на Go с использованием gRPC, PostgreSQL и RabbitMQ для микросервисной архитектуры.

## 🚀 Возможности

- **Регистрация пользователей** - создание новых аккаунтов с валидацией данных
- **Аутентификация пользователей** - вход в систему с JWT токенами
- **Валидация токенов** - проверка JWT токенов и получение информации о пользователе
- **События через RabbitMQ** - публикация событий для других сервисов
- **TLS поддержка** - безопасное соединение с SSL/TLS сертификатами
- **Graceful degradation** - продолжение работы при недоступности RabbitMQ

## 🏗️ Архитектура

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   gRPC Client   │    │   Auth Service  │    │   PostgreSQL    │
│                 │◄──►│                 │◄──►│   Database      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │    RabbitMQ     │
                       │   Event Bus     │
                       └─────────────────┘
```

## 📋 Требования

- Go 1.24+
- PostgreSQL 12+
- RabbitMQ 3.8+
- Docker (опционально)

## 🛠️ Установка и запуск

### 1. Клонирование репозитория

```bash
git clone <repository-url>
cd auth-service
```

### 2. Настройка переменных окружения

Создайте файл `.env` в корне проекта:

```env
# Database Configuration
AUTH_DB_HOST=auth-db
AUTH_DB_PORT=5432
AUTH_DB_USER=postgres
AUTH_DB_PASSWORD=your_password
AUTH_DB_NAME=auth_service
AUTH_DB_SSLMODE=disable

# RabbitMQ Configuration
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
RABBITMQ_EXCHANGE=user_events

# JWT Configuration
JWT_SECRET=your_super_secret_jwt_key_at_least_32_characters_long

# Service Configuration
AUTH_SERVICE_PORT=50051

# TLS Configuration (опционально)
ENABLE_TLS=false
TLS_CERT_FILE=certs/server-cert.pem
TLS_KEY_FILE=certs/server-key.pem
```

### 3. Запуск с Docker Compose

```bash
docker-compose up -d
```

### 4. Запуск локально

```bash
# Установка зависимостей
go mod download

# Запуск миграций
# (убедитесь, что PostgreSQL запущен и настроен)

# Запуск сервиса
go run cmd/auth-service/main.go
```

## 📊 API Endpoints

Сервис предоставляет gRPC API со следующими методами:

### Register
Регистрация нового пользователя

```protobuf
rpc Register(RegisterRequest) returns (RegisterResponse)
```

**Request:**
```json
{
  "email": "user@example.com",
  "password": "secure_password"
}
```

**Response:**
```json
{
  "user_id": "uuid",
  "email": "user@example.com",
  "success": true,
  "error": "",
  "message": "User registered successfully"
}
```

### Login
Аутентификация пользователя

```protobuf
rpc Login(LoginRequest) returns (LoginResponse)
```

**Request:**
```json
{
  "email": "user@example.com",
  "password": "secure_password"
}
```

**Response:**
```json
{
  "token": "jwt_token",
  "user_id": "uuid",
  "email": "user@example.com",
  "success": true,
  "error": "",
  "message": "Login successful"
}
```

### ValidateToken
Валидация JWT токена

```protobuf
rpc ValidateToken(TokenRequest) returns (UserResponse)
```

**Request:**
```json
{
  "token": "jwt_token"
}
```

**Response:**
```json
{
  "user_id": "uuid",
  "email": "user@example.com",
  "valid": true,
  "error": ""
}
```

## 🗄️ База данных

### Схема таблицы users

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);
```

### Миграции

Миграции находятся в папке `migrations/` и используют формат SQL с up/down файлами.

## 🔧 Конфигурация

### Переменные окружения

| Переменная | Описание | Обязательная | По умолчанию |
|------------|----------|--------------|--------------|
| `AUTH_DB_HOST` | Хост PostgreSQL | Нет | `auth-db` |
| `AUTH_DB_PORT` | Порт PostgreSQL | Да | - |
| `AUTH_DB_USER` | Пользователь БД | Да | - |
| `AUTH_DB_PASSWORD` | Пароль БД | Да | - |
| `AUTH_DB_NAME` | Имя БД | Да | - |
| `AUTH_DB_SSLMODE` | SSL режим | Нет | `disable` |
| `RABBITMQ_URL` | URL RabbitMQ | Нет | `amqp://guest:guest@rabbitmq:5672/` |
| `RABBITMQ_EXCHANGE` | Exchange для событий | Нет | `user_events` |
| `JWT_SECRET` | Секрет для JWT | Да | - |
| `AUTH_SERVICE_PORT` | Порт сервиса | Да | - |
| `ENABLE_TLS` | Включить TLS | Нет | `false` |
| `TLS_CERT_FILE` | Путь к сертификату | Нет | `certs/server-cert.pem` |
| `TLS_KEY_FILE` | Путь к ключу | Нет | `certs/server-key.pem` |

## 🧪 Тестирование

### Запуск тестов

```bash
# Все тесты
go test ./...

# Тесты с покрытием
go test -cover ./...

# Тесты конкретного пакета
go test ./internal/services/
```

### Моки

Проект использует `testify/mock` для создания моков. Моки генерируются автоматически и находятся в папках `mocks/`.

## 🚀 CI/CD Pipeline

Проект использует GitHub Actions для автоматизации процессов разработки:

### Основной Pipeline (`.github/workflows/ci.yml`)

- **Тестирование** - запуск тестов на Go 1.24 и 1.25
- **Линтинг** - проверка кода с помощью golangci-lint
- **Безопасность** - сканирование уязвимостей с gosec
- **Сборка** - компиляция для Linux, macOS и Windows
- **Docker** - сборка и публикация Docker образов
- **Релизы** - автоматическое создание релизов при тегах

### Управление зависимостями (`.github/workflows/dependencies.yml`)

- **Проверка обновлений** - еженедельная проверка устаревших зависимостей
- **Автоматическое обновление** - создание PR с обновлениями

### Запуск pipeline

Pipeline автоматически запускается при:
- Push в ветки `main`, `master`, `develop`
- Pull Request в `main` или `master`
- Создание тегов (для релизов)

### Настройка секретов

Для полной функциональности настройте секреты в GitHub:

```bash
DOCKER_USERNAME=your_docker_username
DOCKER_PASSWORD=your_docker_password
```

## 📦 Docker

### Сборка образа

```bash
docker build -t auth-service .
```

### Запуск контейнера

```bash
docker run -p 50051:50051 --env-file .env auth-service
```

## 🔒 Безопасность

- Пароли хешируются с использованием bcrypt
- JWT токены с настраиваемым секретом
- Поддержка TLS для безопасных соединений
- Валидация входных данных
- Graceful degradation при недоступности зависимостей

## 📈 Мониторинг

Сервис включает health check endpoint для мониторинга:

```bash
grpc_health_probe -addr=localhost:50051
```

## 🚨 Логирование

Сервис использует стандартный Go logger с различными уровнями логирования:
- Info: информация о запуске и основных операциях
- Warning: предупреждения (например, недоступность RabbitMQ)
- Error: ошибки, требующие внимания

## 🤝 События

Сервис публикует события в RabbitMQ при следующих действиях:
- Регистрация нового пользователя
- Успешный вход пользователя

## 📁 Структура проекта

```
auth-service/
├── cmd/
│   └── auth-service/
│       ├── main.go                 # Точка входа приложения
│       └── main_unit_test.go       # Тесты main.go
├── internal/
│   ├── authpb/                     # Protobuf определения
│   ├── config/                     # Конфигурация
│   ├── messaging/                  # RabbitMQ адаптер
│   ├── models/                     # Модели данных
│   ├── repositories/               # Слой доступа к данным
│   ├── server/                     # gRPC сервер
│   ├── services/                   # Бизнес-логика
│   └── utils/                      # Утилиты
├── migrations/                     # Миграции БД
├── docs/                          # Документация
├── Dockerfile                     # Docker образ
├── go.mod                         # Go модули
└── README.md                      # Этот файл
```

## 🐛 Устранение неполадок

### Проблемы с подключением к БД
- Проверьте переменные окружения для БД
- Убедитесь, что PostgreSQL запущен и доступен
- Проверьте SSL настройки

### Проблемы с RabbitMQ
- Сервис продолжит работу без RabbitMQ (graceful degradation)
- Проверьте URL и настройки подключения
- Убедитесь, что exchange существует

### Проблемы с JWT
- Проверьте, что JWT_SECRET имеет минимум 32 символа
- Убедитесь, что токены не истекли

## 📄 Лицензия

MIT License

Copyright (c) 2024 Матвей Шмонов

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

## 👥 Авторы

**Матвей Шмонов** - разработчик и архитектор проекта

## 🤝 Вклад в проект

1. Fork репозитория
2. Создайте feature branch (`git checkout -b feature/amazing-feature`)
3. Commit изменения (`git commit -m 'Add amazing feature'`)
4. Push в branch (`git push origin feature/amazing-feature`)
5. Откройте Pull Request
