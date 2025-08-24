# Graceful Degradation Pattern

## Обзор

Graceful Degradation (Плавная деградация) - это архитектурный паттерн, при котором система продолжает работать даже при частичных сбоях или отсутствии некоторых зависимостей, возвращая понятные ошибки вместо паники.

## Принцип работы

Вместо проверки зависимостей в конструкторах (Fail Fast), мы проверяем их в каждом методе и возвращаем осмысленные ошибки:

```go
// ❌ Fail Fast (в конструкторе)
func NewUserRepository(db *gorm.DB) (*UserRepository, error) {
    if db == nil {
        return nil, errors.New("database cannot be nil")
    }
    return &UserRepository{DB: db}, nil
}

// ✅ Graceful Degradation (в методах)
func (ur *UserRepository) CreateUser(user *models.User) error {
    if ur.DB == nil {
        return errors.New("database connection is not initialized")
    }
    // ... бизнес-логика
}
```

## Реализация в проекте

### 1. UserRepository
```go
func (ur *UserRepository) CreateUser(user *models.User) error {
    if ur.DB == nil {
        return errors.New("database connection is not initialized")
    }
    // ... логика создания пользователя
}

func (ur *UserRepository) ValidateUser(email, password string) (*models.User, error) {
    if ur.DB == nil {
        return nil, errors.New("database connection is not initialized")
    }
    // ... логика валидации
}

func (ur *UserRepository) GetUserByEmail(email string) (*models.User, error) {
    if ur.DB == nil {
        return nil, errors.New("database connection is not initialized")
    }
    // ... логика поиска
}
```

### 2. AuthService
```go
func (s *AuthService) Register(ctx context.Context, email, password string) (*models.User, error) {
    if s.userRepo == nil {
        return nil, errors.New("user repository is not initialized")
    }
    // ... логика регистрации
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
    if s.userRepo == nil {
        return "", errors.New("user repository is not initialized")
    }
    // ... логика входа
}
```

### 3. RabbitMQAdapter
```go
func (r *RabbitMQAdapter) PublishUserCreated(user *models.User) error {
    if r.publisher == nil {
        return errors.New("publisher is not initialized")
    }
    if r.config == nil {
        return errors.New("config is not initialized")
    }
    // ... логика публикации
}

func (r *RabbitMQAdapter) PublishUserDeleted(user *models.User) error {
    if r.publisher == nil {
        return errors.New("publisher is not initialized")
    }
    if r.config == nil {
        return errors.New("config is not initialized")
    }
    // ... логика публикации
}
```

### 4. GormAdapter
```go
func (g *GormAdapter) Create(value interface{}) IDatabase {
    if g.db == nil {
        return &GormAdapter{db: nil}
    }
    return &GormAdapter{db: g.db.Create(value)}
}

func (g *GormAdapter) GetError() error {
    if g.db == nil {
        return errors.New("database is nil")
    }
    return g.db.Error
}
```

## Преимущества

### 1. **Гибкость тестирования**
```go
// Можно создавать объекты с nil зависимостями для тестирования edge cases
repo := &UserRepository{DB: nil}
err := repo.CreateUser(user)
// Получаем осмысленную ошибку вместо паники
```

### 2. **Меньше дублирования кода**
- Одна проверка в методе вместо двух (в конструкторе + в методе)
- Более чистый код

### 3. **Лучшая изоляция ошибок**
- Ошибки локализованы в конкретных методах
- Система не падает полностью при проблемах с одной зависимостью

### 4. **Graceful handling**
- Приложение не паникует при неожиданных nil зависимостях
- Пользователь получает понятные сообщения об ошибках

### 5. **Упрощенные конструкторы**
```go
// Простой конструктор без сложной валидации
func NewUserRepository(db *gorm.DB) *UserRepository {
    database := NewGormAdapter(db)
    return &UserRepository{DB: database}
}
```

## Недостатки

### 1. **Потенциальные нерабочие объекты**
- Объект может существовать в нерабочем состоянии
- Ошибки обнаруживаются только при вызове методов

### 2. **Дополнительные проверки в runtime**
- Каждый вызов метода требует проверки зависимостей
- Небольшое снижение производительности

### 3. **Менее явные контракты**
- Не сразу понятно, какие зависимости требуются
- Нужно читать код методов для понимания требований

## Когда использовать

### ✅ Подходит для:
- **Тестирования** - легко создавать моки и тестировать edge cases
- **Гибких систем** - где зависимости могут быть опциональными
- **Микросервисов** - где важна устойчивость к частичным сбоям
- **Прототипирования** - быстрая разработка без строгих контрактов

### ❌ Не подходит для:
- **Критически важных систем** - где нужна гарантия работоспособности
- **Строгих контрактов** - где важно явно определить зависимости
- **Высокопроизводительных систем** - где важна каждая миллисекунда

## Альтернативы

### 1. **Fail Fast (в конструкторах)**
```go
func NewUserRepository(db *gorm.DB) (*UserRepository, error) {
    if db == nil {
        return nil, errors.New("database cannot be nil")
    }
    return &UserRepository{DB: db}, nil
}
```

### 2. **Dependency Injection Container**
```go
container := NewContainer()
container.Register("database", db)
repo := container.Resolve("userRepository").(*UserRepository)
```

### 3. **Builder Pattern**
```go
repo := NewUserRepositoryBuilder().
    WithDatabase(db).
    WithCache(cache).
    Build()
```

## Заключение

Graceful Degradation - это компромисс между простотой разработки и строгостью архитектуры. В нашем проекте этот подход обеспечивает:

1. **Простое тестирование** - легко создавать объекты с nil зависимостями
2. **Устойчивость к ошибкам** - система не паникует при проблемах
3. **Читаемый код** - понятные сообщения об ошибках
4. **Гибкость** - можно создавать объекты разными способами

Этот паттерн особенно хорошо подходит для микросервисной архитектуры, где важна устойчивость и простота тестирования.