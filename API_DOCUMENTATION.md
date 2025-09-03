# Age Analysis Messaging Application (AAMA) - API Documentation

This document provides comprehensive API documentation for the Age Analysis Messaging Application's file upload, Excel parsing, and messaging functionality.

## Overview

The AAMA system provides complete functionality for:
- File upload with validation and progress tracking
- Excel parsing with flexible column mapping
- Account data processing and age analysis
- Bulk messaging with template support
- Real-time progress monitoring

## Authentication

All API endpoints require authentication via JWT token in the `Authorization` header:
```
Authorization: Bearer <jwt_token>
```

## File Upload & Processing

### POST /api/uploads
Upload and process an Excel file containing account age analysis data.

**Request:**
- Content-Type: `multipart/form-data`
- Body: Form data with `file` field containing Excel file

**File Requirements:**
- Format: `.xlsx` or `.xls`
- Max size: 50MB
- Required columns: Account, Name, Telephone, Total Balance
- Optional columns: Contact, Current, 30 Days, 60 Days, 90 Days, 120 Days

**Excel Column Mapping:**
The system automatically maps various column header variations:
- **Account**: "Account", "Account Code", "Account Number", "Acc Code"
- **Name**: "Name", "Customer Name", "Customer", "Client Name", "Company"
- **Contact**: "Contact", "Contact Person", "Rep", "Account Manager"
- **Telephone**: "Telephone", "Phone", "Mobile", "Cell", "Contact Number"
- **Current**: "Current", "Current Balance", "0-30", "0 Days"
- **30 Days**: "30 Days", "30-60", "31-60 Days", "30-59 Days"
- **60 Days**: "60 Days", "60-90", "61-90 Days", "60-89 Days"
- **90 Days**: "90 Days", "90-120", "91-120 Days", "90-119 Days"
- **120+ Days**: "120 Days", "120+", "Over 120", "120 Plus"
- **Total Balance**: "Total Balance", "Total", "Balance", "Amount", "Outstanding"

**Response:**
```json
{
  "message": "File uploaded successfully. Processing started.",
  "upload": {
    "id": "uuid",
    "filename": "generated_filename.xlsx",
    "original_filename": "uploaded_file.xlsx",
    "status": "pending",
    "created_at": "2025-01-01T10:00:00Z"
  }
}
```

### GET /api/uploads/{id}/progress
Get real-time processing progress for an upload.

**Response:**
```json
{
  "upload_id": "uuid",
  "stage": "parsing",
  "total_rows": 1000,
  "processed_rows": 250,
  "success_rows": 240,
  "failed_rows": 10,
  "progress": 0.25,
  "message": "Parsing Excel file... (250/1000 rows, 25%)",
  "start_time": "2025-01-01T10:00:00Z",
  "update_time": "2025-01-01T10:01:30Z",
  "is_complete": false,
  "has_error": false
}
```

**Processing Stages:**
- `starting`: Initial setup
- `parsing`: Reading and parsing Excel file
- `saving`: Saving accounts to database
- `completed`: Processing finished successfully
- `failed`: Processing failed with error

### GET /api/uploads
List all uploads for the current user with pagination.

**Query Parameters:**
- `page`: Page number (default: 1)
- `per_page`: Results per page (default: 20, max: 100)

**Response:**
```json
{
  "uploads": [
    {
      "id": "uuid",
      "filename": "report_2025.xlsx",
      "original_filename": "Age Analysis Report.xlsx",
      "status": "completed",
      "records_processed": 850,
      "records_failed": 15,
      "created_at": "2025-01-01T10:00:00Z",
      "processing_completed_at": "2025-01-01T10:03:45Z"
    }
  ],
  "total": 10,
  "page": 1,
  "per_page": 20,
  "total_pages": 1,
  "has_more": false
}
```

### GET /api/uploads/{id}
Get details for a specific upload.

**Response:**
```json
{
  "id": "uuid",
  "user_id": "uuid",
  "filename": "report_2025.xlsx",
  "original_filename": "Age Analysis Report.xlsx",
  "file_size": 2048576,
  "mime_type": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
  "status": "completed",
  "processing_started_at": "2025-01-01T10:00:30Z",
  "processing_completed_at": "2025-01-01T10:03:45Z",
  "records_processed": 850,
  "records_failed": 15,
  "created_at": "2025-01-01T10:00:00Z",
  "updated_at": "2025-01-01T10:03:45Z"
}
```

## Account Management

### GET /api/uploads/{id}/accounts
Get paginated list of accounts from an upload.

**Query Parameters:**
- `page`: Page number (default: 1)
- `per_page`: Results per page (default: 20, max: 100)

