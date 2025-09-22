FROM python:3.11-slim

ENV PYTHONDONTWRITEBYTECODE=1 \
  PYTHONUNBUFFERED=1 \
  PIP_NO_CACHE_DIR=off \
  PIP_DISABLE_PIP_VERSION_CHECK=on

WORKDIR /app

COPY ./services/markitdown_api/requirements.txt ./requirements.txt
RUN pip install --upgrade pip && \
  pip install --no-cache-dir -r requirements.txt

COPY ./services/markitdown_api ./

EXPOSE 8000

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
