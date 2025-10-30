# Email Integration

NieuwsScraper supports automatic email processing to convert emails into news articles. This feature allows you to receive emails from specific senders and automatically process them into articles in your news database.

## Features

- **IMAP Integration**: Connect to Outlook/Office365 or any IMAP server
- **Sender Filtering**: Only process emails from allowed senders
- **Automatic Processing**: Convert emails to articles with AI enrichment
- **Existing Email Import**: Fetch historical emails on first run
- **Duplicate Detection**: Prevent processing the same email multiple times
- **Configurable Polling**: Set custom polling intervals

## Configuration

Add the following to your `.env` file:

```env
# Email Integration Configuration (Outlook/IMAP)
EMAIL_ENABLED=true
EMAIL_HOST=outlook.office365.com
EMAIL_PORT=993
EMAIL_USERNAME=your-email@outlook.com
EMAIL_PASSWORD=your-app-password
EMAIL_USE_TLS=true

# Email Filtering (comma-separated list of allowed senders)
EMAIL_ALLOWED_SENDERS=noreply@x.ai,noreply@openai.com

# Email Processing Settings
EMAIL_POLL_INTERVAL_MINUTES=5
EMAIL_MAX_RETRIES=3
EMAIL_RETRY_DELAY_SECONDS=5
EMAIL_MARK_AS_READ=true
EMAIL_DELETE_AFTER_READ=false

# Historical Email Import
EMAIL_FETCH_EXISTING=true    # Fetch existing emails from inbox on first run
EMAIL_MAX_DAYS_BACK=30       # How many days back to search for existing emails
```

## Setup Instructions

### 1. Outlook/Office365 Setup

1. **Enable IMAP**: Go to Outlook settings → Mail → Sync email
2. **App Password**: If using 2FA, create an app password instead of your regular password
3. **IMAP Settings**:
   - Server: `outlook.office365.com`
   - Port: `993`
   - SSL/TLS: `true`

### 2. Gmail Setup (Alternative)

```env
EMAIL_HOST=imap.gmail.com
EMAIL_PORT=993
EMAIL_USERNAME=your-email@gmail.com
EMAIL_PASSWORD=your-app-password
```

**Note**: Gmail requires an app password if 2FA is enabled.

## API Endpoints

### Manual Email Fetching

Fetch existing emails manually:

```bash
POST /api/v1/email/fetch-existing
Authorization: Bearer your-api-key
```

Response:
```json
{
  "message": "Existing emails fetched successfully",
  "articles_created": 15,
  "status": "completed"
}
```

### Email Statistics

Get email processing statistics:

```bash
GET /api/v1/email/stats
Authorization: Bearer your-api-key
```

Response:
```json
{
  "stats": {
    "total_emails": 150,
    "processed_emails": 145,
    "failed_emails": 2,
    "pending_emails": 3
  }
}
```

## How It Works

1. **Connection**: Connects to your IMAP server using configured credentials
2. **Filtering**: Only processes emails from allowed senders
3. **Deduplication**: Checks message ID to prevent duplicate processing
4. **Article Creation**: Converts email content to structured article format
5. **AI Processing**: Optionally enriches articles with AI analysis
6. **Storage**: Saves both email metadata and generated articles

## Email to Article Conversion

Emails are converted using this mapping:

- **Title**: Email subject line
- **Content**: Email body (text preferred, HTML fallback)
- **Author**: Email sender
- **Source**: "Email from [sender]"
- **Published Date**: Email received date
- **URL**: Generated as `email://[message-id]`

## Automatic Processing

When enabled, the system:

1. Polls your inbox at configured intervals
2. Fetches unread emails (or all emails if `EMAIL_FETCH_EXISTING=true`)
3. Filters by allowed senders
4. Converts to articles
5. Marks emails as read (optional)
6. Processes with AI (if enabled)

## Security Considerations

- Use app passwords instead of regular passwords
- Limit allowed senders to trusted sources
- Consider enabling `EMAIL_MARK_AS_READ=false` for testing
- Monitor processing logs for unauthorized access attempts

## Troubleshooting

### Connection Issues

- Verify IMAP settings are correct
- Check if app password is required
- Ensure firewall allows IMAP connections
- Test connection manually with email client

### Processing Issues

- Check allowed senders configuration
- Verify email format (text/HTML content)
- Review application logs for error details
- Test with `EMAIL_MARK_AS_READ=false` first

### Performance

- Adjust `EMAIL_POLL_INTERVAL_MINUTES` based on email volume
- Monitor database performance with high email volumes
- Consider batch processing for large inboxes

## Example Use Cases

1. **AI Research**: Receive papers and updates from xAI/OpenAI
2. **News Alerts**: Subscribe to news services that send emails
3. **Personal Research**: Forward interesting articles to yourself
4. **Team Communication**: Process team updates and announcements

## Limitations

- IMAP only (POP3 not supported)
- Email size limits apply (configurable)
- HTML parsing is basic (prefers text content)
- No attachment processing (text content only)
- Rate limited by email provider policies