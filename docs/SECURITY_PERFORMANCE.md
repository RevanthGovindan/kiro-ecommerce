# Security and Performance Optimizations

This document outlines the security and performance optimizations implemented in the ecommerce website.

## Security Features

### 1. Rate Limiting

The application implements Redis-based rate limiting with different limits for different endpoints:

- **General API**: 100 requests per minute per IP
- **Authentication**: 5 attempts per minute per IP
- **Admin Operations**: 50 requests per minute per authenticated user
- **Payment Operations**: 10 requests per minute per authenticated user

Rate limiting headers are included in responses:
- `X-RateLimit-Limit`: Maximum requests allowed
- `X-RateLimit-Remaining`: Remaining requests in current window
- `X-RateLimit-Reset`: Unix timestamp when the rate limit resets

### 2. Input Validation and Sanitization

#### Input Sanitization
- HTML escaping to prevent XSS attacks
- Removal of dangerous script tags and JavaScript protocols
- Null byte and control character removal from request bodies
- Content length validation to prevent DoS attacks

#### Validation Functions
- **Email validation**: RFC-compliant email format checking
- **Password validation**: Enforces strong passwords with:
  - Minimum 8 characters
  - At least one uppercase letter
  - At least one lowercase letter
  - At least one number
  - At least one special character
- **Phone validation**: Supports international formats (10-15 digits)
- **Price validation**: Ensures valid price ranges (0-999999.99)
- **Quantity validation**: Prevents negative or excessive quantities

### 3. Security Headers

The following security headers are automatically added to all responses:

- `X-Content-Type-Options: nosniff` - Prevents MIME type sniffing
- `X-Frame-Options: DENY` - Prevents clickjacking attacks
- `X-XSS-Protection: 1; mode=block` - Enables XSS protection
- `Strict-Transport-Security` - Enforces HTTPS connections
- `Content-Security-Policy` - Restricts resource loading
- `Referrer-Policy: strict-origin-when-cross-origin` - Controls referrer information
- `Permissions-Policy` - Restricts browser features

### 4. Authentication Security

- JWT tokens with secure signing
- Refresh token rotation
- Password hashing using bcrypt
- Role-based access control (RBAC)
- Session management with Redis

## Performance Features

### 1. Redis Caching

Intelligent caching system with different TTL values:

- **Product Catalog**: 5 minutes TTL
- **Categories**: 15 minutes TTL
- **User Profiles**: 2 minutes TTL
- **Search Results**: 10 minutes TTL

#### Cache Features
- Automatic cache invalidation on data updates
- Cache hit/miss headers (`X-Cache: HIT/MISS`)
- User-specific and public cache separation
- Graceful degradation when Redis is unavailable

### 2. Image Optimization

CDN integration with automatic image optimization:

- **Responsive Images**: Multiple sizes (thumbnail, small, medium, large, xlarge)
- **Format Optimization**: WebP format with fallbacks
- **Quality Optimization**: Adaptive quality based on image size
- **Lazy Loading Support**: Optimized URLs for lazy loading

#### Image URL Parameters
- `w`: Width in pixels
- `h`: Height in pixels
- `q`: Quality (1-100)
- `f`: Format (webp, jpg, png)
- `auto`: Automatic optimization (compress,format)

### 3. Database Optimizations

- Connection pooling with GORM
- Proper indexing on frequently queried fields
- Prepared statements to prevent SQL injection
- Query optimization with selective field loading

### 4. Middleware Performance

All middleware is designed for minimal performance impact:
- Early exit conditions to avoid unnecessary processing
- Efficient regex patterns for validation
- Minimal memory allocations
- Asynchronous cache invalidation

## Configuration

### Environment Variables

```bash
# Security Configuration
MAX_REQUEST_SIZE=10485760  # Maximum request size in bytes
ENVIRONMENT=production     # Environment mode

# CDN Configuration
CDN_BASE_URL=https://your-cdn-domain.com

# Rate Limiting (Redis required)
REDIS_URL=redis://localhost:6379
```

### Production Recommendations

1. **Enable HTTPS**: Set up SSL/TLS certificates
2. **Configure CDN**: Set up a CDN for static assets and images
3. **Redis Clustering**: Use Redis cluster for high availability
4. **Database Optimization**: 
   - Enable connection pooling
   - Set up read replicas for read-heavy operations
   - Configure proper indexes
5. **Monitoring**: Set up application monitoring and alerting
6. **Backup Strategy**: Implement automated database backups

## Testing

### Security Tests

Run security tests to verify:
- Input sanitization effectiveness
- Rate limiting functionality
- Security header presence
- Validation function accuracy

```bash
go test ./internal/middleware -v -run TestSecurity
```

### Performance Benchmarks

Run performance benchmarks to measure:
- Middleware overhead
- Validation function performance
- Cache hit/miss ratios

```bash
go test ./internal/middleware -bench=. -benchmem
```

## Monitoring and Alerting

### Key Metrics to Monitor

1. **Rate Limiting**:
   - Rate limit violations per endpoint
   - Top rate-limited IPs
   - Rate limit effectiveness

2. **Cache Performance**:
   - Cache hit ratio
   - Cache invalidation frequency
   - Redis memory usage

3. **Security Events**:
   - Failed authentication attempts
   - Input validation failures
   - Suspicious request patterns

4. **Performance Metrics**:
   - Response times per endpoint
   - Database query performance
   - Memory and CPU usage

### Alerting Rules

Set up alerts for:
- High rate of authentication failures
- Unusual traffic patterns
- Cache hit ratio below threshold
- High response times
- Security header violations

## Security Best Practices

1. **Regular Updates**: Keep dependencies updated
2. **Security Audits**: Regular security assessments
3. **Access Logs**: Monitor and analyze access logs
4. **Incident Response**: Have a security incident response plan
5. **Data Encryption**: Encrypt sensitive data at rest and in transit
6. **Backup Security**: Secure and test backup procedures
7. **User Education**: Train users on security best practices

## Performance Best Practices

1. **Database Optimization**: Regular query analysis and optimization
2. **Caching Strategy**: Implement multi-level caching
3. **CDN Usage**: Leverage CDN for static content delivery
4. **Code Profiling**: Regular performance profiling
5. **Load Testing**: Regular load testing under realistic conditions
6. **Resource Monitoring**: Monitor system resources continuously