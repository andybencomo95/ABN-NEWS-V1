# ABN News

**Autor:** Andy Bencomo del Río — Estudiante de Ingeniería de Software, 7mo semestre.

ABN News es un agregador de noticias que extrae contenido de fuentes RSS (BBC, TechCrunch, NYT) y las organiza por categorías. Está pensado para quienes quieren aprender a montar su propio sitio de noticias sin depender de APIs costosas ni presupuestos grandes.

---

## Tecnologías

- **Backend:** Go + Gin (API REST) + pgx (PostgreSQL)
- **Frontend:** TypeScript + Next.js 16 + Tailwind CSS
- **Base de datos:** PostgreSQL + Redis (caché)
- **Infraestructura:** Docker / Podman
- **Traducción:** MyMemory API (gratuita)
- **Imágenes:** Fuentes RSS + Wikipedia API + Wikimedia Commons

---

## ¿Cómo funciona?

### 1. El Poller (ingesta de noticias)

Un worker en Go se encarga de **descargar los feeds RSS** de cada fuente. Por cada fuente:

- Descarga el XML del feed con un **timeout de 10 segundos**.
- Extrae el título, contenido, fecha, autor e **imagen destacada** (desde `<media:thumbnail>`, `<media:content>` o buscando en Wikipedia si no hay).
- **Si no encuentra imagen**, asigna una imagen genérica de la categoría (por ejemplo, una turbina para tecnología).
- Guarda el artículo en PostgreSQL con un **hash único** (URL + fecha) para evitar duplicados.

**Resiliencia:** cada fuente tiene su propio circuit breaker. Si una fuente falla 3 veces seguidas, el sistema espera 30 segundos, luego 1 minuto, 5 minutos y hasta 30 minutos antes de volver a intentar. Una fuente caída **no afecta a las demás**.

### 2. La API REST (Gin)

La API expone estos endpoints:

| Endpoint | Qué hace |
|----------|----------|
| `GET /api/articles` | Lista artículos (pagina, filtra por categoría) |
| `GET /api/articles/:slug` | Artículo individual |
| `GET /api/categories` | Lista de categorías |
| `GET /api/sources/health` | Estado de cada fuente RSS |

Los endpoints de artículos aceptan `?lang=es` para **traducir títulos y extractos al español** usando la API gratuita de MyMemory.

Las respuestas se cachean en Redis por 30 minutos para aguantar miles de visitas sin problemas.

### 3. La Base de Datos

PostgreSQL con 4 tablas: `categories`, `sources`, `articles` y `article_tags`. Los artículos se indexan por categoría, fecha de publicación y slug.

### 4. El Frontend (Next.js)

Un sitio estático generado con **ISR (Incremental Static Regeneration)** que se actualiza cada hora. Esto significa:

- Las páginas se generan como HTML estático en el servidor.
- Cada hora, Next.js regenera las páginas en segundo plano.
- **Si el backend falla, el CDN sigue sirviendo la última versión** sin que el usuario note nada.

El sitio tiene:
- **Página principal** con la noticia destacada y las últimas 10 noticias.
- **Páginas por categoría** (Tecnología, Deportes, Economía, Cultura, Ciencia, General).
- **Página de artículo** individual con contenido completo.
- **Selector de idioma** (🇺🇸 inglés / 🇨🇺 español) que traduce toda la interfaz y los artículos.

---

## Cómo correrlo

```bash
# 1. Base de datos
docker compose up -d postgres redis

# 2. API
cd abn-api
export DATABASE_URL="postgres://abnnews:abnnews@localhost:5432/abnnews?sslmode=disable"
export REDIS_ADDR="localhost:6379"
export PORT="8080"
go run ./cmd/api/

# 3. Poller (ingesta RSS)
export DATABASE_URL="postgres://abnnews:abnnews@localhost:5432/abnnews?sslmode=disable"
go run ./cmd/poller/

# 4. Frontend
cd abn-news
NEXT_PUBLIC_API_URL=http://localhost:8080 npm run dev
```

Abrir **http://localhost:3000** 🚀

---

## Licencia

Proyecto educativo. Código abierto para que cualquiera aprenda y lo adapte.
