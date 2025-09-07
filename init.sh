#!/bin/bash
set -e

# Vérifier si docker est installé
if ! command -v docker &> /dev/null; then
    echo "❌ Docker n'est pas installé ou introuvable dans le PATH."
    exit 1
else
    echo "✅ Docker est installé."
    docker --version
fi

# Vérifier si Docker est en cours d'exécution
if ! docker info > /dev/null 2>&1; then
    echo "❌ Le démon Docker ne tourne pas. Lance Docker Desktop ou le service docker."
    exit 1
fi

# Init postgres container
echo "🚀 Initialisation du conteneur Postgres..."

# Vérifier si le conteneur existe déjà
if [ "$(docker ps -a -q -f name=chirp_pg_database)" ]; then
    echo "ℹ️  Le conteneur 'chirp_pg_database' existe déjà."
    echo "👉 Redémarrage..."
    docker start chirp_pg_database
else
    docker run -d -p 5432:5432 \
      -e POSTGRES_PASSWORD=password \
      --name chirp_pg_database \
      postgres:15
fi
