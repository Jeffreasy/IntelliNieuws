# Email Integration (Outlook/IMAP)

## Overview

The email integration feature allows the NieuwsScraper to receive and process emails from configured senders (such as `noreply@x.ai`) and automatically convert them into news articles. This enables seamless integration with email-based news sources.

## Features

- **IMAP Protocol Support**: Uses industry-standard IMAP protocol for email retrieval
- **Outlook/Office365 Compatible**: Pre-configured for Outlook but supports any IMAP server
- **Sender Filtering**: Only processes emails from whitelisted senders
- **Deduplication**: Prevents duplicate processing using message IDs
- **Automatic Processing**: Converts emails to articles automatically
- **AI Integration**: Optional AI processing for sentiment, entities, and categories
- **Error Handling**: Robust retry logic with configurable max retries
- **Scheduled Polling**: Configurable polling interval (default: 5 minutes)
- **Database Storage**: Stores emails and tracks processing status

## Architecture

### Components

1. **Email Service** (`internal/email/service.go`)
   - Handles IMAP connection and authentication
   - Fetches unread emails from configured account
   - Filters emails by allowed senders
   - Marks emails as read (configurable)

2. **Email Processor** (`internal/email/processor.go`)
   - Schedules periodic email fetching
   - Converts emails to articles
   - Integrates with AI service for processing
   - Handles retries for failed emails

3. **Email Repository** (`internal/repository/email_repository.go`)
   - Manages email data in PostgreSQL
   - Provides CRUD operations
   - Tracks processing status
   - Generates statistics

4. **Database Schema** (`migrations/007_create_emails_table.sql`)
   - Stores email metadata and content
   - Links processed emails to articles
   - Tracks errors and retry counts

## Setup

### 1. Database Migration

Apply the email integration migration:

```powershell
.\scripts\apply-email-migration.ps1
```

This creates the `emails` table with the following schema:
- `id`: Primary key
- `message_id`: Unique email identifier (for deduplication)
- `sender`: Email sender address
- `subject`: Email subject line
- `body_text`: Plain text email body
- `body_html`: HTML email body
- `received_date`: When email was received
- `processed`: Processing status (boolean)
- `processed_at`: When email was processed
- `article_id`: Link to created article (if any)
- `error`: Error message (if processing failed)
- `retry_count`: Number of retry attempts
- `metadata`: Additional email metadata (JSONB)

### 2. Configuration

Add the following to your `.env` file:

```env
# Email Integration Configuration
EMAIL_ENABLED=true
EMAIL_HOST=outlook.office365.com
EMAIL_PORT=993
EMAIL_USERNAME=your-email@outlook.com
EMAIL_PASSWORD=your-app-password
EMAIL_USE_TLS=true

# Email Filtering
EMAIL_ALLOWED_SENDERS=noreply@x.ai

# Processing Settings
EMAIL_POLL_INTERVAL_MINUTES=5
EMAIL_MAX_RETRIES=3
EMAIL_RETRY_DELAY_SECONDS=5
EMAIL_MARK_AS_READ=true
EMAIL_DELETE_AFTER_READ=false
```

### 3. Outlook App Password

For Outlook/Office365, you'll need to create an app-specific password:

1. Go to https://account.microsoft.com/security
2. Navigate to "Advanced security options"
3. Under "App passwords", click "Create a new app password"
4. Use this password in `EMAIL_PASSWORD` (not your regular password)

### 4. Restart Application

```powershell
.\scripts\restart-with-fmp.ps1
```

## Configuration Options

### Email Server Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `EMAIL_ENABLED` | `false` | Enable/disable email integration |
| `EMAIL_HOST` | `outlook.office365.com` | IMAP server hostname |
| `EMAIL_PORT` | `993` | IMAP server port |
| `EMAIL_USERNAME` | - | Email account username |
| `EMAIL_PASSWORD` | - | Email account password (or app password) |
| `EMAIL_USE_TLS` | `true` | Use TLS/SSL encryption |

### Email Filtering

| Variable | Default | Description |
|----------|---------|-------------|
| `EMAIL_ALLOWED_SENDERS` | `noreply@x.ai` | Comma-separated list of allowed sender emails |

