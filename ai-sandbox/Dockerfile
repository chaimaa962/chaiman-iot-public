FROM python:3.9-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    g++ \
    && rm -rf /var/lib/apt/lists/*

RUN pip install --no-cache-dir \
    tensorflow \
    onnx \
    onnxruntime \
    tf2onnx \
    numpy

WORKDIR /app

COPY convert.py /app/convert.py

RUN chmod +x /app/convert.py

ENTRYPOINT ["python", "/app/convert.py"]
