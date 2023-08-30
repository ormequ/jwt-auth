# JWT Authentication

Реализация сервиса аутентификации при помощи JWT (Access + Refresh) с хранением в MongoDB

---

Запуск: `docker-compose up`

---

Сервис имеет два эндпоинта:
- /api/generate - генерация пары Access + Refresh по заданному в user_id в теле запроса. 
 user_id должен быть в формате UUID с дефисами, например, A0E7DFB1-E5A0-4D59-8DEB-B2B6FEDDE95E
- /api/refresh - обновление пары Access + Refresh. В теле запроса должны быть переданы оба токена

Refresh-токен - случайная строка от 10 до 72 символов, формат передачи - base64. 
Такое количество символов обусловлено хранением в виде brypt-хэша в БД (ограничение сверху)
и безопасностью (ограничение снизу). В БД вместе с токеном хранится время его действия

Access-токен - строка в формате JWT, содержит ID пользователя

При операциях с токеном соответсвующая ему запись в бд обновляется, что делает предыдущий недействительным. 

Например, в случае компрометации refresh-токена и обновлении пары refresh+access злоумышленником
пользовательский refresh-токен станет недействительным для обновления - потребуется повторная генерация.
После повторной генерации предыдущий refresh-токен в БД будет удален, поэтому злоумышленнику будет отказано
в обновлении пары по его refresh-токену
