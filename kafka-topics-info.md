# Информация о топиках Kafka

## Список топиков

```
__consumer_offsets
movie-events
payment-events
user-events
```

## Детальная информация о топиках

### movie-events

**Конфигурация:**
- PartitionCount: 1
- ReplicationFactor: 1
- Configs: segment.bytes=1073741824
- Partition: 0
- Leader: 1001
- Replicas: 1001
- Isr: 1001
- **Количество сообщений:** 3 (offset: 0-2)

**Примеры сообщений:**
```json
{"movie_id":6,"title":"Test Movie Event","action":"viewed","user_id":4,"timestamp":"2025-12-11T19:41:43.712000261Z"}
{"movie_id":1,"title":"","action":"","timestamp":"2025-12-11T19:42:30.312919297Z"}
{"movie_id":7,"title":"Test Movie Event","action":"viewed","user_id":5,"timestamp":"2025-12-11T19:43:25.898559961Z"}
```

### user-events

**Топик создан и используется Events Service для публикации событий пользователей.**
- **Количество сообщений:** 2 (offset: 0-1)

### payment-events

**Топик создан и используется Events Service для публикации событий платежей.**
- **Количество сообщений:** 2 (offset: 0-1)

## Kafka UI

Kafka UI доступен по адресу: **http://localhost:8090**

В интерфейсе можно увидеть:
- Список всех топиков
- Количество сообщений в каждом топике
- Consumer groups
- Детальную информацию о партициях
- Просмотр сообщений в реальном времени

## Логи Events Service

Events Service успешно:
- ✅ Публикует события в топики Kafka
- ✅ Читает события из топиков через Consumer
- ✅ Логирует все обработанные события

Пример логов:
```
2025/12/11 19:41:43 Published event to topic movie-events, partition 0, offset 0
2025/12/11 19:41:43 Received event from topic movie-events, partition 0, offset 0
2025/12/11 19:41:43 Published event to topic user-events, partition 0, offset 0
2025/12/11 19:41:43 Received event from topic user-events, partition 0, offset 0
2025/12/11 19:41:44 Published event to topic payment-events, partition 0, offset 0
2025/12/11 19:41:44 Received event from topic payment-events, partition 0, offset 0
```
