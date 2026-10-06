# Distributed Microservices Architecture — Point of Sale Platform

A production-grade, resilient, and fully observable **point of sale (POS) microservices backend**
built in **Go (Golang)**. Retail workloads — identity, merchants, catalog, cashiers, orders, and
payments — are split across self-contained, independently deployable services that communicate
synchronously over **gRPC** and asynchronously over **Apache Kafka**, behind a unified **REST API
Gateway** (Echo + NGINX).

Analytical reads never hit the transactional tier: a **CQRS-style OLAP layer** powered by
**ClickHouse** is fed by `stats_writer` (a Kafka consumer with backfill bootstrapping) and served
by `stats_reader`, a gRPC analytics query service that returns monthly/yearly revenue, best-selling
products and categories, transaction success rates, and cashier performance.

Persistence is **database-per-bounded-context**: **five isolated PostgreSQL 17 clusters —
`pos_identity`, `pos_merchant`, `pos_catalog`, `pos_sales`, `pos_email` — each fronted by its own
PgBouncer pooler**. There is no shared database and no generic `DB_HOST`; a service can only ever
reach the cluster its context owns.

---

## Table of Contents

1. [Key Features](#key-features)
2. [Architecture Overview](#architecture-overview)
3. [Service Catalog](#service-catalog)
4. [Database Layer — PostgreSQL Cluster & PgBouncer](#database-layer--postgresql-cluster--pgbouncer)
5. [Internal Service Architecture](#internal-service-architecture)
6. [Data & Event Flow](#data--event-flow)
7. [OLAP Analytics Layer](#olap-analytics-layer)
8. [Observability Architecture](#observability-architecture)
9. [Deployment Architectures](#deployment-architectures)
10. [Technology Stack](#technology-stack)
11. [Getting Started](#getting-started)
12. [Port Map Registry](#port-map-registry)
13. [Justfile Reference](#justfile-reference)
14. [Workspace Directory Tree](#workspace-directory-tree)

---

## Key Features

| Domain | Capabilities |
| :--- | :--- |
| **Auth & Users** | Registration, login, stateless JWT access/refresh token lifecycle, password reset, OTP email verification, `GetMe` profile resolver |
| **Roles & RBAC** | Permission configuration, granular access control, sub-second permission evaluation cached in Redis; role↔user assignment now lives in its own `UserRoleService` |
| **Merchants** | Merchant onboarding, profile and business data, soft-delete/restore with full data restoration |
| **Categories** | Category CRUD, search, trash/restore flows, monthly sold-volume analytics |
| **Products** | Product catalog with name, price, category assignment, stock tracking, image upload, monthly sold-quantity analytics |
| **Cashiers** | Cashier accounts per merchant with monthly order/revenue performance analytics |
| **Orders & Order Items** | Cart-style order composition with line items and quantities, automatic total calculation, soft-delete/trash/restore audit records |
| **Transactions** | Payment settlement against orders (cash and others), change calculation, status tracking, monthly success-rate reports |
| **OLAP Analytics** | ClickHouse warehouse fed by `stats_writer` (Kafka consumer + backfill) and served by `stats_reader` (gRPC analytics) |
| **Email Worker** | Kafka-driven worker dispatching OTPs, login alerts, and merchant onboarding notices over SMTP |
| **Persistence** | Five PostgreSQL 17 clusters, one per bounded context, each behind a dedicated PgBouncer pooler |
| **Observability** | Prometheus + Grafana, Loki + Promtail, Jaeger + OpenTelemetry, Pyroscope continuous profiling, Node/Kafka/Postgres exporters, Alertmanager |
| **Deployment** | Docker Compose (full stack and infra-only), Kubernetes manifests with HPA, ArgoCD GitOps |

---

## Architecture Overview

Each service is a logical, decoupled Go binary inside `service/` with its own gRPC boundary. The
API Gateway is the only public edge: it authenticates the JWT, then translates REST/JSON into
downstream gRPC calls.

### Core Architecture Principles

- **Domain-Driven Boundary Isolation** — every service owns its database, cache, and logic. Cross-boundary database sharing is forbidden.
- **Database-per-Bounded-Context** — five physically separate PostgreSQL 17 instances isolate each context's data.
- **Dedicated PgBouncer Pooling** — each cluster is fronted by its own pooler (host ports `6432`–`6436`), so concurrent services cannot exhaust PostgreSQL sockets.
- **Clean Architecture** — `handler → service → repository`, wired in `apps/server.go`.
- **OLAP & CQRS** — OLTP writes go to the per-context PostgreSQL clusters; analytical reads go to ClickHouse via `stats_reader`.
- **Event-Driven Resilience** — Kafka decouples email delivery and stats materialization from the transactional path.
- **OTel Telemetry Integration** — trace IDs propagate from the REST edge through gRPC down to PostgreSQL and ClickHouse.

```mermaid
graph TB
    classDef client fill:#0f172a,stroke:#38bdf8,color:#e0f2fe,stroke-width:2px,font-weight:bold
    classDef gateway fill:#1e293b,stroke:#22d3ee,color:#cffafe,stroke-width:2px,font-weight:bold
    classDef domain fill:#1e1b4b,stroke:#818cf8,color:#e0e7ff,stroke-width:1.5px
    classDef infra fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1.5px
    classDef obs fill:#052e16,stroke:#4ade80,color:#dcfce7,stroke-width:1.5px
    classDef event fill:#431407,stroke:#fb923c,color:#fed7aa,stroke-width:1.5px
    classDef olap fill:#1e293b,stroke:#a855f7,color:#f3e8ff,stroke-width:1.5px

    Client["Client Applications<br/>Web / Mobile / POS Terminal"]:::client

    subgraph APIGateway["API Gateway — NGINX + Echo"]
        direction LR
        REST["REST API Endpoints<br/>/api/*"]
        Swagger["Swagger UI<br/>/swagger/index.html"]
        AuthMW["JWT Auth<br/>Middleware"]
    end
    class APIGateway gateway

    Client --> APIGateway

    subgraph BusinessServices["Business Domain Services"]
        direction TB

        subgraph IdentityDomain["Identity & Access"]
            AUTH["Auth Service<br/>JWT & OTP verification"]
            USER["User Service<br/>Profile management"]
            ROLE["Role Service<br/>RBAC + UserRole RPC"]
        end

        subgraph MerchantDomain["Merchant Management"]
            MERCH["Merchant Service<br/>Onboarding & profiling"]
            CASHIER["Cashier Service<br/>Cashier accounts"]
        end

        subgraph CatalogDomain["Catalog"]
            CATEGORY["Category Service"]
            PRODUCT["Product Service<br/>Catalog & pricing"]
        end

        subgraph SalesDomain["Sales & Payments"]
            ORDER["Order Service<br/>Cart & ordering"]
            ORDERITEM["Order Item Service<br/>Line items"]
            TXN["Transaction Service<br/>Payment settlement"]
        end
    end
    class BusinessServices domain

    subgraph OLAPEngine["OLAP & Analytics Layer"]
        direction TB
        WRITER["Stats Writer<br/>Kafka consumer + backfill"]:::olap
        READER["Stats Reader<br/>gRPC query service"]:::olap
        CLICKHOUSE[("ClickHouse OLAP<br/>Analytics DB")]:::infra
    end

    APIGateway -->|"gRPC — OLTP"| BusinessServices
    APIGateway -->|"gRPC — OLAP"| READER

    subgraph Persistence["PostgreSQL Cluster Tier — 5 contexts"]
        direction LR
        subgraph IdentityPG["Identity"]
            PGB_ID["PgBouncer :6432"]:::infra
            PG_ID[("pos_identity")]:::infra
        end
        subgraph MerchantPG["Merchant"]
            PGB_ME["PgBouncer :6433"]:::infra
            PG_ME[("pos_merchant")]:::infra
        end
        subgraph CatalogPG["Catalog"]
            PGB_CA["PgBouncer :6434"]:::infra
            PG_CA[("pos_catalog")]:::infra
        end
        subgraph SalesPG["Sales"]
            PGB_SA["PgBouncer :6435"]:::infra
            PG_SA[("pos_sales")]:::infra
        end
        subgraph EmailPG["Email"]
            PGB_EM["PgBouncer :6436"]:::infra
            PG_EM[("pos_email")]:::infra
        end
    end

    PGB_ID --> PG_ID
    PGB_ME --> PG_ME
    PGB_CA --> PG_CA
    PGB_SA --> PG_SA
    PGB_EM --> PG_EM

    BusinessServices -->|"SQL via per-context PgBouncer"| Persistence

    REDIS[("Redis Cache<br/>11 isolated instances")]:::infra
    KAFKA[("Kafka<br/>Event bus")]:::event
    PYRO["Pyroscope<br/>Continuous profiler"]:::obs

    BusinessServices -->|"Cache / Invalidate"| REDIS
    BusinessServices -->|"Publish events"| KAFKA
    BusinessServices -.->|"Profiles"| PYRO

    subgraph EventConsumers["Event-Driven Consumers"]
        EMAIL["Email Service<br/>SMTP worker"]
    end
    class EventConsumers event

    KAFKA -->|"Consume"| EMAIL
    KAFKA -->|"Consume"| WRITER
    WRITER -->|"Batch insert"| CLICKHOUSE
    READER -->|"Aggregate queries"| CLICKHOUSE
    READER -->|"Cache stats"| REDIS

    subgraph Observability["Observability Stack"]
        direction LR
        PROM["Prometheus"]
        LOKI["Loki"]
        JAEGER["Jaeger"]
        GRAFANA["Grafana"]
        OTEL["OTel Collector"]
        PROMTAIL["Promtail"]
        NODEX["Node Exporter"]
        KAFKAX["Kafka Exporter"]
        PGX["Postgres Exporter"]
        ALERTMGR["Alertmanager"]
    end
    class Observability obs

    BusinessServices -.->|"/metrics"| PROM
    BusinessServices -.->|"OTLP"| OTEL
    READER -.->|"/metrics"| PROM
    WRITER -.->|"/metrics"| PROM
    OTEL -.-> JAEGER
    PROMTAIL -.-> LOKI
    NODEX -.-> PROM
    KAFKAX -.-> PROM
    PGX -.-> PROM
    PROM -.-> ALERTMGR
    PROM -.-> GRAFANA
    LOKI -.-> GRAFANA
    JAEGER -.-> GRAFANA
```

---

## Service Catalog

The workspace ships **11 domain services**, one REST API gateway, two OLAP workers, and two
operational jobs.

| # | Service | Bounded Context | gRPC | Metrics (HTTP) | Responsibility |
|---|---------|-----------------|------|----------------|----------------|
| 1 | `apigateway` | — | — | — (REST `:5000`) | REST/JSON edge, Swagger UI, JWT middleware, gRPC fan-out |
| 2 | `auth` | identity | `50051` | `8081` | Register, login, refresh, password reset, OTP |
| 3 | `role` | identity | `50052` | `8082` | Role CRUD + `UserRoleService` (assign / remove / find-by-user) |
| 4 | `user` | identity | `50053` | `8083` | User profiles, soft-delete/restore |
| 5 | `category` | catalog | `50054` | `8084` | Category CRUD, search, trash/restore |
| 6 | `cashier` | merchant | `50055` | `8085` | Cashier accounts per merchant |
| 7 | `merchant` | merchant | `50056` | `8086` | Merchant onboarding & profiling |
| 8 | `order_item` | sales | `50057` | `8087` | Order line items |
| 9 | `order` | sales | `50058` | `8088` | Cart checkout & order lifecycle |
| 10 | `product` | catalog | `50059` | `8089` | Product CRUD, stock, pricing |
| 11 | `transaction` | sales | `50060` | `8090` | Payment settlement & status |
| 12 | `email` | email | — | — | Kafka consumer → SMTP notifications |
| 13 | `stats_writer` | ClickHouse | — | — | Kafka consumer + backfill → ClickHouse |
| 14 | `stats_reader` | ClickHouse | — | — | gRPC analytics over ClickHouse |
| 15 | `migrate` | all | — | — | Migration runner across all five contexts |
| 16 | `seeder` | all | — | — | Development fixtures |

```mermaid
graph LR
    classDef svc fill:#1e1b4b,stroke:#a78bfa,color:#ede9fe,stroke-width:1px
    classDef gw fill:#1e293b,stroke:#22d3ee,color:#cffafe,stroke-width:2px,font-weight:bold
    classDef support fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1px
    classDef olap fill:#1e293b,stroke:#a855f7,color:#f3e8ff,stroke-width:1px

    API["API Gateway<br/>Echo + REST + Swagger"]:::gw

    subgraph Identity["Identity (3)"]
        A1["auth"]:::svc
        A2["user"]:::svc
        A3["role"]:::svc
    end

    subgraph Merchant["Merchant (2)"]
        M1["merchant"]:::svc
        M2["cashier"]:::svc
    end

    subgraph Catalog["Catalog (2)"]
        C1["category"]:::svc
        C2["product"]:::svc
    end

    subgraph Sales["Sales (3)"]
        S1["order"]:::svc
        S2["order_item"]:::svc
        S3["transaction"]:::svc
    end

    subgraph OLAP["OLAP (2)"]
        O1["stats_writer"]:::olap
        O2["stats_reader"]:::olap
    end

    subgraph Support["Support (3)"]
        T1["email"]:::support
        T2["migrate"]:::support
        T3["seeder"]:::support
    end

    API --> Identity
    API --> Merchant
    API --> Catalog
    API --> Sales
    API --> OLAP
```

---

## Database Layer — PostgreSQL Cluster & PgBouncer

The POS platform runs **five independent PostgreSQL 17 clusters** — one per bounded context.
"Cluster" means *one dedicated PostgreSQL instance per context*, not a replicated
primary/replica pair. Isolation is the point: cross-context schema coupling becomes impossible,
blast radius stays small, and each context can be migrated and tuned independently.

**Every cluster is fronted by its own PgBouncer pooler.** Services dial the pooler, never
PostgreSQL itself, so the 11 concurrent services cannot exhaust server-side connections.

### Topology

| Bounded Context | PostgreSQL instance | Database | PgBouncer (host port) | Owning services |
| :--- | :--- | :--- | :--- | :--- |
| **Identity** | `postgres_identity` | `pos_identity` | `6432` | `auth`, `user`, `role` |
| **Merchant** | `postgres_merchant` | `pos_merchant` | `6433` | `merchant`, `cashier` |
| **Catalog** | `postgres_catalog` | `pos_catalog` | `6434` | `category`, `product` |
| **Sales** | `postgres_sales` | `pos_sales` | `6435` | `order`, `order_item`, `transaction` |
| **Email** | `postgres_email` | `pos_email` | `6436` | `email` |

> **Pool mode differs per environment** — this is intentional and worth knowing when debugging
> prepared-statement or session-state behaviour:
> - **Local (Docker Compose)**: `POOL_MODE=transaction`
> - **Kubernetes**: `POOL_MODE=session`, with `MAX_CLIENT_CONN=1000`, `DEFAULT_POOL_SIZE=20`,
>   `MAX_PREPARED_STATEMENTS=100`
>
> Auth is `scram-sha-256` in both.

```mermaid
graph TB
    classDef svc fill:#1e1b4b,stroke:#a78bfa,color:#ede9fe,stroke-width:1px
    classDef pool fill:#1e293b,stroke:#22d3ee,color:#cffafe,stroke-width:2px,font-weight:bold
    classDef pg fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1.5px
    classDef obs fill:#052e16,stroke:#4ade80,color:#dcfce7,stroke-width:1px

    subgraph IdentityCtx["Identity Context"]
        direction TB
        S_ID["auth · user · role"]:::svc
        PGB_ID["pgbouncer_identity :6432<br/>transaction pool locally · session in K8s<br/>scram-sha-256"]:::pool
        PG_ID[("postgres_identity<br/>pos_identity")]:::pg
        PX_ID["postgres-exporter"]:::obs
        S_ID -->|"DB_IDENTITY_HOST/PORT/NAME"| PGB_ID
        PGB_ID -->|"bounded server pool"| PG_ID
        PG_ID -.-> PX_ID
    end

    subgraph MerchantCtx["Merchant Context"]
        direction TB
        S_ME["merchant · cashier"]:::svc
        PGB_ME["pgbouncer_merchant :6433"]:::pool
        PG_ME[("postgres_merchant<br/>pos_merchant")]:::pg
        PX_ME["postgres-exporter"]:::obs
        S_ME -->|"DB_MERCHANT_*"| PGB_ME
        PGB_ME --> PG_ME
        PG_ME -.-> PX_ME
    end

    subgraph CatalogCtx["Catalog Context"]
        direction TB
        S_CA["category · product"]:::svc
        PGB_CA["pgbouncer_catalog :6434"]:::pool
        PG_CA[("postgres_catalog<br/>pos_catalog")]:::pg
        PX_CA["postgres-exporter"]:::obs
        S_CA -->|"DB_CATALOG_*"| PGB_CA
        PGB_CA --> PG_CA
        PG_CA -.-> PX_CA
    end

    subgraph SalesCtx["Sales Context"]
        direction TB
        S_SA["order · order_item · transaction"]:::svc
        PGB_SA["pgbouncer_sales :6435"]:::pool
        PG_SA[("postgres_sales<br/>pos_sales")]:::pg
        PX_SA["postgres-exporter"]:::obs
        S_SA -->|"DB_SALES_*"| PGB_SA
        PGB_SA --> PG_SA
        PG_SA -.-> PX_SA
    end

    subgraph EmailCtx["Email Context"]
        direction TB
        S_EM["email"]:::svc
        PGB_EM["pgbouncer_email :6436"]:::pool
        PG_EM[("postgres_email<br/>pos_email")]:::pg
        PX_EM["postgres-exporter"]:::obs
        S_EM -->|"DB_EMAIL_*"| PGB_EM
        PGB_EM --> PG_EM
        PG_EM -.-> PX_EM
    end
```

### Connection Lifecycle

```mermaid
sequenceDiagram
    autonumber
    participant SVC as Domain Service<br/>GORM client
    participant PGB as PgBouncer<br/>per-context pooler
    participant PG as PostgreSQL<br/>per-context instance
    participant PX as postgres-exporter

    SVC->>PGB: Dial DB_<CONTEXT>_HOST:PORT<br/>scram-sha-256 auth
    PGB->>PGB: Admit client within max_client_conn
    alt Server slot available
        PGB->>PG: Assign pooled server connection
    else Pool saturated
        PGB-->>SVC: Queue until a slot frees
    end
    SVC->>PGB: BEGIN / SELECT / INSERT / COMMIT
    PGB->>PG: Forward statement on assigned connection
    PG-->>PGB: Result set
    PGB-->>SVC: Rows
    SVC->>PGB: Close client connection
    PGB->>PG: Return server connection to the pool
    PX-->>PX: Scrape pg_stat_database / pg_stat_activity
```

### Environment Contract

There is no generic `DB_HOST` / `DB_PORT` / `DB_NAME`. Each service declares its cluster once in
`service/<name>/cmd/main.go`:

```go
server.Config{
    DBCluster: "DB_IDENTITY", // → DB_IDENTITY_HOST / _PORT / _NAME
    // ...
}
```

`pkg/database/names.go` is the single source of truth and holds three coordinated maps:

| Symbol | Purpose |
| :--- | :--- |
| `IdentityDB` … `EmailDB` | Logical database names (`pos_identity` … `pos_email`) |
| `BoundedContexts` | Migration/seed order: `identity → merchant → catalog → sales`, `email` independent |
| `ContextPrefix` | `"identity" → "DB_IDENTITY"`, `"merchant" → "DB_MERCHANT"`, `"catalog" → "DB_CATALOG"`, `"sales" → "DB_SALES"`, `"email" → "DB_EMAIL"` |
| `ServiceContext` | `"cashier" → "merchant"`, `"transaction" → "sales"`, … — routes migrations and seed data to the right instance |

Each prefix resolves `<PREFIX>_HOST` / `<PREFIX>_PORT` / `<PREFIX>_NAME` (plus optional per-context
user/password), and those values point at the context's **PgBouncer** service — never at
PostgreSQL directly.

### Migration & Seed Order

```mermaid
flowchart LR
    classDef job fill:#431407,stroke:#fb923c,color:#fed7aa,stroke-width:1.5px
    classDef pool fill:#1e293b,stroke:#22d3ee,color:#cffafe,stroke-width:2px
    classDef pg fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1.5px

    MIG["service/migrate<br/>per-context migrations"]:::job
    SEED["service/seeder<br/>fixtures"]:::job

    MIG -->|"1 identity"| P1["pgbouncer_identity"]:::pool
    MIG -->|"2 merchant"| P2["pgbouncer_merchant"]:::pool
    MIG -->|"3 catalog"| P3["pgbouncer_catalog"]:::pool
    MIG -->|"4 sales"| P4["pgbouncer_sales"]:::pool
    MIG -->|"independent"| P5["pgbouncer_email"]:::pool

    SEED --> P1
    SEED --> P2
    SEED --> P3
    SEED --> P4

    P1 --> D1[("pos_identity")]:::pg
    P2 --> D2[("pos_merchant")]:::pg
    P3 --> D3[("pos_catalog")]:::pg
    P4 --> D4[("pos_sales")]:::pg
    P5 --> D5[("pos_email")]:::pg
```

### Kubernetes Topology

In-cluster each context becomes a **StatefulSet + PVC** with a **Deployment** pooler in front, and
NetworkPolicies enforce who may talk to what:

```mermaid
flowchart TB
    classDef pod fill:#1e1b4b,stroke:#a78bfa,color:#ede9fe,stroke-width:1.5px
    classDef pool fill:#1e293b,stroke:#22d3ee,color:#cffafe,stroke-width:2px
    classDef pg fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1.5px
    classDef np fill:#3b0764,stroke:#c084fc,color:#f3e8ff,stroke-width:1px,font-style:italic

    subgraph NS["namespace: point-of-sale"]
        subgraph IdentityK8s["Identity"]
            PK_ID["Deployment pgbouncer-identity"]:::pool
            PS_ID[("StatefulSet postgres-identity<br/>+ PVC")]:::pg
        end
        subgraph MerchantK8s["Merchant"]
            PK_ME["Deployment pgbouncer-merchant"]:::pool
            PS_ME[("StatefulSet postgres-merchant<br/>+ PVC")]:::pg
        end
        subgraph CatalogK8s["Catalog"]
            PK_CA["Deployment pgbouncer-catalog"]:::pool
            PS_CA[("StatefulSet postgres-catalog<br/>+ PVC")]:::pg
        end
        subgraph SalesK8s["Sales"]
            PK_SA["Deployment pgbouncer-sales"]:::pool
            PS_SA[("StatefulSet postgres-sales<br/>+ PVC")]:::pg
        end
        subgraph EmailK8s["Email"]
            PK_EM["Deployment pgbouncer-email"]:::pool
            PS_EM[("StatefulSet postgres-email<br/>+ PVC")]:::pg
        end

        NPP["NetworkPolicy postgres-network-policy<br/>poolers + exporters only"]:::np
        NPB["NetworkPolicy pgbouncer-network-policy<br/>owning services only"]:::np
        NPM["NetworkPolicy migrate-network-policy<br/>migration job only"]:::np
        MIGJOB["Job migrate-job"]:::pod
    end

    PK_ID --> PS_ID
    PK_ME --> PS_ME
    PK_CA --> PS_CA
    PK_SA --> PS_SA
    PK_EM --> PS_EM

    NPP -.-> PS_ID
    NPB -.-> PK_ID
    NPM -.-> PK_ID
    MIGJOB --> PK_ID
    MIGJOB --> PK_ME
    MIGJOB --> PK_CA
    MIGJOB --> PK_SA
    MIGJOB --> PK_EM
```

> PostgreSQL ports are not published to the host in either target — connect through PgBouncer
> (`6432`–`6436`).

---

## Internal Service Architecture

```mermaid
graph TB
    classDef handler fill:#1e3a5f,stroke:#7dd3fc,color:#e0f2fe,stroke-width:1.5px
    classDef service fill:#1e1b4b,stroke:#a78bfa,color:#ede9fe,stroke-width:1.5px
    classDef repo fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1.5px
    classDef infra fill:#052e16,stroke:#4ade80,color:#dcfce7,stroke-width:1.5px
    classDef shared fill:#431407,stroke:#fb923c,color:#fed7aa,stroke-width:1.5px

    subgraph Service["service/<name>/"]
        direction TB

        CMD["cmd/main.go<br/>Entry point + DBCluster selection"]

        subgraph Internal["internal wiring"]
            direction TB
            APPS["apps/server.go<br/>Dependency injection"]:::handler
            HANDLER["handler/<br/>gRPC handlers"]:::handler
            MW["middleware/<br/>Interceptors"]:::handler
            SVC["service/<br/>Business logic"]:::service
            CACHE["cache/<br/>Redis cache layer"]:::service
            REPO["repository/<br/>Data access — GORM"]:::repo
        end

        CMD --> APPS
        APPS --> HANDLER
        APPS --> SVC
        APPS --> CACHE
        APPS --> REPO
        HANDLER --> SVC
        SVC --> REPO
        SVC --> CACHE
    end

    subgraph SharedLibs["shared/ — Shared Libraries"]
        direction LR
        DOMAIN["domain/<br/>record / request / response"]:::shared
        OBS["observability/<br/>cache & tracing metrics"]:::shared
        CACHESHARED["cache/<br/>redis_cache.go"]:::shared
        MAPPER["mapper/<br/>Domain ↔ Proto"]:::shared
        ERRORS["errors/ + errorhandler/"]:::shared
    end

    subgraph PkgLibs["pkg/ — Platform Libraries"]
        direction LR
        PKGAUTH["auth/<br/>JWT manager"]:::infra
        PKGKAFKA["kafka/<br/>Producer / consumer"]:::infra
        PKGOUTBOX["outbox/<br/>Transactional outbox helpers"]:::infra
        PKGOTEL["otel/<br/>Tracing + metrics init"]:::infra
        PKGRES["resilience/<br/>Circuit breaker, rate limiter,<br/>load monitor, DependencyGuard"]:::infra
        PKGLOG["logger/<br/>Zap structured logging"]:::infra
        PKGSRV["server/<br/>gRPC bootstrap"]:::infra
        PKGDB["database/<br/>Per-context GORM + names.go"]:::infra
        PKGCH["clickhouse/<br/>OLAP connection + schema"]:::infra
        PKGADAPTER["adapter/role · adapter/user_role<br/>guarded gRPC clients"]:::infra
    end

    PB["pb/<br/>Generated protobuf Go code"]:::shared
    PGB_EXT["PgBouncer per context"]:::infra

    REPO --> DOMAIN
    REPO --> PGB_EXT
    SVC --> DOMAIN
    SVC --> OBS
    HANDLER --> PB
    HANDLER --> MAPPER
    APPS --> PKGSRV
    APPS --> PKGOTEL
    APPS --> CACHESHARED
    APPS --> PKGADAPTER
    APPS --> OBS
```

### Cross-Context Access: the Adapter Pattern

`service/auth` and `service/user` need role data but must not open a second database connection.
They use typed gRPC adapters, each wrapped in a `DependencyGuard` (per-call timeout + circuit
breaker + bulkhead):

| Adapter | Package | Backing RPC | Used by |
| :--- | :--- | :--- | :--- |
| `RoleAdapter` | `pkg/adapter/role` | `RoleService` — `FindById`, `FindByName` | `service/auth`, `service/user` |
| `UserRoleAdapter` | `pkg/adapter/user_role` | `UserRoleService` — `FindByUserId`, `AssignRoleToUser`, `RemoveRoleFromUser` | `service/auth`, `service/apigateway` |

---

## Data & Event Flow

### Synchronous Flow — REST → gRPC → Redis → PgBouncer → PostgreSQL

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant GW as API Gateway<br/>Echo + REST
    participant SVC as Domain Service<br/>gRPC server
    participant REDIS as Redis
    participant PGB as PgBouncer<br/>per-context
    participant DB as PostgreSQL<br/>per-context

    C->>GW: REST HTTP request GET/POST/PUT/DELETE
    GW->>GW: JWT authentication check
    GW->>SVC: gRPC call with protobuf payload
    SVC->>REDIS: Check cache
    alt Cache hit
        REDIS-->>SVC: Cached response
    else Cache miss
        SVC->>PGB: Acquire pooled connection
        PGB->>DB: Execute SQL via GORM
        DB-->>PGB: Result set
        PGB-->>SVC: Rows
        SVC->>REDIS: Populate cache for next read
    end
    SVC-->>GW: gRPC response payload
    GW-->>C: REST HTTP response JSON
```

### Asynchronous Flow — Kafka Notification & Stats Pipeline

```mermaid
sequenceDiagram
    autonumber
    participant SVC as Order / Transaction Service
    participant K as Kafka Broker
    participant EMAIL as Email Worker
    participant SMTP as SMTP Server
    participant WRITER as Stats Writer
    participant CH as ClickHouse

    SVC->>K: Publish event order.created / transaction.paid
    K-->>EMAIL: Deliver topic payload
    EMAIL->>EMAIL: Map payload details
    EMAIL->>SMTP: Send styled notification
    SMTP-->>EMAIL: Delivery confirmation
    K-->>WRITER: Deliver stats event
    WRITER->>CH: Batch insert into analytics tables
    CH-->>WRITER: Batch flushed
```

---

## OLAP Analytics Layer

Transactional writes land in the per-context PostgreSQL clusters; analytical reads are served by
ClickHouse through `stats_reader`, which the API Gateway calls with cache-aside Redis.

```mermaid
graph LR
    classDef olap fill:#1e293b,stroke:#a855f7,color:#f3e8ff,stroke-width:1.5px
    classDef store fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1.5px
    classDef api fill:#1e293b,stroke:#22d3ee,color:#cffafe,stroke-width:2px,font-weight:bold
    classDef bus fill:#431407,stroke:#fb923c,color:#fed7aa,stroke-width:1.5px

    TXN["order · transaction<br/>product · category · cashier"]:::olap
    KAFKA[("Kafka<br/>domain events")]:::bus
    WRITER["stats_writer<br/>consumer + backfill"]:::olap
    CH[("ClickHouse<br/>columnar warehouse")]:::store
    READER["stats_reader<br/>gRPC analytics"]:::olap
    GW["API Gateway<br/>/api/stats/*"]:::api
    REDIS[("Redis<br/>stats cache")]:::store

    TXN -->|"publish"| KAFKA
    KAFKA -->|"consume"| WRITER
    WRITER -->|"batch insert"| CH
    GW -->|"gRPC"| READER
    READER -->|"aggregate"| CH
    READER -->|"cache-aside"| REDIS
```

---

## Observability Architecture

```mermaid
graph TB
    classDef service fill:#1e1b4b,stroke:#818cf8,color:#e0e7ff,stroke-width:1.5px
    classDef collector fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1.5px
    classDef storage fill:#052e16,stroke:#4ade80,color:#dcfce7,stroke-width:1.5px
    classDef viz fill:#431407,stroke:#fb923c,color:#fed7aa,stroke-width:2px,font-weight:bold

    subgraph Sources["Telemetry Sources"]
        direction TB
        SVCS["11 domain services<br/>+ apigateway + stats_reader/writer"]:::service
        KAFKA_SRC["Kafka broker"]:::service
        PG_SRC["5 PostgreSQL clusters<br/>+ 5 PgBouncer poolers"]:::service
        NODES["Host / node"]:::service
    end

    subgraph Collectors["Collection Layer"]
        direction TB
        PROM["Prometheus<br/>scrapes /metrics"]:::collector
        PROMTAIL["Promtail<br/>ships container logs"]:::collector
        OTEL["OTel Collector<br/>receives OTLP spans"]:::collector
        NODEX["Node Exporter"]:::collector
        KAFKAX["Kafka Exporter"]:::collector
        PGX["Postgres Exporter"]:::collector
        PYRO["Pyroscope<br/>continuous profiling"]:::collector
    end

    subgraph Storage["Storage Layer"]
        direction TB
        PROM_TSDB["Prometheus TSDB"]:::storage
        LOKI_STORE["Loki chunks"]:::storage
        JAEGER_STORE["Jaeger trace store"]:::storage
        PYRO_STORE["Pyroscope profiles"]:::storage
    end

    subgraph Visualization["Visualization & Alerting"]
        GRAFANA["Grafana<br/>unified dashboards"]:::viz
        ALERTMGR["Alertmanager<br/>alert routing"]:::viz
    end

    SVCS -->|"/metrics"| PROM
    SVCS -->|"OTLP gRPC"| OTEL
    SVCS -->|"profiles"| PYRO
    SVCS -->|"stdout / stderr"| PROMTAIL
    NODES --> NODEX
    KAFKA_SRC --> KAFKAX
    PG_SRC --> PGX

    NODEX --> PROM
    KAFKAX --> PROM
    PGX --> PROM
    PROM --> PROM_TSDB
    PROMTAIL --> LOKI_STORE
    OTEL --> JAEGER_STORE
    PYRO --> PYRO_STORE

    PROM_TSDB --> GRAFANA
    LOKI_STORE --> GRAFANA
    JAEGER_STORE --> GRAFANA
    PYRO_STORE --> GRAFANA
    PROM_TSDB --> ALERTMGR
```

| Pillar | Tooling | What you get |
| :--- | :--- | :--- |
| **Metrics** | Prometheus + Grafana | CPU/memory, request error rates, gRPC latencies, database connection states |
| **Logging** | Loki + Promtail | Structured JSON indexed per service, queryable with LogQL |
| **Tracing** | OpenTelemetry + Jaeger | End-to-end traces across REST → gRPC → PostgreSQL / ClickHouse / Kafka |
| **Profiling** | Pyroscope | Continuous CPU/memory profiling to catch allocation leaks in transaction loops |
| **Alerting** | Alertmanager | Notifications on latency spikes and service disconnects |

---

## Deployment Architectures

### Docker Compose (Local Development)

Two compose files live under `deployments/local/`:

- **`docker-compose.yml`** — the whole platform: 5 PostgreSQL + 5 PgBouncer, 11 isolated Redis
  instances, ClickHouse, Kafka, Pyroscope, all Go services, and the observability stack.
- **`docker-compose.infra.yml`** — **infra-only** for native development: it boots just the
  5 PostgreSQL/PgBouncer pairs, Redis, Kafka, ClickHouse, and observability, publishing PgBouncer
  on host ports `6432`–`6436` so host-run binaries can reach their context.

```mermaid
flowchart TD
    classDef gateway fill:#1e293b,stroke:#22d3ee,color:#cffafe,stroke-width:2px,font-weight:bold
    classDef core fill:#1e1b4b,stroke:#a78bfa,color:#ede9fe,stroke-width:1.5px
    classDef infra fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1.5px
    classDef obs fill:#052e16,stroke:#4ade80,color:#dcfce7,stroke-width:1.5px
    classDef event fill:#431407,stroke:#fb923c,color:#fed7aa,stroke-width:1.5px
    classDef olap fill:#1e293b,stroke:#a855f7,color:#f3e8ff,stroke-width:1.5px

    subgraph DockerCompose["docker-compose.yml — Local Environment"]

        subgraph Gateway["Edge"]
            NGINX["NGINX Proxy :80"]
            APIGW["API Gateway :5000<br/>Echo + REST + Swagger"]:::gateway
        end

        subgraph Services["Domain Service Containers"]
            direction TB
            subgraph IdSvc["Identity"]
                AUTH["auth-service"]
                ROLE["role-service"]
                USER["user-service"]
            end
            subgraph MeSvc["Merchant"]
                MERCH["merchant-service"]
                CASHIER["cashier-service"]
            end
            subgraph CaSvc["Catalog"]
                CATEGORY["category-service"]
                PRODUCT["product-service"]
            end
            subgraph SaSvc["Sales"]
                ORDER["order-service"]
                ORDERITEM["order_item-service"]
                TXN["transaction-service"]
            end
        end
        class Services core

        subgraph Infra["Infrastructure Suite"]
            direction TB
            subgraph PGTier["5 × PostgreSQL 17 — each behind its own PgBouncer"]
                PG1[("identity :6432")]
                PG2[("merchant :6433")]
                PG3[("catalog :6434")]
                PG4[("sales :6435")]
                PG5[("email :6436")]
            end
            REDIS_APIGW[("redis-apigateway :6379")]
            REDIS_AUTH[("redis-auth :6380")]
            REDIS_ROLE[("redis-role :6381")]
            REDIS_USER[("redis-user :6382")]
            REDIS_CATEGORY[("redis-category :6383")]
            REDIS_CASHIER[("redis-cashier :6384")]
            REDIS_MERCH[("redis-merchant :6385")]
            REDIS_ORDERITEM[("redis-orderitem :6386")]
            REDIS_ORDER[("redis-order :6387")]
            REDIS_PRODUCT[("redis-product :6388")]
            REDIS_TXN[("redis-transaction :6389")]
            KAFKA[("Kafka Broker :9092")]:::event
            CH[("ClickHouse OLAP :9000/:8123")]:::infra
            PYRO[("Pyroscope :4040")]:::obs
        end

        subgraph Obs["Observability"]
            PROM["Prometheus :9090"]
            GRAFANA["Grafana :3000"]
            LOKI["Loki :3100"]
            JAEGER["Jaeger :16686"]
            OTEL["OTel Collector :4317"]
            NODEX["Node Exporter"]
            KAFKAX["Kafka Exporter :9308"]
            PGX["Postgres Exporter"]
            ALERTMGR["Alertmanager :9093"]
            PROMTAIL["Promtail"]
        end
        class Obs obs

        subgraph Events["Event Consumers & OLAP"]
            EMAIL["Email Worker"]:::event
            WRITER["Stats Writer"]:::olap
            READER["Stats Reader"]:::olap
        end
    end

    NGINX --> APIGW
    APIGW -->|"gRPC"| Services
    APIGW -->|"gRPC"| READER
    Services -->|"SQL via PgBouncer"| PGTier
    Services --> KAFKA
    KAFKA --> EMAIL
    KAFKA --> WRITER
    WRITER --> CH
    READER --> CH

    AUTH --> REDIS_AUTH
    ROLE --> REDIS_ROLE
    USER --> REDIS_USER
    CATEGORY --> REDIS_CATEGORY
    CASHIER --> REDIS_CASHIER
    MERCH --> REDIS_MERCH
    ORDER --> REDIS_ORDER
    ORDERITEM --> REDIS_ORDERITEM
    PRODUCT --> REDIS_PRODUCT
    TXN --> REDIS_TXN
    APIGW --> REDIS_APIGW

    Services -.->|"/metrics"| PROM
    Services -.->|"OTLP"| OTEL
    Services -.->|"profiles"| PYRO
    OTEL -.-> JAEGER
    PROMTAIL -.-> LOKI
    PROM -.-> GRAFANA
    PROM -.-> ALERTMGR
    LOKI -.-> GRAFANA
```

### Kubernetes (Production) + ArgoCD GitOps

Production runs in the `point-of-sale` namespace: one Deployment + Service + HPA per domain
service, one StatefulSet + PVC per PostgreSQL context, one Deployment per PgBouncer pooler, and
NetworkPolicies that enforce "only the owning context may connect". Delivery is GitOps-driven via
ArgoCD (`deployments/gitops/argocd/`) with base/overlay manifests under
`deployments/kubernetes/overlays/`.

```mermaid
flowchart TD
    classDef k8s fill:#0c1222,stroke:#38bdf8,color:#e0f2fe,stroke-width:2px,font-weight:bold
    classDef pod fill:#1e1b4b,stroke:#a78bfa,color:#ede9fe,stroke-width:1.5px
    classDef hpa fill:#3b0764,stroke:#c084fc,color:#f3e8ff,stroke-width:1px,font-style:italic
    classDef infra fill:#172554,stroke:#60a5fa,color:#dbeafe,stroke-width:1.5px
    classDef obs fill:#052e16,stroke:#4ade80,color:#dcfce7,stroke-width:1.5px
    classDef job fill:#431407,stroke:#fb923c,color:#fed7aa,stroke-width:1.5px
    classDef olap fill:#1e293b,stroke:#a855f7,color:#f3e8ff,stroke-width:1.5px

    ARGO["ArgoCD<br/>GitOps controller"]:::k8s
    REPO[("Git repository<br/>deployments/kubernetes")]:::k8s

    subgraph K8S["Kubernetes Cluster — namespace: point-of-sale"]

        subgraph Ingress["Ingress"]
            NGINX["NGINX Ingress + TLS"]:::k8s
            APIGW["API Gateway Pod"]:::pod
        end

        subgraph CorePods["Domain Service Pods + HPAs"]
            direction TB
            subgraph IdentityPods["Identity"]
                AUTH["auth-pod"]:::pod
                USER["user-pod"]:::pod
                ROLE["role-pod"]:::pod
                AUTH_HPA["↕ HPA"]:::hpa
                USER_HPA["↕ HPA"]:::hpa
            end
            subgraph MerchantPods["Merchant"]
                MERCH["merchant-pod"]:::pod
                CASHIER["cashier-pod"]:::pod
            end
            subgraph CatalogPods["Catalog"]
                CATEGORY["category-pod"]:::pod
                PRODUCT["product-pod"]:::pod
            end
            subgraph SalesPods["Sales"]
                ORDER["order-pod"]:::pod
                ORDERITEM["order_item-pod"]:::pod
                TXN["transaction-pod"]:::pod
            end
            subgraph OLAPPods["OLAP"]
                SWRITER["stats-writer-pod"]:::olap
                SREADER["stats-reader-pod"]:::olap
            end
        end

        subgraph DataPods["PostgreSQL Clusters + PgBouncer"]
            direction TB
            PGB_ID["pgbouncer-identity"]:::infra
            PG_ID[("postgres-identity + PVC")]:::infra
            PGB_ME["pgbouncer-merchant"]:::infra
            PG_ME[("postgres-merchant + PVC")]:::infra
            PGB_CA["pgbouncer-catalog"]:::infra
            PG_CA[("postgres-catalog + PVC")]:::infra
            PGB_SA["pgbouncer-sales"]:::infra
            PG_SA[("postgres-sales + PVC")]:::infra
            PGB_EM["pgbouncer-email"]:::infra
            PG_EM[("postgres-email + PVC")]:::infra
        end

        REDIS_CLUSTER[("Redis Cluster + PVC")]:::infra
        KAFKA[("Kafka StatefulSet")]:::infra
        CLICKHOUSE[("ClickHouse + PVC")]:::infra

        subgraph ObsPods["Observability"]
            PROM["Prometheus"]:::obs
            GRAFANA["Grafana"]:::obs
            LOKI["Loki + PVC"]:::obs
            PROMTAIL["Promtail DaemonSet"]:::obs
            JAEGER["Jaeger"]:::obs
            OTEL["OTel Collector"]:::obs
            NODEX["Node Exporter DaemonSet"]:::obs
            ALERTMGR["Alertmanager"]:::obs
            PYRO["Pyroscope"]:::obs
        end

        subgraph Jobs["Jobs & Workers"]
            MIGRATE["migrate-job"]:::job
            SEEDER["seeder-job"]:::job
            EMAILJ["email worker pod"]:::job
        end
    end

    REPO --> ARGO
    ARGO --> K8S

    NGINX --> APIGW
    APIGW -->|"gRPC"| CorePods
    APIGW -->|"gRPC"| SREADER
    CorePods --> DataPods
    CorePods --> REDIS_CLUSTER
    CorePods --> KAFKA
    KAFKA --> EMAILJ
    KAFKA --> SWRITER
    SWRITER --> CLICKHOUSE
    SREADER --> CLICKHOUSE

    PGB_ID --> PG_ID
    PGB_ME --> PG_ME
    PGB_CA --> PG_CA
    PGB_SA --> PG_SA
    PGB_EM --> PG_EM

    CorePods -.->|"/metrics"| PROM
    CorePods -.->|"OTLP"| OTEL
    OTEL -.-> JAEGER
    PROMTAIL -.-> LOKI
    PROM -.-> GRAFANA
    PROM -.-> ALERTMGR
    MIGRATE --> DataPods
```

---

## Technology Stack

| Category | Technology | Purpose |
| :--- | :--- | :--- |
| **Language** | Go (Golang) | High-performance compiled concurrent backend |
| **API Edge Gateway** | Echo (REST) | REST API gateway with auto-generated Swagger UI |
| **RPC Inter-service** | gRPC + Protobuf | Contract-first synchronous communication |
| **OLTP Database** | PostgreSQL 17 ×5 | Database-per-bounded-context: `pos_identity`, `pos_merchant`, `pos_catalog`, `pos_sales`, `pos_email` |
| **Connection Pooler** | PgBouncer ×5 | One pooler per cluster — `transaction` mode locally, `session` in K8s — host ports `6432`–`6436` |
| **OLAP Database** | ClickHouse | Columnar warehouse for analytics aggregations |
| **ORM** | GORM | Object-relational mapping & query builder |
| **DB Migrations** | Goose | Per-context incremental schema versioning |
| **Caching Tier** | Redis | 11 isolated Redis instances, one per service |
| **Messaging Stream** | Apache Kafka | Asynchronous high-throughput event bus |
| **Token Manager** | JWT | Stateless authentication & authorization |
| **Observability** | OpenTelemetry + Jaeger | Vendor-neutral telemetry pipeline and visualization |
| **Metrics / Dashboards** | Prometheus + Grafana | Scraping, dashboards, and alert rules |
| **Log Aggregation** | Loki + Promtail | Centralized structured log storage and shipping |
| **Continuous Profiler** | Pyroscope | Real-time CPU/memory profiling |
| **Alerting** | Alertmanager | Alert routing & notification dispatch |
| **Reverse Proxy** | NGINX | Edge routing and TLS termination |
| **Containerization** | Docker + Docker Compose | Local orchestration |
| **Orchestrator** | Kubernetes + HPA | Production auto-scaling pod infrastructure |
| **GitOps** | ArgoCD | Declarative continuous delivery |
| **Resilience** | `pkg/resilience` | Circuit breaker, rate limiter, load monitor, `DependencyGuard` |

---

## Getting Started

### Prerequisites

- [Git](https://git-scm.com/)
- [Go](https://go.dev/) (v1.23+)
- [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/)
- [Just Task Runner](https://github.com/casey/just)
- [Protobuf Compiler](https://grpc.io/docs/protoc-installation/) (for codegen updates)

### 1. Clone the Workspace

```sh
git clone https://github.com/MamangRust/monolith-pointofsale-grpc.git
cd monolith-pointofsale-grpc
```

### 2. Prepare Environment Configurations

Environment files are tracked in the repository — no copy step required:

- `.env` — root variables for native (non-containerized) service runs
- `deployments/local/docker.env` — variables injected into Compose containers

The PostgreSQL cluster contract lives there:

```dotenv
# Host = the context's PgBouncer service, never PostgreSQL itself
DB_IDENTITY_HOST=pgbouncer_identity
DB_IDENTITY_PORT=5432
DB_IDENTITY_NAME=pos_identity

DB_SALES_HOST=pgbouncer_sales
DB_SALES_PORT=5432
DB_SALES_NAME=pos_sales
```

For native runs against `docker-compose.infra.yml`, point hosts at `localhost` and ports at the
published pooler ports `6432`–`6436`.

### 3. Start Local Environment

```sh
just build-up      # build images and boot the Compose stack
just migrate       # run migrations across all five contexts
just seeder        # optional: development fixtures
just ps            # verify health and ports
```

**Native development** (services from `bin/`, only infrastructure containerized):

```sh
just infra-up      # 5 PostgreSQL/PgBouncer pairs, Redis, Kafka, ClickHouse, observability
just services-up   # build binaries and start services locally
```

### 4. Access Services

| Service | URL |
| :--- | :--- |
| Swagger UI | [http://localhost/swagger/index.html](http://localhost/swagger/index.html) |
| REST API Gateway Edge | [http://localhost/api/*](http://localhost/api/*) |
| API Gateway Direct | [http://localhost:5000](http://localhost:5000) |
| Grafana | [http://localhost:3000](http://localhost:3000) (`admin`/`admin`) |
| Prometheus | [http://localhost:9090](http://localhost:9090) |
| Jaeger | [http://localhost:16686](http://localhost:16686) |
| Pyroscope | [http://localhost:4040](http://localhost:4040) |
| Loki | [http://localhost:3100](http://localhost:3100) |

```sh
just down
```

---

## Port Map Registry

| Application / Service | Port / URL |
| :--- | :--- |
| **auth** gRPC / metrics | `localhost:50051` / `localhost:8081` |
| **role** gRPC / metrics | `localhost:50052` / `localhost:8082` |
| **user** gRPC / metrics | `localhost:50053` / `localhost:8083` |
| **category** gRPC / metrics | `localhost:50054` / `localhost:8084` |
| **cashier** gRPC / metrics | `localhost:50055` / `localhost:8085` |
| **merchant** gRPC / metrics | `localhost:50056` / `localhost:8086` |
| **order_item** gRPC / metrics | `localhost:50057` / `localhost:8087` |
| **order** gRPC / metrics | `localhost:50058` / `localhost:8088` |
| **product** gRPC / metrics | `localhost:50059` / `localhost:8089` |
| **transaction** gRPC / metrics | `localhost:50060` / `localhost:8090` |
| **PgBouncer — identity** (`pos_identity`) | `localhost:6432` |
| **PgBouncer — merchant** (`pos_merchant`) | `localhost:6433` |
| **PgBouncer — catalog** (`pos_catalog`) | `localhost:6434` |
| **PgBouncer — sales** (`pos_sales`) | `localhost:6435` |
| **PgBouncer — email** (`pos_email`) | `localhost:6436` |
| **Redis — apigateway / auth / role / user** | `6379` / `6380` / `6381` / `6382` |
| **Redis — category / cashier / merchant** | `6383` / `6384` / `6385` |
| **Redis — orderitem / order / product / transaction** | `6386` / `6387` / `6388` / `6389` |
| **Kafka Broker** | `localhost:9092` (internal) / `localhost:29092` (host) |
| **ClickHouse native / HTTP** | `localhost:9000` / `localhost:8123` |

> PostgreSQL is **not** published to the host — connect through the PgBouncer ports above.

---

## Justfile Reference

| Recipe | Scope |
| :--- | :--- |
| `just build-up` | Rebuild service images and launch Compose |
| `just up` | Launch Compose without rebuilding |
| `just down` | Stop Compose stacks and release networks |
| `just ps` | Health, uptime, and mapped ports |
| `just migrate` | Run migrations across all five bounded contexts |
| `just migrate-down` | Roll back the latest migration |
| `just seeder` | Populate mock users, roles, merchants, products, cashiers |
| `just generate-proto` | Compile `.proto` into Go models |
| `just generate-swagger` | Regenerate OpenAPI/Swagger specs |
| `just build` | Compile all services into `bin/` |
| `just build-image` | Build Docker images for all Compose services |
| `just image-load` | Load built images into a local minikube cluster |
| `just infra-up` / `infra-down` / `infra-ps` | Boot/stop the infra-only stack (5 PostgreSQL/PgBouncer pairs, Redis, Kafka, ClickHouse, observability) |
| `just services-up` / `services-down` | Start/stop local service binaries against the infra stack |
| `just e2e-hurl` | Full E2E suite: infra up → migrate → seed → services → hurl checks |
| `just smoke-trace` | Verify trace_id continuity across HTTP → gRPC → Kafka → email via Jaeger |
| `just smoke-stack-lifecycle yes` | Infra lifecycle test: clean volume start, restart/stop/start, no data loss |
| `just tidy-all` | `go mod tidy` across all modules |
| `just test-unit` | Unit tests under `pkg/` |
| `just test-integration` | Integration tests under `tests/` |
| `just test-all` | Unit + integration |

---

## Workspace Directory Tree

```
monolith-pointofsale-grpc/
├── proto/                          # Protobuf contracts
│   ├── role.proto                  #   Role query specifications
│   ├── user_role.proto             #   UserRole assignment specifications
│   └── ...                         #   auth, user, merchant, category, product,
│                                   #   cashier, order, order_item, transaction, stats
├── pb/                             # Compiled protobuf outputs
├── shared/                         # Consolidated workspace module
│   ├── domain/                     #   Domain models & requests
│   ├── mapper/                     #   Proto ↔ Go converters
│   ├── cache/                      #   Redis caching wrappers
│   ├── observability/              #   Cache/tracing interceptors
│   ├── errors/                     #   Error templates
│   └── errorhandler/               #   Error handling utilities
├── pkg/                            # Platform libraries module
│   ├── adapter/                    #   Guarded gRPC adapters (role, user_role)
│   ├── auth/                       #   JWT utilities
│   ├── database/                   #   Per-context GORM + names.go (cluster constants)
│   ├── clickhouse/                 #   OLAP connection + schema bootstrap
│   ├── kafka/                      #   Kafka producer/consumer
│   ├── outbox/                     #   Transactional outbox helpers
│   ├── redis/                      #   Redis connectors
│   ├── resilience/                 #   Circuit breaker, rate limiter, load monitor
│   ├── server/                     #   gRPC bootstrap
│   ├── otel/                       #   Telemetry hooks
│   └── logger/                     #   Zap structured logging
├── service/                        # Business domains
│   ├── apigateway/                 #   REST API gateway (Echo)
│   ├── auth/ user/ role/           #   Identity context
│   ├── merchant/ cashier/          #   Merchant context
│   ├── category/ product/          #   Catalog context
│   ├── order/ order_item/          #   Sales context
│   │   transaction/
│   ├── stats-writer/               #   Kafka → ClickHouse (OLAP write)
│   ├── stats-reader/               #   ClickHouse → gRPC (OLAP read)
│   ├── email/                      #   Kafka notification worker
│   ├── migrate/                    #   Per-context migration runner
│   └── seeder/                     #   Development seeder
├── deployments/
│   ├── local/                      #   docker-compose.yml + docker-compose.infra.yml
│   ├── kubernetes/                 #   Base + overlays, database, cache, messaging,
│   │                               #   networking, observability, security, HPAs
│   └── gitops/argocd/              #   ArgoCD project, root app, production app
├── observability/                  #   Prometheus, Loki, OTel, Promtail, alert rules
├── nginx/                          #   Reverse-proxy edge rules
├── redis/                          #   Redis configuration
├── tests/                          #   Integration tests, hurl E2E, smoke scripts
├── scripts/                        #   Helper shell scripts
├── uploads/                        #   Uploaded product images
└── images/                         #   Architecture diagrams
```

---

## License

This project is open-sourced under the MIT License for educational and development purposes.

---

<p align="center">
  Built with Go, gRPC, Apache Kafka, ClickHouse OLAP, PostgreSQL clusters behind PgBouncer, and a passion for high-performance point of sale microservices.
</p>
