FROM node:18-alpine

WORKDIR /app

# Copier les fichiers de dépendances
COPY iot-dashboard/package*.json ./
RUN npm install

# Copier tout le code
COPY iot-dashboard/ .

# Construire l'application
RUN npm run build

# Installer serve pour servir l'application
RUN npm install -g serve

EXPOSE 3000

CMD ["serve", "-s", "build", "-l", "3000"]
