# Балансировщик нагрузки с ограничением запросов

## Описание проекта

Высокопроизводительный балансировщик нагрузки с:
- Алгоритмами Round Robin и Least Connections
- Ограничением скорости запросов (Rate Limiting)
- Проверкой здоровья бэкендов
- Поддержкой IPv4 и IPv6
- Детальным логированием

## Быстрый старт

### Требования
- Go 1.20+
- SQLite
 ## Установка
- 1. Клонируйте репозиторий:
_git clone https://github.com/OlegrusWR/balancer_to_cloud.git_
- 2. Переите в нужную директорию   
    _cd test_to_cloud/balancer/cmd_
- 3. Запустите программу   
_go run main.go_
- 4. Для остановки сервера нажмите `Ctrl+C` в терминале

## Конфигурация
Перед запуском отредактируйте файл `config.yaml` в корне проекта:
```yaml
listen_port: 8080
algoritm: "least_conn" # или "round_robin"
backends:
  - "http://localhost:8081"
  - "http://localhost:8082"
  - "http://localhost:8083"
health_check:
  interval: "10s"
  timeout: "2s"
  path: "/health"
rate_limiter:
  default_capacity: 100          
  default_rate: "2s"              
  vip_capacity: 300               
  vip_rate: "500ms"               
  db_path: "../data/ratelimit.db"    

```

## Тестовые серверы
Для тестирования рекомендуется запустить тестовые серверы (в отдельных терминалах):
переходим в с разных терминалов директории back1, back2, back3 и запускаем тестовые сервера
```bash
go run backend1
go run backend2
go run backend3
```
соответственно директориям

## Логирование

Логи сохраняются в директории ./log/:

*loadbalancer.log - Основные события*

*ratelimit.log - Ограничение запросов* (с этим проблема)

*database.log - работа с БД* (и с этим тожепроблема)


## Docker and Docker-compose

Необходимо запускать контейнер в корневой папке проекта (test_to_cloud) 

Docker-compose не успел нормально дописать, простите :(
поэтому даже пушить не стал