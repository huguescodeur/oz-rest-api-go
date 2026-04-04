# 🗡️ O'Z REST API

> API REST de gestion de commerce multi-boutiques — construite en Go, inspirée d'un projet Flask, repensée from scratch.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go&logoColor=white)
![MySQL](https://img.shields.io/badge/MySQL-8.0-4479A1?style=flat&logo=mysql&logoColor=white)
![Chi](https://img.shields.io/badge/Router-Chi-v5-black?style=flat)
![Swagger](https://img.shields.io/badge/Docs-Swagger-85EA2D?style=flat&logo=swagger&logoColor=black)
![JWT](https://img.shields.io/badge/Auth-JWT-000000?style=flat&logo=jsonwebtokens&logoColor=white)

---

## 📖 À propos

O'Z est une API REST complète pour la gestion de commerces multi-boutiques.  
Conçue pour des vendeurs qui possèdent plusieurs points de vente et souhaitent centraliser la gestion de leurs produits, stocks et commandes.

Ce projet est une réécriture et amélioration d'un projet Flask existant — migré vers Go pour de meilleures performances, une architecture plus robuste, et une logique métier enrichie (gestion de stock, mouvements, multi-tenant).

---

## ✨ Fonctionnalités

- 🔐 **Authentification JWT** — inscription, connexion, token sécurisé
- 🏪 **Multi-boutiques** — un utilisateur peut gérer plusieurs shops
- 📦 **Produits & Catégories** — CRUD complet avec soft delete et restauration
- 📊 **Gestion de stock** — stock par boutique, entrées, ajustements, historique complet des mouvements
- 🛒 **Commandes** — création avec décrémentation automatique du stock via transaction SQL, annulation avec restauration
- 👥 **Multi-tenant** — chaque utilisateur voit uniquement ses propres données
- 📄 **Documentation Swagger** — interface interactive disponible sur `/swagger`

---

## 🏗️ Architecture

Pattern **Repository + Service + Handler** (architecture en couches) :

```
HTTP Request
     ↓
[ Transport / Handler ]   → valide la requête, renvoie la réponse HTTP
     ↓
[    Service          ]   → logique métier et règles de validation
     ↓
[    Store            ]   → accès base de données uniquement
     ↓
   MySQL / vous pouvez utiliser PostgreSQL ce que je recommande
```

```
internal/
├── app/
│   ├── app.go            # Initialisation de l'application
│   └── routes.go         # Déclaration des routes
├── middlewares/
│   └── auth_middleware.go
├── models/               # Structs Go (Product, Order, Stock…)
├── pkg/
│   ├── config/           # Load variables d'environnement
│   ├── errs/             # Erreurs custom et mapping HTTP
│   └── responses/        # Helpers de réponses
├── services/             # Logique métier
├── store/                # Requêtes SQL
└── transport/            # Handlers HTTP
```

---

## 🗄️ Modèle de données

![Database Schema][def]
_(Vous avez un fichier sql pour une base de données test)_

| Table             | Description                                                     |
| ----------------- | --------------------------------------------------------------- |
| `users`           | Utilisateurs avec rôles (`super`, `admin`, `vendeur`)           |
| `shops`           | Boutiques appartenant à un user                                 |
| `user_shops`      | Table de liaison user ↔ shop                                    |
| `categories`      | Catégories de produits                                          |
| `products`        | Produits globaux (non liés à une boutique)                      |
| `stocks`          | Stock d'un produit dans une boutique (`product_id` + `shop_id`) |
| `stock_movements` | Historique de tous les mouvements de stock                      |
| `orders`          | Commandes par boutique                                          |
| `order_items`     | Détail des articles d'une commande                              |

---

## 🚀 Lancer le projet

### Prérequis

- Go 1.21+
- MySQL 8.0+

### Installation

```bash
git clone https://github.com/huguescodeur/zoro_rest_api_go.git
cd zoro_rest_api_go
go mod download
```

### Configuration

Créer un fichier `.env` à la racine :

```env
DB_USER=dbuser
DB_PASS=root
DB_ADDR=db address ex: 127.0.0.1:8889
DB_NAME=dbname
JWT_SECRET=secret (peut être généré avec: openssl rand -base64 32)
```

### Lancer

```bash
go run main.go
```

L'API sera disponible sur `http://localhost:8080`  
La documentation Swagger sur `http://localhost:8080/swagger/index.html`

---

## 📡 Endpoints

### Auth

| Méthode | Route                         | Description                                     |
| ------- | ----------------------------- | ----------------------------------------------- |
| POST    | `/api/v1/auth/register`       | Inscription                                     |
| POST    | `/api/v1/auth/login`          | Connexion → retourne JWT                        |
| POST    | `/api/v1/auth/logout`         | Déconnexion _(auth required)_                   |
| PATCH   | `/api/v1/auth/reset-password` | Réinitialiser le mot de passe _(auth required)_ |

### Produits _(auth required)_

| Méthode | Route                     | Description             |
| ------- | ------------------------- | ----------------------- |
| GET     | `/api/v1/products`        | Liste tous les produits |
| POST    | `/api/v1/products`        | Créer un produit        |
| GET     | `/api/v1/products/{uuid}` | Détail d'un produit     |
| PUT     | `/api/v1/products/{uuid}` | Modifier un produit     |
| DELETE  | `/api/v1/products/{uuid}` | Soft delete             |
| PATCH   | `/api/v1/products/{uuid}` | Restaurer               |

### Stocks _(auth required)_

| Méthode | Route                                          | Description                           |
| ------- | ---------------------------------------------- | ------------------------------------- |
| GET     | `/api/v1/stocks`                               | Tous les stocks                       |
| GET     | `/api/v1/stocks/detail?product_id=X&shop_id=Y` | Stock précis                          |
| POST    | `/api/v1/stocks/init`                          | Initialiser un stock à 0              |
| POST    | `/api/v1/stocks/in`                            | Entrée de marchandise                 |
| POST    | `/api/v1/stocks/adjust`                        | Ajustement manuel (SALE, GIFT, LOSS…) |
| GET     | `/api/v1/stocks/movements`                     | Historique des mouvements             |
| GET     | `/api/v1/stocks/movements/product/{id}`        | Mouvements par produit                |
| GET     | `/api/v1/stocks/movements/shop/{id}`           | Mouvements par boutique               |

### Commandes _(auth required)_

| Méthode | Route                   | Description                  |
| ------- | ----------------------- | ---------------------------- |
| GET     | `/api/v1/orders`        | Toutes les commandes         |
| POST    | `/api/v1/orders`        | Créer une commande           |
| GET     | `/api/v1/orders/{uuid}` | Détail d'une commande        |
| DELETE  | `/api/v1/orders/{uuid}` | Annuler + restaurer le stock |

### Boutiques _(auth required)_

| Méthode | Route                          | Description                        |
| ------- | ------------------------------ | ---------------------------------- |
| GET     | `/api/v1/shops`                | Liste toutes les boutiques         |
| POST    | `/api/v1/shops`                | Créer une boutique                 |
| GET     | `/api/v1/shops/vendeur`        | Boutiques d'un vendeur             |
| POST    | `/api/v1/shops/assign-vendeur` | Assigner un vendeur à une boutique |
| GET     | `/api/v1/shops/{uuid}`         | Détail d'une boutique              |
| PUT     | `/api/v1/shops/{uuid}`         | Modifier une boutique              |
| DELETE  | `/api/v1/shops/{uuid}`         | Soft delete                        |
| PATCH   | `/api/v1/shops/{uuid}`         | Restaurer                          |

### Utilisateurs _(auth required)_

| Méthode | Route                  | Description                 |
| ------- | ---------------------- | --------------------------- |
| GET     | `/api/v1/users`        | Liste tous les utilisateurs |
| POST    | `/api/v1/users`        | Ajouter un utilisateur      |
| GET     | `/api/v1/users/{uuid}` | Détail d'un utilisateur     |
| PUT     | `/api/v1/users/{uuid}` | Modifier un utilisateur     |
| DELETE  | `/api/v1/users/{uuid}` | Soft delete                 |
| PATCH   | `/api/v1/users/{uuid}` | Restaurer                   |

---

## 🔑 Authentification

Toutes les routes protégées nécessitent un header :

```
Authorization: Bearer <token>
```

Le token est retourné à la connexion (`/auth/login`).

---

## 📦 Exemple de requête

### Créer une commande

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "shop_id": 1,
    "items": [
      { "product_id": 3, "quantity": 2 },
      { "product_id": 7, "quantity": 1 }
    ]
  }'
```

> Le stock est décrémenté automatiquement via transaction SQL. Si le stock est insuffisant pour un article, toute la commande est annulée (rollback).

---

## 🛠️ Stack technique

| Outil                                                                 | Usage                   |
| --------------------------------------------------------------------- | ----------------------- |
| [Go](https://golang.org/)                                             | Langage principal       |
| [Chi](https://github.com/go-chi/chi)                                  | Router HTTP             |
| [MySQL](https://www.mysql.com/)                                       | Base de données         |
| [JWT (golang-jwt)](https://github.com/golang-jwt/jwt)                 | Authentification        |
| [go-playground/validator](https://github.com/go-playground/validator) | Validation des requêtes |
| [google/uuid](https://github.com/google/uuid)                         | Génération d'UUID       |
| [swaggo/swag](https://github.com/swaggo/swag)                         | Génération Swagger      |
| [go-chi/cors](https://github.com/go-chi/cors)                         | Gestion CORS            |

---

## 📄 Documentation complète

La documentation interactive Swagger est générée automatiquement depuis les annotations du code.

```bash
# Regénérer la doc Swagger après modifications
swag init
```

Puis accéder à : `http://localhost:8080/swagger/index.html`

---

## 🤝 Contribuer

Les PR sont les bienvenues. Pour les changements majeurs, ouvrir une issue d'abord.

---

## 👤 Auteur

**Goli Yao Hugues / Hugues Codeur**  
Projet personnel — réécriture et amélioration d'un projet d'apprentissage Flask.

---

_O'Z — parce que même une API doit couper net et aussi open source._ 🗡️

[def]: ./docs/schema.png