### Processing Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `EMAIL_POLL_INTERVAL_MINUTES` | `5` | How often to check for new emails |
| `EMAIL_MAX_RETRIES` | `3` | Maximum retry attempts for failed processing |
| `EMAIL_RETRY_DELAY_SECONDS` | `5` | Delay between connection retry attempts |
| `EMAIL_MARK_AS_READ` | `true` | Mark processed emails as read |
| `EMAIL_DELETE_AFTER_READ` | `false` | Delete emails after processing |

## How It Works

### Email Processing Flow

1. **Polling Cycle**
   - Email processor runs every `EMAIL_POLL_INTERVAL_MINUTES`
   - Connects to IMAP server using configured credentials
   - Searches for unread emails in INBOX

2. **Email Fetching**
   - Fetches email envelope (sender, subject, date, etc.)
   - Fetches email body (both text and HTML)
   - Extracts metadata (To, CC, In-Reply-To, etc.)

3. **Sender Filtering**
   - Checks sender against `EMAIL_ALLOWED_SENDERS` whitelist
   - Skips emails from non-allowed senders
   - Logs skipped emails for monitoring

4. **Deduplication**
   - Checks if email with same `message_id` already exists
   - Skips duplicate emails
   - Prevents reprocessing of the same email

5. **Storage**
   - Stores email in `emails` table
   - Marks as unprocessed initially
   - Includes all metadata for tracking

6. **Article Creation**
   - Converts email subject to article title
   - Uses email body as article content
   - Sets source as "Email from {sender}"
   - Creates URL as `email://{message_id}`
   - Links email to created article

7. **AI Processing** (Optional)
   - If AI is enabled, processes article asynchronously
   - Extracts sentiment, entities, categories, keywords
   - Updates article with AI-generated metadata

8. **Status Updates**
   - Marks email as processed on success
   - Records article ID for reference
   - Logs errors and increments retry count on failure

9. **Retry Logic**
   - Failed emails are retried automatically
   - Maximum retry count enforced (`EMAIL_MAX_RETRIES`)
   - Exponential backoff between retries

## Usage Examples

### Check Email Statistics

Query the database to see email processing stats:

```sql
SELECT 
    COUNT(*) as total_emails,
    COUNT(*) FILTER (WHERE processed = true) as processed,
    COUNT(*) FILTER (WHERE processed = false) as pending,
    COUNT(*) FILTER (WHERE error IS NOT NULL) as failed,
    COUNT(*) FILTER (WHERE article_id IS NOT NULL) as articles_created
FROM emails;
```

### View Recent Emails

```sql
SELECT id, sender, subject, received_date, processed, article_id
FROM emails
ORDER BY received_date DESC
LIMIT 10;
```

### Find Failed Emails

```sql
SELECT id, sender, subject, error, retry_count
FROM emails
WHERE processed = false AND error IS NOT NULL
ORDER BY received_date DESC;
```

### Get Articles Created from Emails

```sql
SELECT 
    e.id as email_id,
    e.subject,
    e.sender,
    e.received_date,
    a.id as article_id,
    a.title,
    a.published
FROM emails e
JOIN articles a ON e.article_id = a.id
ORDER BY e.received_date DESC;
```

## Monitoring

### Logs

Email integration logs are tagged with `[email]` and `[email-processor]`:

```
INFO  [email-processor] Starting email processing cycle
INFO  [email] Connecting to email server
INFO  [email] Connected to INBOX, 42 messages total
INFO  [email] Found 3 unread messages
INFO  [email-processor] Fetched 3 new emails
INFO  [email-processor] Stored email: "Daily Update" from noreply@x.ai
INFO  [email-processor] Processing email 123 into article: Daily Update
INFO  [email-processor] Created article 456 from email 123
INFO  [email-processor] Email processing cycle completed: stored=3, skipped=0, duration=2.5s
```

### Error Logs

Failed operations are logged with ERROR level:

```
ERROR [email] Failed to fetch emails: connection timeout
ERROR [email-processor] Failed to process email 123 into article: invalid content
ERROR [email-processor] Retry failed for email 123
```

## Troubleshooting

### Connection Issues

