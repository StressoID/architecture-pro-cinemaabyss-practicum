# Результаты развертывания через Helm

## Развертывание

### Команда
```bash
helm install cinemaabyss ./src/kubernetes/helm --namespace cinemaabyss --create-namespace
```

### Результат
[Здесь опишите результат развертывания: статус установки, список развернутых ресурсов, возможные ошибки]

## Проверка подов

### Команда
```bash
kubectl get pods -n cinemaabyss
```

### Результат
[Здесь опишите статус всех подов: какие поды запущены, их статусы (Running/Pending/Error), количество реплик]

## Вызов API /api/movies

### Команда
```bash
curl https://cinemaabyss.example.com/api/movies
```

### Результат
[Здесь опишите результат выполнения команды: статус ответа, тело ответа, время выполнения]

## Проверка работы Proxy Service

### Команда
```bash
kubectl -n cinemaabyss logs -l app=proxy-service --tail=50
```

### Результат
[Здесь опишите логи proxy-service: маршрутизация запросов, распределение трафика между монолитом и movies-service]
