FROM python:3.10-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY ml-monitor/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY ml-monitor/monitor.py .

CMD ["python", "monitor.py"]
