# Notification Service Microservice
FROM openjdk:17-jdk-slim

LABEL maintainer="DevOps Team"
LABEL service="notification-service"
LABEL version="1.0.0"

WORKDIR /app

# Install curl for health checks
RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

# Copy Maven wrapper and pom.xml
COPY mvnw pom.xml ./
COPY .mvn .mvn

# Download dependencies
RUN ./mvnw dependency:go-offline -B

# Copy source code
COPY src/ ./src/

# Build the application
RUN ./mvnw clean package -DskipTests

# Create non-root user
RUN useradd -m -u 1001 appuser

# Change ownership
RUN chown -R appuser:appuser /app
USER appuser

# Expose port
EXPOSE 3003

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:3003/health || exit 1

CMD ["java", "-jar", "target/notification-service-1.0.0.jar"]
