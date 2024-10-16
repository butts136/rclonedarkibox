# Étape 1 : Utiliser une image de base avec Go pour le builder
FROM golang:1.21-alpine as builder

# Installer les dépendances nécessaires
RUN apk --no-cache add ca-certificates fuse3 tzdata make bash gawk git

# Définir le répertoire de travail
WORKDIR /go/src/github.com/rclone/rclone

# Copier les fichiers du projet
COPY . .

# Nettoyer et installer les dépendances, puis construire le projet
RUN go mod tidy && go mod vendor && CGO_ENABLED=0 make

# Étape 2 : Créer une image légère pour exécuter l'application
FROM alpine:3.18

# Installer les certificats et fuse
RUN apk --no-cache add ca-certificates fuse3 tzdata && \
    echo "user_allow_other" >> /etc/fuse.conf

# Copier le binaire rclone de l'étape précédente
COPY --from=builder /go/src/github.com/rclone/rclone/rclone /usr/local/bin/

# Exposer les ports nécessaires (si requis)
EXPOSE 8080

# Point d'entrée pour l'image
ENTRYPOINT ["rclone"]
