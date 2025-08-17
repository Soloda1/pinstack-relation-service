# Pinstack Relation Service 💕 Pinstack Relation Service �

**Pinstack Relation Service** — микросервис для управления связями между пользователями в системе **Pinstack**.

## Основные функции:
- CRUD-операции для связей между пользователями (подписки, отписки).
- Получение списка подписчиков и подписок пользователя.
- Взаимодействие с другими микросервисами через gRPC.
- Асинхронная обработка событий через Kafka.

## Технологии:
- **Go** — основной язык разработки.
- **gRPC** — для межсервисной коммуникации.
- **Kafka** — асинхронная обработка событий и интеграция с другими сервисами.
- **Docker** — для контейнеризации.
- **Prometheus** — для сбора метрик и мониторинга.
- **Grafana** — для визуализации метрик.
- **Loki** — для централизованного сбора логов.
- **Redis** — для кэширования.

## Архитектура

Проект построен на основе **гексагональной архитектуры (Hexagonal Architecture)** с четким разделением слоев:

### Структура проекта
```
├── cmd/                    # Точки входа приложения
│   ├── server/             # gRPC сервер
│   └── migrate/            # Миграции БД
├── internal/
│   ├── domain/             # Доменный слой
│   │   ├── models/         # Доменные модели
│   │   └── ports/          # Интерфейсы (порты)
│   │       ├── input/      # Входящие порты (use cases)
│   │       └── output/     # Исходящие порты (репозитории, Kafka, метрики)
│   ├── application/        # Слой приложения
│   │   └── service/        # Бизнес-логика и сервисы
│   └── infrastructure/     # Инфраструктурный слой
│       ├── inbound/        # Входящие адаптеры (gRPC)
│       └── outbound/       # Исходящие адаптеры (PostgreSQL, Kafka, Redis)
├── migrations/             # SQL миграции
└── mocks/                 # Моки для тестирования
```

### Принципы архитектуры
- **Dependency Inversion**: Зависимости направлены к доменному слою
- **Clean Architecture**: Четкое разделение ответственности между слоями
- **Port & Adapter Pattern**: Интерфейсы определяются в domain, реализуются в infrastructure
- **Testability**: Легкое модульное тестирование благодаря dependency injection
- **Event-Driven Architecture**: Асинхронная обработка событий через Kafka

### Мониторинг и метрики
Сервис включает полную интеграцию с системой мониторинга:
- **Prometheus метрики**: Автоматический сбор метрик gRPC, базы данных, Kafka
- **Structured logging**: Интеграция с Loki для централизованного сбора логов
- **Health checks**: Проверки состояния всех компонентов
- **Performance monitoring**: Метрики времени ответа и throughput

## CI/CD Pipeline 🚀

### GitHub Actions
Проект использует GitHub Actions для автоматического тестирования при каждом push/PR.

**Этапы CI:**
1. **Unit Tests** — юнит-тесты с покрытием кода
2. **Integration Tests** — интеграционные тесты с полной инфраструктурой 
3. **Auto Cleanup** — автоматическая очистка Docker ресурсов

### Makefile команды 📋

#### Команды разработки

### Настройка и запуск
```bash
# Создание необходимых сетей
docker network create pinstack
docker network create pinstack-test

# Запуск легкой среды разработки (только Prometheus stack)
make start-dev-light

# Запуск полной среды разработки (с мониторингом и ELK)
make start-dev-full

# Остановка среды разработки
make stop-dev-full
```

### Мониторинг
```bash
# Запуск полного стека мониторинга (Prometheus, Grafana, Loki, ELK)
make start-monitoring

# Запуск только Prometheus stack
make start-prometheus-stack

# Запуск только ELK stack
make start-elk-stack

# Остановка мониторинга
make stop-monitoring

# Проверка состояния мониторинга
make check-monitoring-health

# Просмотр логов мониторинга
make logs-prometheus
make logs-grafana
make logs-loki
make logs-elasticsearch
make logs-kibana
```

