## User Service based on -> Golang (Gin framework) \ Postgresql (goose migrator) \ Redis \ Docker \ Makefile \ Unit Tests

- [x] ⚡ Высокопроизводительный фреймворк - Gin
- [x] ⚡ Высокопроизводительный логгер - Uber Zap Logger
####
- [x] Аутентификая с помощью JWT
- [x] Роли юзеров (user, moderator, admin)
- [x] Middleware для каждой роли
- [X] Шифрование паролей, с использованием secret key
- [X] Сохранение всех сессий аккаунта юзера, с дальнейшим получением информаций всех сессий (user-agent, ip, refresh tokens, fingerprint, datetime)
- [X] Бесконечная сессия, до тех пор, пока человек не будет в афк месяц
- [X] Использование Postgres (goose) \ Redis \ Docker \ Makefile
- [x] Высокая защита аккаунта от кражи access / refresh токенов (при авторизации сверяется fingerprint устрйоства у которого украли токены)
- [x] Сохранение логов


