# User Service

Microservice **Go** propriétaire des **profils clients** et de leurs **adresses de
livraison**. L'identité et les mots de passe restent dans `auth-service` : ce
service ne gère que les données personnelles, indexées par l'identifiant
utilisateur émis par l'authentification.

| | |
|---|---|
| **Langage / techno** | Go 1.26, chi (routeur), pgx (PostgreSQL), zerolog, golang-migrate |
| **Base de données** | PostgreSQL (port hôte `5437`) |
| **Port HTTP** | `8087` |

---

## Architecture — Clean / Hexagonale

```
cmd/main.go                # Démarrage, migrations, injection des dépendances
internal/
├── domain/                # Profile, Address, erreurs métier typées,
│                          # ports (interfaces des repositories)
├── application/           # Cas d'usage — orchestration pure,
│                          # aucune dépendance à HTTP ni à SQL
├── adapter/
│   ├── http/              # Routeur chi, middleware JWT, DTO, mapping des erreurs
│   └── postgres/          # Repositories pgx + migrations SQL embarquées
└── config/                # Configuration typée depuis les variables d'environnement
```

**Règle** : le domaine ne connaît ni HTTP ni SQL. Les erreurs métier sont typées
et ne sont converties en codes HTTP que dans la couche adapter.

---

## Fonctionnalités

### Profil
- **Consultation du profil** (prénom, nom, téléphone)
- **Mise à jour du profil**
- **Création automatique** du profil à l'inscription — `auth-service` appelle
  `POST /internal/profiles` juste après la création du compte (idempotent)
- **Création à la volée** : un utilisateur inscrit avant l'existence de ce service
  obtient son profil au premier accès, sans erreur

### Adresses de livraison
- **Liste** des adresses enregistrées
- **Ajout** d'une adresse (libellé, rue, code postal, ville)
- **Suppression** d'une adresse
- **Une seule adresse par défaut** à la fois : en définir une nouvelle retire
  automatiquement le drapeau de l'ancienne
- Rendu d'une adresse formatée sur une ligne (utilisée au moment du paiement)

### Cloisonnement
Un utilisateur ne voit et ne modifie **que ses propres données**. Toute tentative
d'accès à l'adresse d'un autre utilisateur renvoie `403`.

---

## Endpoints

| Méthode | Route | Accès |
|---|---|---|
| GET | `/api/users/me` | authentifié |
| PUT | `/api/users/me` | authentifié |
| GET | `/api/users/me/addresses` | authentifié |
| POST | `/api/users/me/addresses` | authentifié |
| DELETE | `/api/users/me/addresses/{id}` | authentifié (la sienne) |
| POST | `/internal/profiles` | interne (service à service, sans JWT) |
| GET | `/healthz`, `/readyz` | public (sondes) |

---

## Dépendances

> **Légende** — 🔴 indispensable (le service ne démarre pas ou ne sert à rien) ·
> 🟠 nécessaire à une fonctionnalité (le reste continue de marcher) ·
> 🟡 optionnelle (dégradation silencieuse, journalisée)

| Dépendance | Type | Conséquence si absente |
|---|---|---|
| **PostgreSQL** (`user-db`) | 🔴 | Le service ne démarre pas |
| **auth-service** | 🟠 | Ce service **n'appelle jamais** `auth-service`. Mais sans lui, personne ne peut obtenir de jeton, donc plus aucun appel n'aboutit (`401`). |

**Aucun appel sortant vers un autre service.** `user-service` est autonome.

### Qui dépend de ce service

| Service | Type | Conséquence si `user-service` est arrêté |
|---|---|---|
| `auth-service` | 🟡 | L'inscription fonctionne toujours (appel non bloquant) |
| `web-app` | 🟠 | Seule la page **Profil** est cassée ; le reste de l'application fonctionne |

> ℹ️ **La validation du JWT est locale** : chaque service vérifie la signature
> avec le secret partagé, **sans appel réseau à `auth-service`**. Si
> `auth-service` tombe, les jetons déjà émis continuent donc de fonctionner —
> seules la connexion et l'inscription sont impossibles.

---

## Lancement

```bash
docker network create microservices-net   # une seule fois, partagé
cp .env.example .env                      # renseigner POSTGRES_PASSWORD et JWT_SECRET
docker compose up -d --build
```

⚠️ `JWT_SECRET` doit être **identique** à celui de `auth-service`.

### Variables d'environnement

| Variable | Requis | Description |
|---|---|---|
| `PORT` | non (8087) | Port HTTP |
| `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB` | oui | Base dédiée `user-db` |
| `DATABASE_URL` | oui | Chaîne pgx (le compose la construit pour le conteneur) |
| `JWT_SECRET` | oui | Secret HS256 partagé avec `auth-service` |
| `LOG_LEVEL` | non (info) | `debug`, `info`, `warn`, `error` |

---

## Tests

```bash
go test ./internal/... -cover     # couverture ~77 % sur la couche application
go vet ./... && gofmt -l .
```

Les tests utilisent des repositories en mémoire — aucune base n'est requise.

> ⚠️ **Aucune CI n'est configurée sur ce projet** — les tests doivent être lancés
> manuellement.
