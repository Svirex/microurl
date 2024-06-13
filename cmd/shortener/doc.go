// Сервис сокращения ссылок.
//
// ## Сборка исходников
//
// Из директории с файлом `go.mod` выполнить сборку
//
//	go build -o app cmd/shortener/main.go
//
// Есть три флага компиляции, в которые можно передать информацию о сборке:
// - main.buildVersion - строка, версия сборки
// - main.buildDate - строка, дата сборки
// - main.buildCommit - строка, хэш коммита, который используется при сборке
//
// Сборка с флагами компиляции:
//
//	go build -o app -ldflags "-X main.buildVersion=v1.0.0 -X 'main.buildDate=$(date +'%d-%m-%Y')' -X 'main.buildCommit=$(git log -n 1 --pretty=format:'%H')'" cmd/shortener/main.go
//
// Дата и хэш коммита автоматически подставятся, требуется только изменить версию сборки.
//
// ## Запуск сервиса
//
// Для запуска сервиса предусмотрены следующие флаги (переменные окружения):
//
// - a (SERVER_ADDRESS) - адрес сервиса, по умолчанию localhost:8080
// - b (BASE_URL) - адрес, который будет использоваться в сокращенной ссылке, по умолчанию равен адресу сервера
// - f (FILE_STORAGE_PATH) - путь до файла, где будут хранится записи в случае запуска сервиса с хранением в файле
// - d (DATABASE_DSN) - адрес подключения к БД Postgres, может имееть вид `postgres://username:password@localhost:5432/database_name` или `host=localhost port=5432 dbname=mydb user=user password=pass`
// - k (SECRET_KEY) - секретный ключ для создания JWT токена
//
// Запуск сервиса может быть выполнен с тремя хранилищами: в памяти, в файле или в БД.
//
// ### Запуск сервиса с подключение к БД
//
//	./app -d postgres://username:password@localhost:5432/database_name
//
// Флаг `-f` в этом случае игнорируется.
//
// ### Запуск сервиса с файловым хранилищем
//
//	./app -f db.txt
//
// ### Запуск сервиса с хранилищем в памяти
//
//	./app -f ''
package main
