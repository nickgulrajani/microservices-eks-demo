/**
 * API Gateway Microservice
 * Handles routing, authentication, rate limiting, and load balancing
 */

const express = require('express');
const cors = require('cors');
const helmet = require('helmet');
const morgan = require('morgan');
const rateLimit = require('express-rate-limit');
const { createProxyMiddleware } = require('http-proxy-middleware');
require('dotenv').config();

const app = express();
const PORT = process.env.PORT || 3000;
const NODE_ENV = process.env.NODE_ENV || 'development';

// Security middleware
app.use(helmet());
app.use(cors({
  origin: process.env.ALLOWED_ORIGINS?.split(',') || ['http://localhost:3000'],
  credentials: true
}));

// Logging
app.use(morgan('combined'));

// Rate limiting
const limiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 100, // limit each IP to 100 requests per windowMs
  message: 'Too many requests from this IP'
});
app.use(limiter);

// Body parsing
app.use(express.json({ limit: '10mb' }));
app.use(express.urlencoded({ extended: true }));

// Health check endpoint
app.get('/health', (req, res) => {
  res.status(200).json({
    status: 'healthy',
    timestamp: new Date().toISOString(),
    environment: NODE_ENV,
    version: process.env.npm_package_version || '1.0.0',
    uptime: process.uptime()
  });
});

// Service discovery configuration
const services = {
  'user-service': process.env.USER_SERVICE_URL || 'http://user-service:3001',
  'order-service': process.env.ORDER_SERVICE_URL || 'http://order-service:3002',
  'notification-service': process.env.NOTIFICATION_SERVICE_URL || 'http://notification-service:3003'
};

// Create proxy middlewares for each service
Object.entries(services).forEach(([serviceName, serviceUrl]) => {
  app.use(`/api/${serviceName}`, createProxyMiddleware({
    target: serviceUrl,
    changeOrigin: true,
    pathRewrite: {
      [`^/api/${serviceName}`]: ''
    },
    onError: (err, req, res) => {
      console.error(`Error proxying to ${serviceName}:`, err.message);
      res.status(503).json({
        error: 'Service Unavailable',
        service: serviceName,
        message: 'The requested service is currently unavailable'
      });
    },
    onProxyReq: (proxyReq, req, res) => {
      // Add correlation ID for distributed tracing
      proxyReq.setHeader('X-Correlation-ID', req.headers['x-correlation-id'] || generateCorrelationId());
    }
  }));
});

// API documentation endpoint
app.get('/api/docs', (req, res) => {
  res.json({
    title: 'Microservices API Gateway',
    version: '1.0.0',
    services: Object.keys(services),
    endpoints: {
      '/health': 'Health check',
      '/api/user-service/*': 'User management operations',
      '/api/order-service/*': 'Order processing operations',
      '/api/notification-service/*': 'Notification operations'
    }
  });
});

// Global error handler
app.use((err, req, res, next) => {
  console.error('Global error handler:', err);
  res.status(500).json({
    error: 'Internal Server Error',
    message: NODE_ENV === 'development' ? err.message : 'Something went wrong'
  });
});

// 404 handler
app.use('*', (req, res) => {
  res.status(404).json({
    error: 'Not Found',
    message: 'The requested endpoint does not exist'
  });
});

// Graceful shutdown
process.on('SIGTERM', () => {
  console.log('SIGTERM received, shutting down gracefully');
  server.close(() => {
    console.log('Process terminated');
  });
});

function generateCorrelationId() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0;
    const v = c == 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

const server = app.listen(PORT, '0.0.0.0', () => {
  console.log(`🚀 API Gateway running on port ${PORT} in ${NODE_ENV} mode`);
  console.log(`📋 Available services: ${Object.keys(services).join(', ')}`);
});

module.exports = app;
