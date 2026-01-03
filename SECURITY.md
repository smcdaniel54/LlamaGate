# Security Policy

## Supported Versions

We release patches for security vulnerabilities. Which versions are eligible for receiving such patches depends on the CVSS v3.0 Rating:

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability in LlamaGate, please report it responsibly:

1. **Do NOT** open a public GitHub issue
2. Create a private security advisory at: https://github.com/llamagate/llamagate/security/advisories/new
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

We will:
- Acknowledge receipt within 48 hours
- Provide an initial assessment within 7 days
- Keep you informed of our progress
- Credit you in the security advisory (if desired)

## Security Best Practices

### For Production Deployments

1. **Use HTTPS**: Always use a reverse proxy (nginx, Caddy, etc.) with TLS/SSL in production
   ```nginx
   # Example nginx configuration
   server {
       listen 443 ssl;
       server_name your-domain.com;
       ssl_certificate /path/to/cert.pem;
       ssl_certificate_key /path/to/key.pem;
       
       location / {
           proxy_pass http://localhost:8080;
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
       }
   }
   ```

2. **Set Strong API Keys**: Use a strong, randomly generated API key
   ```bash
   # Generate a secure API key
   openssl rand -hex 32
   ```

3. **Restrict Network Access**: 
   - Use firewall rules to restrict access
   - Bind to localhost only if not needed externally
   - Use VPN or private networks for remote access

4. **Secure Log Files**: 
   - Log files may contain sensitive data
   - Ensure proper file permissions (automatically set to 0600)
   - Rotate logs regularly
   - Consider log encryption for sensitive deployments

5. **Rate Limiting**: 
   - Configure appropriate rate limits for your use case
   - Consider per-IP rate limiting for multi-tenant scenarios

6. **Environment Variables**: 
   - Never commit `.env` files to version control
   - Use secure secret management in production
   - Rotate API keys regularly

7. **Keep Dependencies Updated**: 
   ```bash
   go get -u ./...
   go mod tidy
   ```

8. **Monitor and Audit**: 
   - Review logs regularly
   - Monitor for unusual activity
   - Set up alerts for authentication failures

### For Development

1. **Use `.env.example`**: Copy `.env.example` to `.env` and customize
2. **Disable Auth for Testing**: Set `API_KEY=` (empty) for local development
3. **Use Debug Mode**: Enable `DEBUG=true` only in development

## Known Security Considerations

1. **API Key in Headers**: API keys are transmitted in HTTP headers. Always use HTTPS in production.

2. **In-Memory Cache**: Cached responses are stored in memory. Consider cache size limits for memory-constrained environments.

3. **No Built-in HTTPS**: LlamaGate does not include TLS/SSL. Use a reverse proxy for production.

4. **Rate Limiting**: Current implementation is global, not per-IP. Consider per-IP limiting for multi-tenant scenarios.

## Security Updates

Security updates will be released as patch versions (e.g., 1.0.1, 1.0.2) and will be clearly marked in the CHANGELOG.

## Acknowledgments

We appreciate responsible disclosure of security vulnerabilities. Contributors who report security issues will be credited (if desired) in security advisories.

