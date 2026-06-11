# O'Z REST API

API REST de gestion de commerce multi-boutiques — construite en Go.

![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=flat&logo=postgresql&logoColor=white)
![Chi](https://img.shields.io/badge/Router-Chi_v5-black?style=flat)
![JWT](https://img.shields.io/badge/Auth-JWT-000000?style=flat&logo=jsonwebtokens&logoColor=white)

---

## À propos

O'Z est une API REST pour la gestion de commerces multi-boutiques et multi-tenants.  
Chaque administrateur gère ses propres boutiques, produits, stocks et commandes de façon isolée.

---

## Fonctionnalités

- **Authentification JWT** — inscription, connexion, reset de mot de passe par email
- **Multi-boutiques** — un admin peut gérer plusieurs points de vente
- **Produits & Catégories** — CRUD complet avec soft delete et restauration
- **Gestion de stock** — stock par boutique, entrées, ajustements, historique des mouvements, seuil d'alerte
- **Alertes stock par email** — notification automatique quand le stock passe sous le seuil minimum
- **Commandes** — création avec décrémentation automatique du stock, confirmation, livraison, annulation avec restauration du stock
- **Multi-tenant** — chaque utilisateur voit uniquement ses propres données (admin/vendeur/super)
- **Dashboard** — statistiques et notifications
- **Super admin** — gestion globale de la plateforme
- **Documentation Swagger** — interface interactive sur `/swagger`

---

## Architecture

Pattern **Store → Service → Handler** :

```
HTTP Request
     ↓
[ Handler ]    → valide la requête, renvoie la réponse HTTP
     ↓
[ Service ]    → logique métier
     ↓
[ Store   ]    → requêtes SQL (pgx/v5)
     ↓
  PostgreSQL
```

```
internal/
├── app/
│   ├── app.go            # Initialisation des dépendances
│   └── routes.go         # Déclaration des routes
├── middlewares/          # Auth, rate limiter, logger
├── models/               # Structs Go
├── pkg/
│   ├── config/           # Chargement des variables d'environnement
│   ├── errs/             # Erreurs custom avec codes HTTP
│   └── mailer/           # Envoi d'emails (SMTP / Mailjet / Resend)
├── services/             # Logique métier
├── store/                # Requêtes SQL
└── transport/            # Handlers HTTP
```

---

## Stack technique

| Outil | Usage |
|-------|-------|
| Go 1.26 | Langage principal |
| Chi v5 | Router HTTP |
| pgx/v5 | Driver PostgreSQL |
| golang-jwt | Authentification JWT |
| google/uuid | Génération d'UUID |
| go-chi/cors | Gestion CORS |
| Mailjet / Resend / SMTP | Envoi d'emails |
| swaggo/swag | Documentation Swagger |

---

## Déploiement (production)

| Service | Usage |
|---------|-------|
| [Supabase](https://supabase.com) | PostgreSQL hébergé (free tier) |
| [Render](https://render.com) | API Go via Docker (free tier) |
| [Vercel](https://vercel.com) | Frontend React (free tier) |

---

## Lancer en local

### Prérequis

- Go 1.26+
- PostgreSQL 16+

### Installation

```bash
git clone https://github.com/huguescodeur/oz-rest-api-go.git
cd oz-rest-api-go
go mod download
```

### Configuration

Copier `.env.example` en `.env` et remplir les valeurs :

```bash
cp .env.example .env
```

Variables minimales pour le dev local :

```env
DB_USER=oz
DB_PASS=oz_secret
DB_ADDR=localhost:5432
DB_NAME=oz_db
JWT_SECRET=          # openssl rand -base64 32
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=ton_email@gmail.com
SMTP_PASS=           # App Password Google
MAIL_FROM=ton_email@gmail.com
FRONTEND_URL=http://localhost:5173
```

### Base de données

```bash
psql -U oz -d oz_db -f zorodb.sql
psql -U oz -d oz_db -f migrations/001_order_features_and_categories.sql
psql -U oz -d oz_db -f migrations/002_delivery_address.sql
```

### Lancer

```bash
go run main.go
```

API disponible sur `http://localhost:8080`  
Swagger sur `http://localhost:8080/swagger/index.html`

---

## Rôles utilisateurs

| Rôle | Accès |
|------|-------|
| `super` | Accès global à toute la plateforme |
| `admin` | Gère ses propres boutiques, produits, stocks, commandes |
| `vendeur` | Accès restreint aux boutiques qui lui sont assignées |

---

## Auteur

**Hugues Codeur**
