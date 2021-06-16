# REST API для для обмена аудиофайлами формата AAC

REST API был написан с целью практики в программировании на языке GO ([задание](TASK.md))

В проекте были использованы следующие концепции и технологии:
- HTTP фреймворк Gin <a href="https://github.com/gin-gonic/gin">gin-gonic/gin</a>
- БД Postgresql и библиотека для работы с ней <a href="https://github.com/jmoiron/sqlx">sqlx</a>
- Инициализация БД с помощью <a href="https://github.com/golang-migrate/migrate">migrate</a>
- Аутентификация  с помощью JWT + Refresh Token 
- Конфигурация приложения с помощь библиотеки <a href="https://github.com/spf13/viper">spf13/viper</a>
- Юнит тестирование (<a href="github.com/golang/mock/gomock">gomock</a>, <a href="https://github.com/DATA-DOG/go-sqlmock">go-sqlmock</a>)
- Запуск проекта с помощью Docker (docker-compose)
- Описание API с помощью swagger (<a href="https://github.com/swaggo/swag">swag</a>)


### Запуск
```
make build && make run
```

### Инициализация БД
```
make migrate_up
```


# TODO
 - Логирование
 - ~~Refresh Token~~