**Response:**
```json
{
  "accounts": [
    {
      "id": "uuid",
      "upload_id": "uuid",
      "account_code": "ACC001",
      "customer_name": "ABC Company Ltd",
      "contact_person": "John Smith",
      "telephone": "27821234567",
      "amount_current": 5000.00,
      "amount_30d": 2500.00,
      "amount_60d": 1000.00,
      "amount_90d": 500.00,
      "amount_120d": 0.00,
      "total_balance": 9000.00,
      "is_selected": false,
      "created_at": "2025-01-01T10:02:15Z",
      "updated_at": "2025-01-01T10:02:15Z"
    }
  ],
  "total": 850,
  "page": 1,
  "per_page": 20,
  "total_pages": 43,
  "has_more": true
}
```

### GET /api/uploads/{id}/summary
Get comprehensive analysis summary for an upload.

**Response:**
```json
{
  "upload_id": "uuid",
  "upload": {
    "id": "uuid",
    "status": "completed",
    "records_processed": 850
  },
  "summary": {
    "total_accounts": 850,
    "selected_accounts": 125,
    "total_balance": 2500000.00,
    "selected_balance": 450000.00,
    "overdue_accounts": 320,
    "selected_overdue": 95,
    "average_balance": 2941.18,
    "selected_avg_balance": 3600.00
  },
  "age_analysis": {
    "total_balance": 2500000.00,
    "selected_balance": 450000.00,
    "buckets": [
      {
        "name": "Current",
        "days_range": "0 days",
        "total_amount": 1200000.00,
        "selected_amount": 180000.00,
        "account_count": 450,
        "selected_count": 65,
        "percentage": 48.0,
        "selected_percentage": 40.0
      },
      {
        "name": "30 Days",
        "days_range": "30-59 days",
        "total_amount": 650000.00,
        "selected_amount": 135000.00,
        "account_count": 200,
        "selected_count": 35,
        "percentage": 26.0,
        "selected_percentage": 30.0
      }
    ],
    "top_accounts": [
      {
        "id": "uuid",
        "account_code": "ACC001",
        "customer_name": "Big Debtor Inc",
        "total_balance": 75000.00,
        "overdue_amount": 45000.00,
        "oldest_bucket": "90-119 days",
        "is_selected": true
      }
    ],
    "statistics": {
      "total_accounts": 850,
      "selected_accounts": 125,
      "overdue_accounts": 320,
      "selected_overdue": 95,
      "average_balance": 2941.18,
      "selected_average_balance": 3600.00,
      "median_balance": 2150.00,
      "largest_balance": 75000.00,
      "smallest_balance": 250.00,
      "overdue_percentage": 37.6,
      "selection_rate": 14.7
    }
  }
}
```

### PUT /api/uploads/{id}/selection
Update selection status for multiple accounts.

**Request Body:**
```json
{
  "account_ids": ["uuid1", "uuid2", "uuid3"],
  "is_selected": true
}
```

**Response:**
```json
{
  "message": "Updated selection for 3 accounts",
  "updated_count": 3
}
```

## Messaging

### GET /api/messaging/templates
Get available message templates.

**Response:**
```json
{
  "templates": [
    {
      "name": "Default Payment Reminder",
      "template": "Hi [CustomerName], this is a reminder regarding account [AccountCode] for an outstanding balance of R[TotalBalance]. Please contact us to arrange payment."
    },
    {
      "name": "Friendly Reminder",
      "template": "Hi [CustomerName], hope you're well! This is a friendly reminder about your account [AccountCode] with an outstanding balance of R[TotalBalance]. Please get in touch when convenient."
    }
  ]
}
```

**Available Placeholders:**
- `[CustomerName]`: Customer name
- `[AccountCode]`: Account code/number
- `[TotalBalance]`: Total outstanding balance
- `[Current]`: Current amount (0-30 days)
- `[30Days]`: 30-59 days amount
- `[60Days]`: 60-89 days amount
- `[90Days]`: 90-119 days amount
- `[120Days]`: 120+ days amount
- `[ContactPerson]`: Contact person name
- `[Telephone]`: Phone number
- `[OverdueAmount]`: Total overdue amount
- `[OldestAgeBracket]`: Oldest overdue age bracket

### POST /api/messaging/send
Send messages to selected accounts.

**Request Body:**
```json
{
  "account_ids": ["uuid1", "uuid2", "uuid3"],
  "message_template": "Hi [CustomerName], this is a reminder regarding account [AccountCode] for an outstanding balance of R[TotalBalance]. Please contact us.",
  "max_retries": 3
}
```

**Response:**
```json
{
  "message": "Messaging job started successfully",
  "account_count": 3,
  "upload_id": "uuid",
  "template_used": "Hi [CustomerName]...",
  "estimated_time": "6 seconds"
}
```

