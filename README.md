# Fraud Detection System

Gerçek zamanlı işlem dolandırıcılığı tespiti: bir transaction POST edildiği anda kural tabanlı analiz yapılır, sonuç WebSocket üzerinden dashboard'a iletilir ve MCP Server aracılığıyla yapay zeka ajanlarına sorgulanabilir hâle gelir.

---

## İçindekiler

1. [Proje Amacı ve Kapsamı](#1-proje-amacı-ve-kapsamı)
2. [Sistem Mimarisi](#2-sistem-mimarisi)
3. [Teknoloji Seçimleri ve Gerekçeleri](#3-teknoloji-seçimleri-ve-gerekçeleri)
4. [Kurulum](#4-kurulum)
5. [Kullanım Rehberi](#5-kullanım-rehberi)
6. [API Dokümantasyonu](#6-api-dokümantasyonu)
7. [MCP Sunucusu](#7-mcp-sunucusu)
8. [Scriptlerin Kullanımı](#8-scriptlerin-kullanımı)
9. [Sorun Giderme](#9-sorun-giderme)

---

## 1. Proje Amacı ve Kapsamı

Bu sistem, finansal işlemleri gerçek zamanlı olarak üç farklı kural ile analiz eder:

| Kural | Açıklama | Tetikleyici |
|-------|----------|-------------|
| **Velocity Limit** | Aynı kullanıcıdan 1 dakikada 5'ten fazla işlem | Hız tabanlı saldırı |
| **Amount Anomaly** | İşlem tutarı kullanıcının 24 saatlik ortalamasının 3 katını geçiyor | Anormal harcama |
| **Impossible Travel** | İki işlem arasındaki mesafe/süre oranı uçak hızını (900 km/h) aşıyor | Konum sahteciliği |

İki veya daha fazla kural ihlal edildiğinde işlem `fraud` olarak işaretlenir. Tek ihlal `approved` sonucu vermez — bu bilinçli bir tasarım tercihidir: false positive oranını düşük tutmak için eşik yüksek tutulmuştur.

---

## 2. Sistem Mimarisi

```
┌─────────────────────────────────────────────────────────────────────┐
│                          Docker Network                             │
│                                                                     │
│  ┌──────────┐   POST /transactions   ┌───────────────────────────┐  │
│  │          │ ─────────────────────► │   API Server              │  │
│  │ Frontend │                        │   (Fiber, :8080)          │  │
│  │ Next.js  │   WebSocket :8081      │                           │  │
│  │ :3000    │ ◄───────────────────── │   WebSocket Hub (:8081)   │  │
│  └──────────┘                        └─────────────┬─────────────┘  │
│                                                    │               │
│                                  transactions exchange (RabbitMQ)  │
│                                                    │               │
│                                                    ▼               │
│  ┌──────────┐    UpdateStatus       ┌───────────────────────────┐  │
│  │ MongoDB  │ ◄─────────────────────│   Fraud Worker            │  │
│  │          │                       │                           │  │
│  │          │   GetUserStats        │  ┌─────────────────────┐  │  │
│  └──────────┘ ◄─────────────────────│  │ Velocity Rule       │  │  │
│                                     │  │ Amount Anomaly Rule  │  │  │
│  ┌──────────┐   Cache Ops           │  │ Impossible Travel   │  │  │
│  │  Redis   │ ◄────────────────────►│  └─────────────────────┘  │  │
│  │          │                       └─────────────┬─────────────┘  │
│  └──────────┘                                     │               │
│                                    events exchange (RabbitMQ)      │
│                                                    │               │
│                                  API Server → WebSocket Broadcast  │
│                                                                     │
│  ┌─────────────────────────────┐                                   │
│  │  MCP Server (stdio binary)  │  ← AI ajanları buraya bağlanır   │
│  │  - get_recent_frauds        │    (Claude Desktop vb.)          │
│  │  - check_user_status        │                                   │
│  └─────────────────────────────┘                                   │
└─────────────────────────────────────────────────────────────────────┘
```

### Bileşenler

| Bileşen | Port | Sorumluluk |
|---------|------|-----------|
| **API Server** (`cmd/server`) | 8080 (HTTP), 8081 (WS) | REST endpoint'leri, WebSocket hub, transaction kabulü |
| **Fraud Worker** (`cmd/worker`) | — | RabbitMQ'dan okur, kuralları çalıştırır, MongoDB'yi günceller, event yayar |
| **MCP Server** (`cmd/mcp`) | stdio | AI ajanları için `get_recent_frauds` ve `check_user_status` araçları |
| **Frontend** | 3000 | Dashboard, canlı akış görünümü, kullanıcı detayları |
| **MongoDB** | 27017 | Kalıcı transaction deposu |
| **RabbitMQ** | 5672, 15672 | Async kuyruk (`transactions`, `events` exchange) |
| **Redis** | 6379 | Fraud kural cache'i (velocity, amount avg, son konum) |

### Mesaj Akışı

```
POST /transactions
    → MongoDB'ye yaz (status=pending)
    → RabbitMQ transactions exchange'e publish
    → HTTP 201 dön  (async — kullanıcı beklemez)

Worker:
    → transactions.process queue'yu oku
    → 3 kuralı çalıştır (Redis cache ile)
    → MongoDB'de status güncelle (approved / fraud)
    → events exchange'e publish

API Server (event consumer):
    → events.<uuid> queue'yu oku
    → Bağlı WebSocket istemcilerine broadcast
```

---

## 3. Teknoloji Seçimleri ve Gerekçeleri

### Go — Backend

Düşük gecikme, yüksek eşzamanlılık gerektiren fraud analizi için Go'nun goroutine modeli idealdir. Worker birden fazla transaction'ı eş zamanlı işleyebilir;

### Fiber v3 — HTTP Framework

Fasthttp tabanlı, `net/http`'den ~2–3× daha hızlı. Fraud tespiti gibi düşük gecikme gerektiren sistemlerde routing overhead önemlidir.

### RabbitMQ — Mesaj Kuyruğu

API Server → Worker iletişimi için asenkron kuyruk kullanılmasının nedenleri:

- **Backpressure:** Worker meşgulse transaction kaybolmaz, kuyrukta birikir.
- **Dayanıklılık:** `DeliveryMode=Persistent` + `durable=true` → restart'ta mesaj kaybolmaz.
- **Decoupling:** Worker ölse API yanıt vermeye devam eder.

Fanout exchange tercih edildi: Hem `transactions` hem `events` exchange'i fanout olarak tanımlandı. Routing key gereksizdir; ileride birden fazla consumer eklenebilir.

### Redis — Cache ve Anomali Tespitindeki Cache Yönetimi Kararları

#### Velocity Cache (`fraud:velocity:{userID}`)

```
INCR  fraud:velocity:user-001   # sayaç arttır
EXPIRE fraud:velocity:user-001 60   # her işlemde 60 sn TTL yenile
```

**Karar — fixed window:** TTL her `INCR`'de yenilenir. Kullanıcı dakikada 4 işlem yaparsa sayaç sıfırlanmadan devam eder. **Bypass riski:** Her 60 saniyede tam 5 işlem göndererek tespit atlanabilir. Bu, sliding window yerine fixed window kullanmanın bilinçle kabul edilen bir tradeoff'udur — implementasyon sadeliği için Redis sorted set yerine basit INCR/EXPIRE kullanıldı.

#### Amount Average Cache (`fraud:avg_amount:{userID}`)

```
HSET fraud:avg_amount:user-001 sum 345.50 count 3
EXPIRE  (sadece count==1 koşulunda — ilk işlemde)
```

**Karar — fixed window, sadece approved işlemler:** TTL yalnızca ilk hash oluşturulduğunda set edilir. Sliding window için Redis sorted set ile zaman damgalı değerler tutulması gerekirdi (~3× bellek ve işlem maliyeti). Fraud tespiti için "son N saatin yaklaşık ortalaması" yeterlidir.

**Neden sadece approved işlemler ortalamayı etkiliyor?** `UpdateAmountAverage` yalnızca fraud tespit edilmediğinde çağrılır. Fraud işlemler ortalamayı yukarı çekmemeli; aksi takdirde saldırgan yüksek tutarlı işlemler göndererek eşiği manipüle edebilir.

#### Location Cache (`fraud:location:{userID}`)

```
HSET fraud:location:user-001 lat 41.0082 lon 28.9784 created_at <unix>
EXPIRE fraud:location:user-001 86400   # 24 saat
```

**Karar — sadece approved işlemler günceller:** Fraud işlemin lokasyonu cache'e yazılmaz. Sonraki işlem, son bilinen **geçerli** konumla karşılaştırılmalıdır, fraud lokasyonuyla değil.

**Haversine formülü** kullanıldı: Küre yüzeyi mesafesi, düz mesafeden ~%3–4 daha doğrudur. 900 km/h eşiği ticari uçakların azami cruising hızıdır — fiziksel olarak imkânsız seyahati tespit eder.

### MongoDB

Transaction verisinin doğal yapısı document-oriented'a uygundur (`fraud_reasons` dizisi, esnek status alanı). Compound index `{user_id: 1, created_at: -1}` her iki kritik sorguyu tek index ile karşılar:

- `FindByUserID` → `user_id` prefix + `created_at` sıralama (sort bedava gelir)
- `FindFraudsBetween` → status filtresi + `created_at` range

### Next.js + HeroUI v3

HeroUI v3, React Aria tabanlıdır — erişilebilirlik primitifleri built-in gelir. Tailwind CSS v4 entegrasyonu, design token'ları CSS custom property olarak tutar; tema değişkeni override edilebilir. TanStack Query, sunucu durumunu cache'ler ve stale-while-revalidate stratejisiyle gereksiz ağ isteklerini önler.

### MCP — Model Context Protocol

JSON-RPC 2.0 over stdio. Bu standart MCP transport'u Claude Desktop, Continue, Cursor gibi AI araçlarıyla doğrudan çalışır. HTTP transport seçilmedi: stdio, ek bir port/sunuç sürecine gerek kalmadan doğrudan process fork ile çalışır ve güvenlik yüzeyi küçüktür.

---

## 4. Kurulum

### Ön Gereksinimler

| Araç | Minimum Versiyon | Kontrol |
|------|-----------------|---------|
| Docker | 24.0+ | `docker --version` |
| Docker Compose | 2.20+ (plugin) | `docker compose version` |
| Git | herhangi | `git --version` |

> MCP Server'ı yerel çalıştırmak için ek olarak **Go 1.25+** gereklidir.

---

### Adım 1 — Repoyu klonla

```bash
git clone https://github.com/AkifhanIlgaz/fraud-detection
cd fraud-detection
```

### Adım 2 — Tüm sistemi başlat

```bash
docker compose up --build -d
```

Bu komut şu sırayla başlatır:

1. MongoDB, Redis, RabbitMQ (infrastructure — health check tamamlanana kadar bekler)
2. API Server ve Fraud Worker (infrastructure `healthy` olduktan sonra)
3. Frontend (API `healthy` olduktan sonra)

> **İlk başlatma:** Image'lar sıfırdan build edildiği için 3–5 dakika sürebilir.

### Adım 3 — Servislerin durumunu doğrula

```bash
docker compose ps
```

Beklenen çıktı — tüm servisler `running`, health check'ler `healthy`:

```
NAME              IMAGE            STATUS
fraud-mongo       mongo:7          running (healthy)
fraud-redis       redis:7-alpine   running (healthy)
fraud-rabbitmq    rabbitmq:3-...   running (healthy)
fraud-api         fraud-api        running (healthy)
fraud-worker      fraud-worker     running
fraud-frontend    fraud-frontend   running (healthy)
```

API'yi doğrula:

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

### Adım 4 — Demo veri yükle (isteğe bağlı)

Dashboard'da anlamlı bir grafik görmek için 30 günlük seed verisi yükle:

```bash
docker compose --profile seed run --rm seed
```

Bu komut ~5 dakika sürer (her fraud pair için worker'ın işlemesi beklenir).

### Adım 5 — Dashboard'u aç

```
http://localhost:3000
```

---

### Sistemi Durdurma

```bash
# Servisleri durdur, veriler korunur
docker compose down

# Servisleri ve tüm verileri tamamen sil
docker compose down -v
```

---

### Servis URL'leri

| Servis | URL | Kimlik Bilgisi |
|--------|-----|----------------|
| Frontend Dashboard | http://localhost:3000 | — |
| REST API | http://localhost:8080/api/v1 | — |
| WebSocket | ws://localhost:8081/transactions | — |
| RabbitMQ Management UI | http://localhost:15672 | guest / guest |
| MongoDB | mongodb://localhost:27017 | — |
| Redis | redis://localhost:6379 | — |

---

## 5. Kullanım Rehberi

### Dashboard Sayfaları

| Sayfa | URL | İçerik |
|-------|-----|--------|
| Ana Sayfa | `/` | Genel istatistikler, 30 günlük fraud grafiği |
| Canlı Akış | `/live` | WebSocket tabanlı gerçek zamanlı işlem feed'i |
| Fraud Listesi | `/frauds` | Sayfalı fraud transaction listesi |
| Kullanıcı Detayı | `/users/{userID}` | İşlem geçmişi ve trust score |

### Transaction Oluşturma

```bash
curl -s -X POST http://localhost:8080/api/v1/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-001",
    "amount": 250.00,
    "lat": 41.0082,
    "lon": 28.9784
  }'
```

İşlem `pending` olarak kaydedilir, worker anında analiz eder ve durumu günceller.

### Fraud Senaryolarını Manuel Tetikleme

**Impossible Travel:**

```bash
# İstanbul
curl -X POST http://localhost:8080/api/v1/transactions \
  -H "Content-Type: application/json" \
  -d '{"user_id":"test-travel","amount":150,"lat":41.0082,"lon":28.9784}'

sleep 2   # worker konumu cache'lesin

# New York — fiziksel olarak imkânsız
curl -X POST http://localhost:8080/api/v1/transactions \
  -H "Content-Type: application/json" \
  -d '{"user_id":"test-travel","amount":150,"lat":40.7128,"lon":-74.006}'
```

**Velocity (1 dakikada 6+ işlem):**

```bash
for i in $(seq 1 7); do
  curl -s -X POST http://localhost:8080/api/v1/transactions \
    -H "Content-Type: application/json" \
    -d '{"user_id":"test-velocity","amount":100,"lat":41.0082,"lon":28.9784}' &
done
```

**Amount Anomaly:**

```bash
# 3 normal işlemle ortalama oluştur (~$100)
for i in 1 2 3; do
  curl -s -X POST http://localhost:8080/api/v1/transactions \
    -H "Content-Type: application/json" \
    -d '{"user_id":"test-amount","amount":100,"lat":41.0082,"lon":28.9784}'
  sleep 0.5
done

sleep 3   # worker işlesin

# 10× spike — average * 3 eşiğini aşar
curl -X POST http://localhost:8080/api/v1/transactions \
  -H "Content-Type: application/json" \
  -d '{"user_id":"test-amount","amount":1500,"lat":41.0082,"lon":28.9784}'
```

---

## 6. API Dokümantasyonu

**Base URL:** `http://localhost:8080`

**Yanıt formatı:**
- Başarı: `{"success": true, "data": {...}}`
- Hata: `{"error": "mesaj"}`

---

### `GET /health`

Servis sağlık kontrolü.

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

---

### `POST /api/v1/transactions`

Yeni transaction oluşturur ve fraud analizini kuyruğa alır.

**İstek Gövdesi:**

| Alan | Tip | Zorunlu | Açıklama |
|------|-----|---------|----------|
| `user_id` | string | Evet | Kullanıcı tanımlayıcısı |
| `amount` | float | Evet | Tutar (> 0) |
| `lat` | float | Evet | Enlem (−90 ile 90) |
| `lon` | float | Evet | Boylam (−180 ile 180) |
| `created_at` | string RFC3339 | Hayır | İşlem zamanı (varsayılan: şimdi) |

**Yanıt `201`:**

```json
{
  "success": true,
  "data": {
    "id": "6645f3a2b1c8d9e0f1a2b3c4",
    "user_id": "user-001",
    "amount": 750.00,
    "lat": 41.0082,
    "lon": 28.9784,
    "status": "pending",
    "created_at": "2025-01-15T10:30:00Z",
    "fraud_reasons": null
  }
}
```

---

### `GET /api/v1/transactions/user/:userID`

Kullanıcının işlemlerini sayfalı döner.

**Query Parametreleri:** `page` (varsayılan: 1), `limit` (varsayılan: 20, maks: 100)

```bash
curl "http://localhost:8080/api/v1/transactions/user/user-001?page=1&limit=20"
```

**Yanıt `200`:**

```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "...",
        "user_id": "user-001",
        "amount": 150.00,
        "status": "approved",
        "fraud_reasons": null
      }
    ],
    "meta": {
      "page": 1,
      "limit": 20,
      "total": 47,
      "total_pages": 3,
      "has_next": true,
      "has_prev": false
    }
  }
}
```

---

### `GET /api/v1/transactions/user/:userID/trust-score`

Kullanıcı güven skoru ve risk seviyesi.

```bash
curl http://localhost:8080/api/v1/transactions/user/user-001/trust-score
```

**Yanıt `200`:**

```json
{
  "success": true,
  "data": {
    "user_id": "user-001",
    "score": 85.7,
    "risk_level": "low",
    "total": 21,
    "fraud_count": 3
  }
}
```

| Score | Risk Level |
|-------|-----------|
| ≥ 80 | `low` |
| 50–79 | `medium` |
| < 50 | `high` |

---

### `GET /api/v1/transactions/frauds`

Zaman aralığındaki fraud işlemleri.

**Query Parametreleri:**

| Parametre | Tip | Zorunlu | Açıklama |
|-----------|-----|---------|----------|
| `from` | RFC3339 | Evet | Başlangıç |
| `to` | RFC3339 | Evet | Bitiş |
| `page` | int | Hayır | Sayfa (varsayılan: 1) |
| `limit` | int | Hayır | Limit (varsayılan: 20) |

```bash
curl "http://localhost:8080/api/v1/transactions/frauds?from=2025-01-01T00:00:00Z&to=2025-01-31T23:59:59Z"
```

---

### `PATCH /api/v1/transactions/:id/status`

Transaction durumunu manuel güncelle.

```bash
curl -X PATCH http://localhost:8080/api/v1/transactions/<id>/status \
  -H "Content-Type: application/json" \
  -d '{"status": "approved"}'
```

Geçerli değerler: `pending`, `approved`, `fraud`

---

### WebSocket

```
ws://localhost:8081/transactions
```

Her analiz tamamlandığında yayımlanan mesaj:

```json
{
  "transaction_id": "6645f3a2b1c8d9e0f1a2b3c4",
  "user_id": "user-001",
  "status": "fraud",
  "amount": 2500.00,
  "fraud_reasons": ["impossible_travel", "amount_anomaly"],
  "created_at": "2025-01-15T10:30:05Z"
}
```

Fraud nedenleri:

| Değer | Açıklama |
|-------|----------|
| `velocity_limit_exceeded` | 1 dakikada > 5 işlem |
| `amount_anomaly` | Tutar ortalamayı 3× aşıyor |
| `impossible_travel` | Mesafe/süre > 900 km/h |

---

## 7. MCP Sunucusu

MCP (Model Context Protocol) sunucusu **JSON-RPC 2.0 over stdio** kullanır. Claude Desktop veya başka bir MCP-uyumlu AI ajanına local binary olarak eklenir ve doğrudan MongoDB'ye bağlanarak veri sorgular.

### Araçlar

#### `get_recent_frauds`

Son N saatteki fraud işlemlerini döner.

| Parametre | Tip | Varsayılan | Açıklama |
|-----------|-----|-----------|----------|
| `hours_back` | integer | 24 | Kaç saat geriye bakılacağı (maks: 720) |
| `limit` | integer | 10 | Maksimum kayıt sayısı (maks: 100) |

**Örnek çıktı:**
```json
{
  "items": [
    {
      "id": "...",
      "user_id": "user-003",
      "amount": 1850.00,
      "status": "fraud",
      "fraud_reasons": ["impossible_travel", "amount_anomaly"],
      "created_at": "2025-01-15T09:45:00Z"
    }
  ],
  "meta": {"page": 1, "limit": 10, "total": 3, "total_pages": 1}
}
```

#### `check_user_status`

Kullanıcının güven skoru ve risk durumu.

| Parametre | Tip | Zorunlu | Açıklama |
|-----------|-----|---------|----------|
| `user_id` | string | Evet | Sorgulanacak kullanıcı |

**Örnek çıktı:**
```json
{
  "user_id": "user-001",
  "score": 42.8,
  "risk_level": "high",
  "total": 14,
  "fraud_count": 8
}
```

---

### MCP Binary'yi Derleme

Docker servislerine gerek yoktur; sadece Go gereklidir:

```bash
cd backend
go build -o mcp ./cmd/mcp/
```

---

### Claude Desktop Entegrasyonu

`~/.claude/claude_desktop_config.json` dosyasını düzenle:

```json
{
  "mcpServers": {
    "fraud-detection": {
      "command": "/tam/yol/fraud-detection/backend/mcp",
      "env": {
        "MONGO_URI":    "mongodb://localhost:27017",
        "REDIS_ADDR":   "localhost:6379",
        "RABBITMQ_URL": "amqp://guest:guest@localhost:5672/"
      }
    }
  }
}
```

> **Not:** `localhost` portları Docker'ın host-mapped portlarını kullanır. MCP binary host üzerinde çalışır, container içinde değil.

Claude Desktop'ı yeniden başlatın. Ardından Claude'a şöyle sorabilirsiniz:

```
Son 24 saatte kaç fraud işlem var?
user-001'in risk durumu nedir?
Son 3 gün içinde hangi fraud kararları verildi?
```

---

### MCP'yi Manuel Test Etme

MCP protokolü newline-delimited JSON-RPC mesajlarıdır. `printf` ile test edilebilir:

**1. Binary'yi derle:**

```bash
cd backend && go build -o /tmp/fraud-mcp ./cmd/mcp/
```

**2. Env var'ları export et (Docker çalışıyor olmalı):**

```bash
export MONGO_URI="mongodb://localhost:27017"
export REDIS_ADDR="localhost:6379"
export RABBITMQ_URL="amqp://guest:guest@localhost:5672/"
```

**3. Initialize + araç listesi:**

```bash
printf '%s\n%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' \
  | /tmp/fraud-mcp
```

Beklenen çıktı:
```json
{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2024-11-05","capabilities":{"tools":{}},"serverInfo":{"name":"fraud-detection","version":"1.0.0"}}}
{"jsonrpc":"2.0","id":2,"result":{"tools":[{"name":"get_recent_frauds",...},{"name":"check_user_status",...}]}}
```

**4. `get_recent_frauds` çağrısı:**

```bash
printf '%s\n%s\n%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_recent_frauds","arguments":{"hours_back":48,"limit":5}}}' \
  | /tmp/fraud-mcp
```

**5. `check_user_status` çağrısı:**

```bash
printf '%s\n%s\n%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"check_user_status","arguments":{"user_id":"user-001"}}}' \
  | /tmp/fraud-mcp
```

---

## 8. Scriptlerin Kullanımı

Script'ler `scripts/` dizinindedir. Çalıştırmadan önce:

```bash
chmod +x scripts/auto-test.sh scripts/manual-input.sh
```

---

### `auto-test.sh` — Otomatik Yük Testi

API'yi rastgele ve anomalili işlemlerle bombalar. Flood testi ve dashboard demo'su için kullanılır.

**Kullanım:**

```bash
./scripts/auto-test.sh [seçenekler]
```

**Parametreler:**

| Parametre | Varsayılan | Açıklama |
|-----------|-----------|----------|
| `--duration=<sn>` | 60 | Kaç saniye çalışacağı |
| `--rate=<istek/sn>` | 2 | Hedef istek hızı |
| `--anomaly-chance=<0-100>` | 20 | Anomali tetikleme olasılığı (%) |
| `--users=<adet>` | 10 | Sentetik kullanıcı sayısı |
| `--api=<url>` | http://localhost:8080/api/v1/transactions | API endpoint |

**Örnekler:**

```bash
# Varsayılan — 60 sn, 2 istek/sn, %20 anomali
./scripts/auto-test.sh

# Yoğun test — 5 dakika, 5 istek/sn, %30 anomali
./scripts/auto-test.sh --duration=300 --rate=5 --anomaly-chance=30

# Sadece anomali
./scripts/auto-test.sh --anomaly-chance=100 --duration=30
```

**Üretilen Anomali Tipleri:**

| Tip | Açıklama |
|-----|----------|
| `velocity` | Aynı kullanıcıdan 8 hızlı işlem (> 5/dk) |
| `amount` | 3 baseline + 10× spike (tutar eşiğini aşar) |
| `travel` | İstanbul → New York (1 sn arayla, imkânsız seyahat) |

---

### `manual-input.sh` — Manuel İşlem Gönderimi

Tek bir transaction gönderir. Belirli fraud senaryolarını hedefli test etmek için kullanılır.

**Kullanım:**

```bash
./scripts/manual-input.sh <user_id> <amount> <lokasyon>
```

**Lokasyon formatları:**

| Tip | Değerler |
|-----|---------|
| Şehir adı | `istanbul`, `newyork`, `tokyo`, `sydney`, `dubai`, `losangeles`, `singapore`, `london`, `saopaulo` |
| Koordinat | `"41.0082,28.9784"` |

**Örnekler:**

```bash
# Normal işlem
./scripts/manual-input.sh user-007 250.00 istanbul

# Yüksek tutarlı işlem
./scripts/manual-input.sh user-007 9500.00 newyork

# Koordinat ile
./scripts/manual-input.sh user-007 120.50 "41.0082,28.9784"

# Farklı API URL
API_URL=http://localhost:8080/api/v1/transactions \
  ./scripts/manual-input.sh user-001 500 tokyo
```

**Çıktı:**
```
Sending transaction...
  user_id : user-007
  amount  : $250.00
  location: istanbul (41.0082, 28.9784)
  endpoint: http://localhost:8080/api/v1/transactions

OK  (HTTP 201)
  transaction id: 6645f3a2b1c8d9e0f1a2b3c4
```

---

## 9. Sorun Giderme

### Servislerin durumunu görme

```bash
docker compose ps
docker compose logs -f api
docker compose logs -f worker
docker compose logs -f frontend
```

---

### `fraud-api` başlamıyor: "mongo connect" / "rabbitmq dial"

Infrastructure servisleri henüz hazır olmayabilir. Manuel kontrol:

```bash
docker exec fraud-mongo    mongosh --eval "db.adminCommand('ping')"
docker exec fraud-redis    redis-cli ping
docker exec fraud-rabbitmq rabbitmq-diagnostics check_port_connectivity
```

Tüm servisler `healthy` değilse bekleyin ve yeniden başlatın:

```bash
docker compose restart api worker
```

---

### Frontend API'ye bağlanamıyor

`NEXT_PUBLIC_*` değişkenleri Next.js'te **build time**'da sabitlenir — runtime'da değiştirilemez. Farklı bir URL için yeniden build gereklidir:

```bash
docker compose build \
  --build-arg NEXT_PUBLIC_API_URL=http://your-server:8080/api/v1 \
  --build-arg NEXT_PUBLIC_WS_URL=ws://your-server:8081/transactions \
  frontend
docker compose up -d frontend
```

---

### WebSocket bağlantısı kurulamıyor

```bash
# Port açık mı?
docker compose port api 8081

# WebSocket handshake testi
curl -i -N \
  -H "Connection: Upgrade" \
  -H "Upgrade: websocket" \
  -H "Sec-WebSocket-Version: 13" \
  -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
  http://localhost:8081/transactions
```

---

### Fraud işlemleri oluşmuyor — worker sorunu

```bash
docker compose logs -f worker

# RabbitMQ Management UI'da kuyruğu kontrol et:
# http://localhost:15672 → Queues → transactions.process
# Messages birikiyorsa worker consume etmiyor:
docker compose restart worker
```

---

### Seed verisi yüklenmiyor

```bash
# API sağlıklı mı?
curl http://localhost:8080/health

# Seed logları
docker compose --profile seed run --rm seed 2>&1 | tee seed.log
```

---

### MCP Server "mongo connect" hatası

MCP binary host üzerinde çalışır. Env var'ların Docker'ın host-mapped portlarını işaret ettiğini doğrulayın:

```bash
MONGO_URI=mongodb://localhost:27017 \
REDIS_ADDR=localhost:6379 \
RABBITMQ_URL=amqp://guest:guest@localhost:5672/ \
  ./backend/mcp

# MongoDB'ye host'tan erişebildiğinizi test edin:
mongosh mongodb://localhost:27017 --eval "db.adminCommand('ping')"
```

---

### Port çakışması

Sistemde zaten çalışan MongoDB/Redis/RabbitMQ varsa host portları değiştirilebilir:

```bash
# Kullanımdaki portları gör
lsof -i :27017 -i :6379 -i :5672 -i :8080 -i :8081 -i :3000
```

`docker-compose.yml`'de ilgili servisin `ports` satırını değiştirin:

```yaml
# Örnek: MongoDB'yi 27018 host portuna taşı
ports:
  - "27018:27017"
```

---

### Docker build başarısız — Go versiyonu

```bash
# golang:1.25-alpine mevcut değilse:
docker pull golang:1.25-alpine || echo "Mevcut değil, 1.24 kullanın"
```

`backend/Dockerfile` içindeki `FROM golang:1.25-alpine AS builder` satırını `golang:1.24-alpine` ile değiştirin.

---

### Veriyi tamamen sıfırla

```bash
docker compose down -v          # container ve volume'ları sil
docker compose up --build -d    # yeniden başlat
docker compose --profile seed run --rm seed  # demo verisi yükle
```
