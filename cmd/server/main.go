package main

import (
	"fmt"
	"github.com/iliaonishchenko/aggreg8/internal/handler"
	"github.com/iliaonishchenko/aggreg8/internal/service"
	"net/http"
)

/*
Сервер должен быть доступен по адресу http://localhost:8080, а также:

    Принимать и хранить произвольные метрики двух типов:
        Тип gauge, float64 — новое значение должно замещать предыдущее.
        Тип counter, int64 — новое значение должно добавляться к предыдущему, если какое-то значение уже было известно серверу.
    Принимать метрики по протоколу HTTP методом POST.
    Принимать данные в формате http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>, Content-Type: text/plain.
    При успешном приёме возвращать http.StatusOK.
    При попытке передать запрос без имени метрики возвращать http.StatusNotFound.
    При попытке передать запрос с некорректным типом метрики или значением возвращать http.StatusBadRequest.

Редиректы не поддерживаются.
Для хранения метрик объявите тип MemStorage. Рекомендуем использовать тип struct с полем-коллекцией внутри (slice или map).
В будущем это позволит добавлять к объекту хранилища новые поля — например, логгер или мьютекс, чтобы можно было использовать их в методах.
Опишите интерфейс для взаимодействия с этим хранилищем.
*/

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	fmt.Println("starting server")

	mux := http.NewServeMux()
	memStorage := service.NewMemStorage()
	updateHandler := handler.NewUpdateHandler(memStorage)
	
	mux.HandleFunc("/update/{type}/{name}/{value}", updateHandler.ServeHTTP)

	return http.ListenAndServe(`:8080`, mux)
}