### GET /api/messaging/logs/{upload_id}
Get message logs for an upload with pagination.

**Query Parameters:**
- `page`: Page number (default: 1)
- `per_page`: Results per page (default: 20, max: 100)

**Response:**
```json
{
  "message_logs": [
    {
      "id": "uuid",
      "account_id": "uuid",
      "upload_id": "uuid",
      "user_id": "uuid",
      "message_template": "Hi [CustomerName]...",
      "message_content": "Hi John Smith, this is a reminder...",
      "recipient_telephone": "27821234567",
      "status": "sent",
      "external_message_id": "msg_12345",
      "sent_at": "2025-01-01T15:30:00Z",
      "retry_count": 0,
      "max_retries": 3,
      "response_from_service": "Message sent successfully",
      "created_at": "2025-01-01T15:29:45Z",
      "updated_at": "2025-01-01T15:30:00Z"
    }
  ],
  "total": 125,
  "page": 1,
  "per_page": 20,
  "total_pages": 7,
  "has_more": true
}
```

**Message Status Values:**
- `pending`: Message queued for sending
- `sent`: Message sent successfully
- `failed`: Message failed to send
- `delivered`: Message delivered (if supported by service)
- `read`: Message read by recipient (if supported)

### GET /api/messaging/logs/{upload_id}/summary
Get messaging summary for an upload.

**Response:**
```json
{
  "upload_id": "uuid",
  "total_messages": 125,
  "sent_messages": 118,
  "failed_messages": 7,
  "delivered_messages": 95,
  "first_sent_at": "2025-01-01T15:30:00Z",
  "last_sent_at": "2025-01-01T15:34:30Z"
}
```

## Error Handling

All endpoints return appropriate HTTP status codes:

- `200 OK`: Successful request
- `201 Created`: Resource created successfully
- `202 Accepted`: Request accepted for processing
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Access denied
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

**Error Response Format:**
```json
{
  "message": "Error description",
  "code": "ERROR_CODE",
  "details": "Additional error details if available"
}
```

## Rate Limiting

The messaging service implements rate limiting:
- Default: 5 messages per second
- Configurable via environment variables
- Automatic retry with exponential backoff for rate limit errors

## Configuration

### Environment Variables

**File Upload:**
- `UPLOAD_MAX_SIZE`: Maximum file size in bytes (default: 52428800 = 50MB)
- `UPLOAD_DIR`: Directory for storing uploaded files (default: "uploads")

**Messaging Service:**
- `MESSAGING_API_URL`: External messaging service URL
- `MESSAGING_API_KEY`: API key for messaging service
- `MESSAGING_TIMEOUT`: Request timeout (default: "30s")
- `MESSAGING_RATE_LIMIT`: Messages per second (default: 5)
- `MESSAGING_MAX_RETRIES`: Maximum retry attempts (default: 3)
- `MESSAGING_RETRY_DELAY`: Delay between retries (default: "5s")
- `MESSAGING_FROM`: Default sender identifier (default: "DebtCollection")
- `MESSAGING_SIMULATION`: Enable simulation mode for testing (default: true)

## Data Validation

### Excel File Validation:
- File format must be .xlsx or .xls
- Must contain header row and at least one data row
- Required columns: Account, Name, Telephone, Total Balance
- Amount fields parsed with currency symbol handling
- Phone numbers cleaned and formatted for South African numbers
- Aging bucket amounts validated against total balance (1% variance allowed)

### Phone Number Processing:
- Removes formatting characters: (, ), -, space, .
- Adds South African country code (27) if missing
- Validates format and length
- Rejects invalid numbers during messaging

### Message Template Validation:
- Maximum 1600 characters (SMS limit)
- Required placeholders: [CustomerName], [AccountCode], [TotalBalance]
- Validates against known placeholder list
- Prevents template injection attacks

## Performance Considerations

- **Async Processing**: File uploads are processed asynchronously to avoid timeouts
- **Batch Operations**: Database operations use batch inserts for efficiency
- **Progress Tracking**: Real-time progress updates for large file processing
- **Rate Limiting**: Prevents overwhelming external messaging services
- **Pagination**: All list endpoints support pagination to handle large datasets
- **Connection Pooling**: Database connections are pooled for optimal performance

## Security Features

- **JWT Authentication**: All endpoints require valid JWT tokens
- **User Isolation**: Users can only access their own uploads and data
- **File Validation**: Strict file type and size validation
- **Input Sanitization**: All user inputs are validated and sanitized
- **SQL Injection Prevention**: Parameterized queries throughout
- **CORS Configuration**: Proper CORS headers for frontend integration