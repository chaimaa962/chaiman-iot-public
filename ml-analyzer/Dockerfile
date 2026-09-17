FROM python:3.10-slim

WORKDIR /app

# Dépendances système pour TensorFlow Lite
RUN apt-get update && apt-get install -y --no-install-recommends \
    libgomp1 \
    && rm -rf /var/lib/apt/lists/*

# Dépendances Python
COPY ml-analyzer/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Code
COPY ml-analyzer/app /app
COPY ml-analyzer/models /app/models

EXPOSE 5000
CMD ["python", "api.py"]
