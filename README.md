# devmetrics

Учебное клиент-серверное приложение сбора метрик для трека «Go в DevOps».

## Агент
Собирает метрики целевой системы и передает их на сервер хранения.

### Основной функционал
* Отправляет метрики в JSON формате на HTTP endpoint, поддерживает передачу данных по протоколу gRPC
* Настраиваемый интервал сбора метрик
* Настраиваемый интервал отправки метрик
* При указании ключа подписи вычисляет хеш и подписывает передаваемые метрики
* При указании публичного ключа асинхронно шифрует пакеты метрик

## Сервер
Принимает и сохраняет метрики в JSON формате или по протоколу gRPC. 
Обеспечивает интерфейс доступа к метрикам (web / api).

### Основной функционал
* Обеспечивает постоянное хранение метрик. 
Поддерживается 2 типа Storage через универсальный интерфейс: In-Mem, PostgreSQL
* Принимает метрики по протоколу http в формате JSON, поддерживает работу по протоколу gRPC
* Поддерживает обработку данных с gzip сжатием
* Поддерживает работу с асинхронным шифрованием пакетов метрик
* Поддерживает хеш-подпись метрик
* Поддерживает ограничение входящих запросов по маске подсети


## Дополнительная информация
* Доступна Godoc документация
* Для серверной части описан Swagger
* В комплекте сконфигурированный линтер /cmd/staticlint/
