# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.3.x   | :white_check_mark: |
| < 0.3   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability in Chotko, please report it responsibly:

1. **Do not** open a public GitHub issue for security vulnerabilities
2. Email the maintainers directly or use GitHub's private vulnerability reporting feature
3. Include as much detail as possible:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

## Security Considerations

### Credentials

- API tokens and passwords are stored in `~/.config/chotko/config.yaml`
- The config file should have restricted permissions (600)
- Consider using API tokens instead of username/password when possible
- Chotko does not transmit credentials to any third parties

### Network Security

- All communication with Zabbix servers uses HTTPS by default
- Certificate validation is enabled
- No data is sent to external services (except your configured Zabbix server)

### Dependencies

- Dependencies are regularly updated via Dependabot
- Security advisories are monitored through GitHub's security features
