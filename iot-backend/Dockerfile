FROM python:3.9-slim
WORKDIR /app
RUN pip install flask web3 requests python-dotenv pymongo
COPY . .
EXPOSE 5000
CMD ["python", "app.py"]
