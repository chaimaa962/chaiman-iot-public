FROM python:3.10-slim

WORKDIR /app

# Installer les dépendances système
RUN apt-get update && apt-get install -y \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Copier les requirements
COPY iot-simulator/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copier tout le code
COPY iot-simulator/ .

CMD ["python", "geth_nodes_simulator.py"]
