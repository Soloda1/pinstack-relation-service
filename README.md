# Pinstack Relation Service 🔗

**Pinstack Relation Service** — микросервис для управления связями между пользователями в системе **Pinstack**.

## Основные функции:
- CRUD-операции для связей между пользователями (подписки, отписки).
- Получение списка подписчиков и подписок пользователя.
- Взаимодействие с другими микросервисами через gRPC.

## Технологии:
- **Go** — основной язык разработки.
- **gRPC** — для межсервисной коммуникации.
- **Kafka** — асинхронная обработка событий и интеграция с другими сервисами.
- **Docker** — для контейнеризации.

## CI/CD Pipeline 🚀

### GitHub Actions
Проект использует GitHub Actions для автоматического тестирования при каждом push/PR.

**Этапы CI:**
1. **Unit Tests** — юнит-тесты с покрытием кода
2. **Integration Tests** — интеграционные тесты с полной инфраструктурой 
3. **Auto Cleanup** — автоматическая очистка Docker ресурсов

### Makefile команды 📋

#### Основные команды разработки:
```bash
# Проверка кода и тесты
make fmt                    # Форматирование кода (gofmt)
make lint                   # Проверка кода (go vet)
make test-unit              # Юнит-тесты с покрытием
make test-integration       # Интеграционные тесты (с Docker)
make test-all               # Все тесты: форматирование + линтер + юнит + интеграционные

# CI локально
make ci-local               # Полный CI процесс локально (имитация GitHub Actions)
```

#### Управление инфраструктурой:
```bash
# Настройка репозитория
make setup-system-tests           # Клонирует/обновляет pinstack-system-tests репозиторий

# Запуск инфраструктуры
make start-relation-infrastructure # Поднимает все Docker контейнеры для тестов
make check-services               # Проверяет готовность всех сервисов

# Интеграционные тесты
make test-relation-integration    # Запускает только интеграционные тесты
make quick-test                  # Быстрый запуск тестов без пересборки контейнеров

# Остановка и очистка
make stop-relation-infrastructure  # Останавливает все тестовые контейнеры
make clean-relation-infrastructure # Полная очистка (контейнеры + volumes + образы)
make clean                        # Полная очистка проекта + Docker
```

#### Логи и отладка:
```bash
# Просмотр логов сервисов
make logs-relation          # Логи Relation Service
make logs-notification      # Логи Notification Service
make logs-kafka             # Логи Kafka
make logs-user              # Логи User Service
make logs-auth              # Логи Auth Service  
make logs-gateway           # Логи API Gateway
make logs-relation-db       # Логи Relation Database
make logs-notification-db   # Логи Notification Database
make logs-db                # Логи User Database
make logs-auth-db           # Логи Auth Database

# Экстренная очистка
make clean-docker-force     # Удаляет ВСЕ Docker ресурсы (с подтверждением)
```

### Зависимости для интеграционных тестов 🐳

Для интеграционных тестов автоматически поднимаются контейнеры:
- **relation-db-test** — PostgreSQL для Relation Service
- **relation-migrator-test** — миграции Relation Service  
- **relation-service-test** — сам Relation Service
- **notification-db-test** — PostgreSQL для Notification Service
- **notification-migrator-test** — миграции Notification Service
- **notification-service-test** — Notification Service
- **kafka-test** — Kafka для асинхронных событий
- **kafka-topics-init-test** — инициализация Kafka топиков
- **user-db-test** — PostgreSQL для User Service
- **user-migrator-test** — миграции User Service  
- **user-service-test** — User Service
- **auth-db-test** — PostgreSQL для Auth Service
- **auth-migrator-test** — миграции Auth Service
- **auth-service-test** — Auth Service
- **api-gateway-test** — API Gateway

> 📍 **Требования:** Docker, docker-compose  
> 🚀 **Все сервисы собираются автоматически из Git репозиториев**  
> 🔄 **Репозиторий `pinstack-system-tests` клонируется автоматически при запуске тестов**

### Быстрый старт разработки ⚡

```bash
# 1. Проверить код
make fmt lint

# 2. Запустить юнит-тесты
make test-unit

# 3. Запустить интеграционные тесты
make test-integration

# 4. Или всё сразу
make ci-local

# 5. Очистка после работы
make clean
```

### Особенности 🔧

- **Отключение кеша тестов:** все тесты запускаются с флагом `-count=1`
- **Фокус на Relation Service:** интеграционные тесты тестируют только Relation endpoints
- **Автоочистка:** CI автоматически удаляет все Docker ресурсы после себя
- **Параллельность:** в CI юнит и интеграционные тесты запускаются последовательно
- **Kafka интеграция:** поддержка асинхронных событий между сервисами

> ✅ Сервис готов к использованию.
