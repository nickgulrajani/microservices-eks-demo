# User Service Microservice
FROM python:3.11-slim

LABEL maintainer="DevOps Team"
LABEL service="user-service"
LABEL version="1.0.0"

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements
COPY requirements.txt .

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy source code
COPY src/ ./src/

# Create non-root user
RUN useradd -m -u 1001 appuser && \
    chown -R appuser:appuser /app
USER appuser

# Expose port
EXPOSE 3001

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python src/health_check.py || exit 1

CMD ["python", "src/app.py"]
