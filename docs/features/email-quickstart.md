# Email Integration Quick Start

Get started with email integration in 5 minutes!

## Prerequisites

- PostgreSQL database running
- Outlook/Office365 account (or any IMAP-compatible email account)
- Application installed and configured

## Quick Setup

### Step 1: Apply Database Migration

```powershell
.\scripts\apply-email-migration.ps1
```

This creates the `emails` table for storing email data.

### Step 2: Get Outlook App Password

1. Go to https://account.microsoft.com/security
2. Click "Advanced security options"
3. Under "App passwords", create a new one
4. Copy the generated password

### Step 3: Configure Email Settings

Add to your `.env` file:

```env
# Enable email integration
EMAIL_ENABLED=true

# Outlook/Office365 settings
EMAIL_HOST=outlook.office365.com
EMAIL_PORT=993
EMAIL_USERNAME=your-email@outlook.com
EMAIL_PASSWORD=your-app-password-here
EMAIL_USE_TLS=true

# Only process emails from this sender
EMAIL_ALLOWED_SENDERS=noreply@x.ai

# Check for new emails every 5 minutes
EMAIL_POLL_INTERVAL_MINUTES=5
```

### Step 4: Restart Application

```powershell
.\scripts\restart-with-fmp.ps1
```

### Step 5: Verify It's Working

Check the logs for:

```
INFO  [email] Email connection test successful
INFO  [email-processor] Email processor started with interval: 5m0s
```

## Testing

### Send a Test Email

1. Send an email from `noreply@x.ai` (or configured sender) to your configured email address
2. Wait for next polling cycle (max 5 minutes)
3. Check logs for processing confirmation

### Verify in Database

```sql
-- Check for emails
SELECT * FROM emails ORDER BY received_date DESC LIMIT 5;

-- Check for created articles
SELECT 
    e.subject as email_subject,
    a.title as article_title,
    e.sender,
    e.received_date,
    e.processed
FROM emails e
LEFT JOIN articles a ON e.article_id = a.id
ORDER BY e.received_date DESC
LIMIT 5;
```

## Common Issues

### "Connection failed"
- Verify credentials are correct
- Check if using app-specific password (required for Outlook with 2FA)
- Ensure firewall allows port 993

### "No emails being processed"
- Check sender is in `EMAIL_ALLOWED_SENDERS`
- Verify emails are unread
- Check polling interval hasn't been set too high

### "Authentication failed"
- Use app-specific password, not regular password
- Verify username is full email address
- Check if IMAP is enabled on the account

## Configuration Options

### For Gmail

```env
EMAIL_HOST=imap.gmail.com
EMAIL_PORT=993
EMAIL_USERNAME=your-email@gmail.com
EMAIL_PASSWORD=your-app-password
EMAIL_USE_TLS=true
```

Note: Gmail requires app-specific passwords when 2FA is enabled.

### For Custom IMAP Server

```env
EMAIL_HOST=your-imap-server.com
EMAIL_PORT=993
EMAIL_USERNAME=username
EMAIL_PASSWORD=password
EMAIL_USE_TLS=true
```

### Multiple Allowed Senders

```env
EMAIL_ALLOWED_SENDERS=noreply@x.ai,alerts@example.com,news@source.com
```

## What Happens Next?

Once configured:

1. **Every 5 minutes** (or your configured interval):
   - System checks for new emails
   - Filters by allowed senders
   - Stores emails in database

2. **For each valid email**:
   - Creates article from email content
   - Links email to article
   - (Optional) Processes with AI for metadata

3. **Failed emails**:
   - Automatically retried up to 3 times
   - Error tracked in database
   - Logged for debugging

## Next Steps

- [Full Email Integration Guide](./email-integration.md) - Detailed documentation
- [AI Processing](./ai-processing.md) - Enable AI enrichment for emails
- [Configuration Reference](../getting-started/quick-start.md) - All config options

## Need Help?

- Check logs: `.\scripts\restart-with-fmp.ps1` shows startup logs
- Query database: Use SQL queries above to check status
- Review errors: `SELECT * FROM emails WHERE error IS NOT NULL`
- Read full docs: [Email Integration Documentation](./email-integration.md)