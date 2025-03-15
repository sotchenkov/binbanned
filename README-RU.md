# binbanned

**binbanned** — это утилита для мониторинга логов Nginx в режиме реального времени и автоматической блокировки на уровне Nginx IP-адресов с подозрительными запросами. Поддерживает как JSON-логи, так и стандартный формат логов (Common Log Format).

### Возможности
- **Мониторинг логов в реальном времени:** Чтение логов с помощью режима tail и поддержка ротации файлов.
- **Поддержка различных форматов логов:** Обработка JSON и стандартных логов.
- **Фильтрация подозрительных запросов:** Блокировка IP-адресов, если URI содержит скрытые файлы (например, `/.env`, `/.git/config`) или если в user-agent присутствуют подозрительные шаблоны. Этот паттерн охватывает большинство плохих веб-ботов.
- **Whitelist:** Возможность запретить блокировку IP, указанных в отдельном файле.
- **Отправка сиганала Nginx на перечитывание конфигурации:** При добавлении ip в блок-лист, утилита отправляет nginx сигнал на перечитывание конфигурации.
- **Telegram уведомления:** Отправка уведомлений о новых блокировках через Telegram-бота.
- **Кастомные лейблы:** Возможность добавлять пользовательские лейблы в алерты и логи через флаг `--labels` (например, `--labels '{"server name": "my-server", "region":"us"}'`).


### Установка и сборка
Вы можете скачать уже собранный бинарный файл, либо собрать его самостоятельно: 

* **Скачивание бинаря**
Перейдите в релизы, выберите интересующую вас версию и скачайте бинарь
```bash
wget  https://github.com/sotchenkov/binbanned/releases/ (нужный файл)
```

* **Сборка:**
```bash
git clone git@github.com:sotchenkov/binbanned.git
cd binbanned
go build -o binbanned ./cmd/binbanned/main.go
```
Или для статической сборки:
```bash
CGO_ENABLED=0 go build -ldflags="-extldflags=-static" -o binbanned ./cmd/binbanned/main.go
```
### Использование

1. **Создайте необходимые файлы и задайте права:**

```bash
sudo touch /etc/nginx/conf.d/binbanned.conf
sudo touch /etc/nginx/ip-whitelist
sudo mkdir /var/log/binbanned
sudo chmod +x binbanned
sudo mv binbanned /usr/bin/
```

2. **Создайте systemd-сервис и укажите нужные параметры:**
```bash
sudo vim /etc/systemd/system/binbanned.service
```

```ini
[Unit]
Description=Binbanned service
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/bin/binbanned -telegram-token 'YOUR_TELEGRAM_BOT_TOKEN' -telegram-chat 'YOUR_TELEGRAM_CHAT_ID' --labels '{"server name": "my-server", "region":"ru"}'
Restart=on-failure
RestartSec=20
StandardOutput=append:/var/log/binbanned/binbanned.log
StandardError=append:/var/log/binbanned/binbanned.log

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl start binbanned
sudo systemctl status binbanned
```

### Параметры командной строки

    --logdir
    Директория с логами Nginx (по умолчанию: /var/log/nginx/).
    
    --banned
    Файл для записи заблокированных IP (по умолчанию: /etc/nginx/conf.d/binbanned.conf).
    
    --whitelist
    Файл с белым списком IP (по умолчанию: /etc/nginx/ip-whitelist).
    
    --reload-interval
    Интервал проверки новых банов и перезагрузки Nginx (по умолчанию: 10s).
    
    --parse-all
    Парсить логи с начала файлов (если задан).
    
    --telegram-token
    Telegram Bot token для отправки уведомлений.
    
    --telegram-chat
    Telegram Chat ID для уведомлений.
    
    --labels
    Пользовательские лейблы для алертов/логов в формате JSON.

### Настройка Nginx

Убедитесь, что в основном конфигурационном файле Nginx (/etc/nginx/nginx.conf) присутствует директива:
```bash
include /etc/nginx/conf.d/*.conf;
```
Это гарантирует, что Nginx будет применять настройки из файла с забаненными IP.


