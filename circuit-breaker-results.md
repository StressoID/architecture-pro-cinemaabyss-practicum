# Результаты работы Circuit Breaker

## Применение конфигурации Circuit Breaker

### Команда
```bash
kubectl apply -f ./src/kubernetes/circuit-breaker-config.yaml -n cinemaabyss
```

### Результат
[Здесь опишите результат применения конфигурации: созданные DestinationRule, статус применения]

## Тестирование с помощью Fortio

### Команда
```bash
kubectl exec -n cinemaabyss $FORTIO_POD -c fortio -- fortio load -c 50 -qps 0 -n 500 -loglevel Warning http://movies-service:8081/api/movies
```

### Результат
[Здесь опишите результаты нагрузочного тестирования:
- Распределение IP адресов
- Коды ответов (200, 500, 503) и их процентное соотношение
- Количество запросов, обработанных через circuit breaker]

### Пример вывода
```
IP addresses distribution:
10.106.113.46:8081: 421
Code 200 : 79 (15.8 %)
Code 500 : 22 (4.4 %)
Code 503 : 399 (79.8 %)
```

## Статистика Circuit Breaker

### Команда
```bash
kubectl exec -n cinemaabyss $FORTIO_POD -c istio-proxy -- pilot-agent request GET stats | grep movies-service | grep pending
```

### Результат
[Здесь опишите статистику:
- Значение `upstream_rq_pending_total` - количество раз срабатывания circuit breaker
- Значение `upstream_rq_pending_overflow` - количество вызовов, заблокированных circuit breaker]

### Пример вывода
```
cluster.outbound|8081||movies-service.cinemaabyss.svc.cluster.local;.upstream_rq_pending_total: 311
cluster.outbound|8081||movies-service.cinemaabyss.svc.cluster.local;.upstream_rq_pending_overflow: 21
```

## Анализ работы Circuit Breaker

[Здесь опишите анализ работы circuit breaker:
- При каком количестве ошибок срабатывает circuit breaker
- Как долго сервис остается в состоянии "открыт" (open state)
- Процент запросов, которые были заблокированы
- Восстановление сервиса после срабатывания circuit breaker]
