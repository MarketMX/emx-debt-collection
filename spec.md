Technical Specification: Age Analysis Messaging Application (AAMA)
Version: 1.1
Date: 2025-09-01
Author: Gemini Solutions

1. Introduction
1.1. Project Overview
The Age Analysis Messaging Application (AAMA) is a web-based tool designed to streamline the process of sending payment reminders based on an accounts age analysis report. The application will allow an authenticated user to upload a report (in Excel or PDF format), which the system will parse to display overdue accounts. The user can then select accounts and trigger a pre-defined messaging workflow to send reminders via an external messaging service.

1.2. Goals & Objectives
Efficiency: Automate the manual process of identifying and contacting overdue accounts.

Accuracy: Reduce human error by directly parsing data from source documents.

User Experience: Provide a sleek, modern, and intuitive user interface that requires minimal training.

Integration: Seamlessly connect with the existing messaging service API.

Security: Ensure that all data is handled securely and that access is restricted to authenticated users.

2. User Stories
As an Accounts Receivable Clerk, I want to:

AA-01: Log in to the application securely so that only authorized personnel can access sensitive financial data.

AA-02: Be presented with a clean dashboard where I can initiate the upload of a new age analysis report.

AA-03: Upload an age analysis report in .xlsx format.

AA-04: Have the system automatically parse the document and display the account data in a clear, readable table.

AA-05: View key information for each account, including Customer Name, Contact Details, and amounts overdue across different aging brackets.

AA-06: Select individual, multiple, or all accounts from the parsed list to include in a messaging batch.

AA-07: Trigger the sending of reminders to all selected accounts with a single click.

AA-08: Receive clear feedback in the UI on the status of the messaging job (e.g., "Sent reminders to 85 accounts successfully").

AA-09: View a history of past uploads and the messaging jobs associated with them.

3. System Architecture
3.1. High-Level Design
The system will be a client-server application. A modern Single Page Application (SPA) will run in the user's browser, communicating with a backend server via a RESTful API.

Frontend (Client): A reactive web interface built with a modern JavaScript framework. It will handle user authentication, file uploads, and data visualization.

Backend (Server): A Go application that exposes a set of REST API endpoints. It will manage user authentication, file storage, document parsing, and business logic for interacting with the external messaging service.

Database: A relational database to store user information, metadata about uploads, and logs.

External Services: The existing messaging service, which will be integrated via its API.

3.2. Technology Stack
Component

Technology

Justification

Frontend

React (with Vite) + Tailwind CSS

Fast development, massive ecosystem, component-based, excellent performance.

UI Components

Shadcn/UI

Provides elegant, accessible, and composable components for a modern look.

Backend

Go with the Echo framework

High performance, concurrency, type-safe. Echo is fast and has good middleware.

Database

PostgreSQL

Robust, reliable, and scalable for production workloads.

Authentication

JWT (JSON Web Tokens)

Stateless, secure, and industry-standard for SPA authentication.

3.3. Data Flow
Login: User enters credentials. Frontend sends them to the backend. Backend validates against the database and returns a JWT.

Upload: User selects an Excel file. Frontend sends the file to a secure backend endpoint.

Processing: Backend saves the file and starts a background process (goroutine) to parse it. It reads the specific columns and saves the data to the accounts table.

Display: Frontend requests the parsed data and displays it in a responsive table.

Trigger Messaging: User selects accounts and clicks "Send." Frontend sends account IDs to the backend.

Dispatch: Backend retrieves account details and makes an API call for each one to the messaging service, using the telephone field.

Logging: Each message attempt is logged.

4. Functional Requirements
4.1. User Authentication
Secure login page (/login).

Passwords must be securely hashed (e.g., bcrypt).

API endpoints protected by JWT.

4.2. File Upload & Parsing
Intuitive file upload component.

Initial Supported Format: .xlsx. PDF support with LLM extraction is a future enhancement.

The backend will use a Go library (e.g., excelize) to parse the Excel file. It must be configured to read the following columns:

Account

Name

Contact

Telephone

Current

30 Days

60 Days

90 Days

120 Days

Total Balance

Clear UI feedback on upload/parsing status.

4.3. Data Display & Interaction
Parsed data displayed in a clean, sortable, and filterable table.

Checkboxes for multi-selection, including "Select All".

UI summary of selected accounts.

4.4. Messaging Integration
Securely store messaging service API credentials.

The backend will send messages to the number in the Telephone column.

Message content will use a pre-defined template. Example: Hi [Name], this is a reminder regarding account [Account] for an outstanding balance of [Total Balance]. Please contact us.

5. API Specification (High-Level)
Method

Endpoint

Description

POST

/api/v1/auth/login

Authenticates a user and returns a JWT.

POST

/api/v1/uploads

Uploads a new age analysis file for processing.

GET

/api/v1/uploads/{id}

Retrieves the status and data of a specific upload.

POST

/api/v1/messaging/send

Triggers the sending of messages to selected accounts.

GET

/api/v1/logs/messaging

Retrieves a log of past messaging jobs.

6. Database Schema
users

id (PK, UUID)

email (VARCHAR, UNIQUE)

password_hash (VARCHAR)

created_at (TIMESTAMP)

uploads

id (PK, UUID)

user_id (FK to users.id)

filename (VARCHAR)

status (VARCHAR: pending, processing, completed, failed)

created_at (TIMESTAMP)

accounts

id (PK, UUID)

upload_id (FK to uploads.id)

account_code (VARCHAR) - Maps to "Account"

customer_name (VARCHAR) - Maps to "Name"

contact_person (VARCHAR, NULLABLE) - Maps to "Contact"

telephone (VARCHAR) - Maps to "Telephone"

amount_current (DECIMAL) - Maps to "Current"

amount_30d (DECIMAL) - Maps to "30 Days"

amount_60d (DECIMAL) - Maps to "60 Days"

amount_90d (DECIMAL) - Maps to "90 Days"

amount_120d (DECIMAL) - Maps to "120 Days"

total_balance (DECIMAL) - Maps to "Total Balance"

message_logs

id (PK, UUID)

account_id (FK to accounts.id)

status (VARCHAR: sent, failed)

sent_at (TIMESTAMP)

response_from_service (TEXT)

7. Future Considerations
PDF Parsing: Integrate an LLM-based service to extract data from PDF uploads.

Customizable Message Templates: Allow users to create and manage their own message templates in the UI.

User Roles & Permissions: Introduce an "Admin" role that can manage users and system settings.

Scheduled Reminders: Allow users to schedule messaging jobs to run at a future date/time.