**Problem**: Cannot connect to email server

**Solutions**:
1. Verify credentials in `.env` file
2. Check if using app-specific password for Outlook
3. Verify firewall allows port 993 (IMAP with TLS)
4. Check server hostname is correct
5. Try with `EMAIL_USE_TLS=false` for debugging (not recommended for production)

### Authentication Failures

**Problem**: Login failed or authentication error

**Solutions**:
1. For Outlook: Create app-specific password
2. Verify username is full email address
3. Check if 2FA is enabled (requires app password)
4. Ensure email account has IMAP enabled

### No Emails Processing

**Problem**: Emails not being processed even though they're arriving

**Solutions**:
1. Check `EMAIL_ALLOWED_SENDERS` includes the sender
2. Verify emails are not being marked as read by another client
3. Check email polling interval (`EMAIL_POLL_INTERVAL_MINUTES`)
4. Look for errors in logs
5. Verify email integration is enabled (`EMAIL_ENABLED=true`)

### Duplicate Emails

**Problem**: Same email being processed multiple times

**Solutions**:
1. This shouldn't happen due to message_id deduplication
2. Check database for unique constraint on message_id
3. Verify migration was applied correctly
4. Check logs for deduplication skips

### Processing Failures

**Problem**: Emails are received but fail to convert to articles

**Solutions**:
1. Check logs for specific error messages
2. Verify database connection is working
3. Check if article repository is functioning
4. Ensure email has content (body_text or body_html)
5. Review retry_count in database

## Security Considerations

### Credentials

- Never commit credentials to version control
- Use environment variables for sensitive data
- For Outlook, always use app-specific passwords
- Rotate passwords periodically

### Email Content

- Email bodies are stored in database
- Consider data retention policies
- Implement cleanup for old emails if needed
- Be aware of PII in email content

### Network Security

- Always use TLS (`EMAIL_USE_TLS=true`)
- Restrict email account permissions
- Monitor for suspicious email patterns
- Consider IP whitelisting on email server

## Integration with AI Processing

When AI processing is enabled (`AI_ENABLED=true`), emails converted to articles are automatically processed for:

- **Sentiment Analysis**: Positive/negative/neutral sentiment
- **Entity Extraction**: People, organizations, locations
- **Category Classification**: News categories
- **Keyword Extraction**: Important keywords and phrases

This enriches articles created from emails with valuable metadata.

## Performance Considerations

### Polling Interval

- Default: 5 minutes
- Lower interval = faster processing but more server load
- Higher interval = less load but slower processing
- Recommend: 5-15 minutes for most use cases

### Batch Size

- Currently processes all unread emails per cycle
- For high-volume scenarios, consider limiting fetch count
- Monitor database performance with large email volumes

### Resource Usage

- IMAP connections are created per cycle (not persistent)
- Each email creates one article in database
- AI processing is asynchronous (doesn't block email fetching)
- Consider scaling horizontally for high volumes

## Future Enhancements

Potential improvements for future versions:

1. **Multiple Accounts**: Support for multiple email accounts
2. **Folder Filtering**: Process emails from specific folders
3. **Content Templates**: Custom parsing rules per sender
4. **Attachments**: Extract and store email attachments
5. **Threading**: Link related emails (replies, forwards)
6. **Webhooks**: Real-time email notifications via webhooks
7. **Rules Engine**: Advanced filtering and routing rules
8. **Archive Management**: Automatic archiving of old emails

## API Endpoints (Future)

Planned API endpoints for email management:

- `GET /api/emails` - List emails with filtering
- `GET /api/emails/{id}` - Get specific email
- `GET /api/emails/stats` - Get processing statistics
- `POST /api/emails/{id}/reprocess` - Manually reprocess email
- `DELETE /api/emails/{id}` - Delete email

## Related Documentation

- [AI Processing](./ai-processing.md)
- [Database Migrations](../getting-started/installation.md#database-setup)
- [Configuration Guide](../getting-started/quick-start.md)
- [Deployment Guide](../deployment/deployment-guide.md)

## Support

For issues or questions:
1. Check logs for error messages
2. Review this documentation
3. Check database for email status
4. Open an issue on GitHub with relevant logs