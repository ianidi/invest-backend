Golang + gin + postgresql.

# Run

cd /var/www
curl -OL "https://github.com/caddyserver/caddy/releases/latest/download/caddy_2.0.0_linux_amd64.deb"

sudo dpkg -i caddy_2.0.0_linux_amd64.deb

sudo nano caddy.json

```json
{
  "apps": {
      "http": {
          "servers": {
              "srv0": {
                  "automatic_https": {
                    "disable": true,
                    "disable_redirects": true
                  },
                  "listen": [«:80»],
                  "routes": [{
                      "handle": [{
                          "handler": "subroute",
                          "routes": [{
                              "handle": [{
                                  "handler": "reverse_proxy",
                                  "upstreams": [{
                                      "dial": "188.225.74.11:4001"
                                  }]
                              }]
                          }]
                      }],
                      "match": [{
                          "host": ["api.acces-plateforme.online"]
                      }],
                      "terminal": true
                  }, {
                      "handle": [{
                          "handler": "subroute",
                          "routes": [{
                              "handle": [{
                                  "handler": "reverse_proxy",
                                  "upstreams": [{
                                      "dial": "188.225.74.11:4000"
                                  }]
                              }]
                          }]
                      }],
                      "match": [{
                          "host": ["acces-plateforme.online"]
                      }],
                      "terminal": true
                  }]
              }
          }
      }
  }
}
```

caddy start
curl localhost:2019/config/
curl localhost:2019/load -X POST -H "Content-Type: application/json" -d @caddy.json

docker build . -t ianidi/exchange-server:1.0.6
docker push ianidi/exchange-server:1.0.6
docker run -d -p 4001:4000 ianidi/exchange-server:1.0.6

docker build . -t ianidi/exchange:1.0.6
docker push ianidi/exchange:1.0.6
docker run -d -p 4000:3000 ianidi/exchange:1.0.6

# Настройка переменных окружения

connection формата postgresql://user:password@dn_host.amazonaws.com/dbname
port формата :3000
jwt_secret формата набор символов

# Документация API

Swagger по адресу http://адрес_сервера:порт_сервера/swagger/index.html, логин: eun3denw, пароль: fnj43jnfi3

# Структура

```
|
├── main.go
├── log.log              //Логи сервера, полученные в процессе работы
├── static               //Медиафайлы, доступные по адресу http://адрес_сервера:порт_сервера/static (к примеру фотографии переговорной комнаты)
├── migrations           //Миграции БД
├── internal
│   ├── cometchat        //cometchat.com API клиент для регистрации пользователей в чате
│   ├── jwt              //Функции авторизации с помощью JWT токенов
│   ├── mail             //Отправка почты с помощью SMTP
│   └── timezone         //Настройки часового пояса
└── api
    ├── public           // Эндпоинты, доступные любому пользователю без авторизации
    ├── member           // Эндпоинты, предназначенные для гостей, вошедших в систему
    ├── device           // Эндпоинты, предназначенные для планшетов в переговорных комнатах
    ├── operator         // Эндпоинты, предназначенные для операторов, позволяет управлять коворкингами, к которым они принадлежат
    └── admin           // Эндпоинты, предназначенные для администраторов, позволяет в полной мере управлять системой

```

# Миграции БД

https://github.com/golang-migrate/migrate/tree/master/cmd/migrate

migrate -database YOUR_DATBASE_URL -path /migrations up

# Дополнительная информация

Комментарии в main.go и других файлах
