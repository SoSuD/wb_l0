Запускать через docker-compose

после запуска кафки прописать

```
docker exec -it kafka kafka-topics.sh \
  --create \
  --topic test-topic \
  --bootstrap-server localhost:9094 \
  --partitions 1 \
  --replication-factor 1
```

Пример кода для заполнения кафки:
```
import json
import sys
from confluent_kafka import Producer

CONFIG = {
    "brokers": ["localhost:9094"],
    "topic": "test-topic",
    "group_id": "1",  # для консьюмера, продюсеру не нужен — просто оставим тут для единообразия
}

DATA = {
   "order_uid": "b1513f1b712b84b6test",
   "track_number": "WBILMTESTTRACK",
   "entry": "WBIL",
   "delivery": {
      "name": "Test Testov",
      "phone": "+9720000000",
      "zip": "2639809",
      "city": "Kiryat Mozkin",
      "address": "Ploshad Mira 15",
      "region": "Kraiot",
      "email": "test@gmail.com"
   },
   "payment": {
      "transaction": "b563feb7b2b84b6test",
      "request_id": "",
      "currency": "USD",
      "provider": "wbpay",
      "amount": 1817,
      "payment_dt": 1637907727,
      "bank": "alpha",
      "delivery_cost": 1500,
      "goods_total": 317,
      "custom_fee": 0
   },
   "items": [
      {
         "chrt_id": 9934930,
         "track_number": "WBILMTESTTRACK",
         "price": 453,
         "rid": "ab4219087a764ae0btest",
         "name": "Mascaras",
         "sale": 30,
         "size": "0",
         "total_price": 317,
         "nm_id": 2389212,
         "brand": "Vivienne Sabo",
         "status": 202
      }
   ],
   "locale": "en",
   "internal_signature": "",
   "customer_id": "test",
   "delivery_service": "meest",
   "shardkey": "9",
   "sm_id": 99,
   "date_created": "2021-11-26T06:22:19Z",
   "oof_shard": "1"
}

def on_delivery(err, msg):
    if err is not None:
        print(f"Ошибка доставки: {err}", file=sys.stderr)
    else:
        print(f"OK → {msg.topic()}[{msg.partition()}] @ {msg.offset()}")

def main():
    producer = Producer({
        "bootstrap.servers": ",".join(CONFIG["brokers"]),
        # Ниже — опциональные, но полезные настройки:
        "enable.idempotence": True,     # безопасная доставка без дубликатов
        "acks": "all",
        "retries": 3,
        "linger.ms": 5,
        "batch.num.messages": 1000,
    })

    for i in range(1000, 2000):
        try:
            orderuid = f"b1513f1b712b84b6{i}"
            DATA["order_uid"] = orderuid
            DATA["payment"]["transaction"] = orderuid
            producer.produce(
                topic=CONFIG["topic"],
                value=json.dumps(DATA, ensure_ascii=False).encode("utf-8"),
                key=DATA.get("order_uid", None),
                on_delivery=on_delivery,
            )

        except BufferError as e:
            print(f"Локальный буфер переполнен: {e}", file=sys.stderr)

    # Обработка очереди событий и ожидание отправки
    producer.flush(15_000)

if __name__ == "__main__":
    main()
```
