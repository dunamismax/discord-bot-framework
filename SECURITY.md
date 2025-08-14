# Security Policy

## Supported Versions

We release patches for security vulnerabilities in the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 2.0.x   | Supported |
| 1.x.x   | Not Supported |

## Reporting a Vulnerability

The Discord Bot Framework team takes security bugs seriously. We appreciate your efforts to responsibly disclose your findings, and will make every effort to acknowledge your contributions.

### How to Report a Security Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please send an email to [security@discord-bot-framework.com](mailto:security@discord-bot-framework.com) with:

1. **Description**: A clear description of the vulnerability
2. **Impact**: The potential impact of the vulnerability
3. **Reproduction**: Steps to reproduce the vulnerability
4. **Environment**: The environment where you discovered the vulnerability
5. **Suggested Fix**: If you have suggestions for how to fix the vulnerability

### What to Expect

- **Acknowledgment**: We will acknowledge receipt of your vulnerability report within 2 business days
- **Investigation**: We will investigate and validate the vulnerability within 5 business days
- **Communication**: We will keep you informed of our progress throughout the process
- **Resolution**: We aim to resolve critical vulnerabilities within 30 days
- **Credit**: We will credit you in our security advisories (unless you prefer to remain anonymous)

### Security Response Process

1. **Report received** - We acknowledge your report within 2 business days
2. **Initial triage** - We assess the severity and impact within 5 business days
3. **Investigation** - We investigate the issue and develop a fix
4. **Testing** - We test the fix in our development environment
5. **Release** - We release a security patch and publish a security advisory
6. **Disclosure** - We coordinate disclosure with you

## Security Best Practices for Users

### Bot Token Security

- **Never commit Discord bot tokens to version control**
- **Use environment variables or secure secret management**
- **Rotate tokens regularly**
- **Use separate tokens for different environments**
- **Limit bot permissions to the minimum required**

### Deployment Security

- **Run applications as non-root users**
- **Use secure systemd service configurations**
- **Restrict network access where appropriate**
- **Regularly update dependencies with `mage vulnCheck`**

### Monitoring and Alerting

- **Enable security event logging**
- **Monitor for suspicious activity**
- **Set up alerts for security incidents**
- **Review logs regularly**

## Security Features

### Input Validation

- All user inputs are validated and sanitized
- Protection against XSS, SQL injection, and command injection
- Rate limiting on user commands
- Token pattern detection and blocking

### Authentication and Authorization

- Secure bot authentication with Discord
- Role-based command access control
- Admin-only sensitive operations
- Audit logging for privileged actions

### Data Protection

- Minimal data collection and storage
- Secure handling of user data
- Data encryption in transit
- Secure database configurations

### Infrastructure Security

- Binary security scanning
- Dependency vulnerability scanning with `mage vulnCheck`
- Manual security updates via `mage setup`
- Secure deployment practices with mage commands

## Vulnerability Disclosure Timeline

We are committed to working with security researchers to verify, reproduce, and respond to legitimate reported vulnerabilities. We will make a good faith effort to:

- Respond to your report within 2 business days
- Work with you to understand the scope and severity of the issue
- Keep you informed of our progress toward a fix
- Credit your responsible disclosure when we publish a security advisory

## Security Updates

Security updates will be released as patch versions (e.g., 2.0.1, 2.0.2) and will include:

- A security advisory describing the vulnerability
- The fixed version number
- Credit to the security researcher (if desired)
- Upgrade instructions

Subscribe to security updates:
- GitHub Security Advisories: Follow this repository
- Security mailing list: [security-announce@discord-bot-framework.com](mailto:security-announce@discord-bot-framework.com)

## Security Contact Information

- **Security Email**: [security@discord-bot-framework.com](mailto:security@discord-bot-framework.com)
- **PGP Key**: Available at [https://discord-bot-framework.com/.well-known/pgp-key.txt](https://discord-bot-framework.com/.well-known/pgp-key.txt)
- **Security Page**: [https://discord-bot-framework.com/security](https://discord-bot-framework.com/security)

## Legal

This security policy is subject to our [Terms of Service](https://discord-bot-framework.com/terms) and [Privacy Policy](https://discord-bot-framework.com/privacy).

By participating in our security disclosure program, you agree to:

- Make a good faith effort to avoid privacy violations and disruption to others
- Only interact with accounts you own or have explicit permission to access
- Not access, modify, or delete data belonging to others
- Report vulnerabilities as soon as possible after discovery
- Not demand compensation for reporting vulnerabilities

We will not pursue legal action against security researchers who:

- Follow this security policy
- Report vulnerabilities in good faith
- Do not violate any applicable laws
- Do not compromise user data or service availability

---

**Last Updated**: January 15, 2025