### Доступ к сервисам мониторинга
- **Prometheus**: http://localhost:9090 - метрики и мониторинг
- **Grafana**: http://localhost:3000 (admin/admin) - дашборды и визуализация
- **Loki**: http://localhost:3100 - централизованные логи
- **Kibana**: http://localhost:5601 - анализ логов ELK
- **Elasticsearch**: http://localhost:9200 - поисковая система
- **PgAdmin**: http://localhost:5050 (admin@admin.com/admin) - управление БД
- **Kafka UI**: http://localhost:9091 - управление Kafka

### Redis и кэширование
```bash
# Утилиты Redis для отладки
make redis-cli              # Подключение к Redis CLI
make redis-info             # Информация о Redis
make redis-keys             # Показать все ключи
make redis-flush            # Очистить все данные Redis (с подтверждением)

# Просмотр логов Redis
make logs-redis
```

#### Управление инфраструктурой:
```bash
# Настройка репозиториев
make setup-system-tests        # Клонирует/обновляет pinstack-system-tests репозиторий
make setup-monitoring          # Клонирует/обновляет pinstack-monitoring-service репозиторий

# Запуск инфраструктуры
make start-relation-infrastructure  # Поднимает все Docker контейнеры для тестов
make check-services                 # Проверяет готовность всех сервисов

# Интеграционные тесты
make test-relation-integration      # Запускает только интеграционные тесты
make quick-test                     # Быстрый запуск тестов без пересборки контейнеров
make quick-test-local               # Быстрый тест с локальным relation-service

# Остановка и очистка
make stop-relation-infrastructure   # Останавливает все тестовые контейнеры
make clean-relation-infrastructure  # Полная очистка (контейнеры + volumes + образы)
make clean                         # Полная очистка проекта + Docker
```

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

#### Логи и отладка:
```bash
# Просмотр логов сервисов
make logs-relation          # Логи Relation Service
make logs-notification      # Логи Notification Service
make logs-user              # Логи User Service
make logs-auth              # Логи Auth Service  
make logs-gateway           # Логи API Gateway
make logs-kafka             # Логи Kafka
make logs-redis             # Логи Redis

# Просмотр логов баз данных
make logs-relation-db       # Логи Relation Database
make logs-notification-db   # Логи Notification Database
make logs-db                # Логи User Database
make logs-auth-db           # Логи Auth Database

# Экстренная очистка
make clean-docker-force     # Удаляет ВСЕ Docker ресурсы (с подтверждением)
```

### Зависимости для интеграционных тестов 🐳

Для интеграционных тестов автоматически поднимаются контейнеры:
- **user-db-test** — PostgreSQL для User Service
- **user-migrator-test** — миграции User Service  
- **user-service-test** — сам User Service
- **auth-db-test** — PostgreSQL для Auth Service
- **auth-migrator-test** — миграции Auth Service
- **auth-service-test** — Auth Service
- **relation-db-test** — PostgreSQL для Relation Service
- **relation-migrator-test** — миграции Relation Service
- **relation-service-test** — сам Relation Service
- **notification-db-test** — PostgreSQL для Notification Service
- **notification-migrator-test** — миграции Notification Service
- **notification-service-test** — Notification Service
- **kafka-test** — Kafka для асинхронной обработки
- **kafka-topics-init-test** — инициализация топиков Kafka
- **redis** — Redis для кэширования
- **api-gateway-test** — API Gateway

> 📍 **Требования:** Docker, docker-compose  
> 🚀 **Все сервисы собираются автоматически из Git репозиториев**  
> 🔄 **Репозитории `pinstack-system-tests` и `pinstack-monitoring-service` клонируются автоматически при запуске**

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

# 5. Запуск полной среды разработки с мониторингом
make start-dev-full

# 6. Очистка после работы
make clean
```

### Особенности 🔧

- **Event-Driven Architecture:** Сервис интегрирован с Kafka для асинхронной обработки событий
- **Отключение кеша тестов:** все тесты запускаются с флагом `-count=1`
- **Фокус на Relation Service:** интеграционные тесты тестируют только Relation endpoints
- **Автоочистка:** CI автоматически удаляет все Docker ресурсы после себя
- **Параллельность:** в CI юнит и интеграционные тесты запускаются последовательно
- **Полный мониторинг:** Включены Prometheus, Grafana, Loki и ELK stack
- **Redis кэширование:** Встроенная поддержка Redis для улучшения производительности

> ✅ Сервис готов к использованию.


