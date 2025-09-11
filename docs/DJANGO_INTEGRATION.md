# Django Admin Integration for User Provisioning

This document explains how to integrate Django Admin with the debt collection application for user provisioning and management.

## Overview

The debt collection application provides REST API endpoints for user provisioning that can be called from Django Admin to:
- Create new users with multitenancy support
- Update existing users
- Deactivate users
- Bulk provision users
- List users by client

## API Endpoints

### Base URL
```
http://localhost:8080/api/provisioning
```

### Authentication
All provisioning endpoints require an API key. Set the `PROVISIONING_API_KEY` environment variable in the Go application.

**Headers Required:**
```
Authorization: Bearer <your-api-key>
# OR
Authorization: ApiKey <your-api-key>
# OR
X-API-Key: <your-api-key>
```

### 1. Provision Single User
**POST** `/api/provisioning/users`

```json
{
  "keycloak_id": "user-123-uuid",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "engagemx_client_id": "client_medical_group_1",
  "is_active": true
}
```

**Response:**
```json
{
  "id": "generated-uuid",
  "keycloak_id": "user-123-uuid",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "engagemx_client_id": "client_medical_group_1",
  "is_active": true,
  "action": "created",
  "message": "User successfully created"
}
```

### 2. Bulk Provision Users
**POST** `/api/provisioning/users/bulk`

```json
[
  {
    "keycloak_id": "user-1-uuid",
    "email": "user1@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "engagemx_client_id": "client_medical_group_1"
  },
  {
    "keycloak_id": "user-2-uuid", 
    "email": "user2@example.com",
    "first_name": "Jane",
    "last_name": "Smith",
    "engagemx_client_id": "client_medical_group_2"
  }
]
```

**Response:**
```json
{
  "success_count": 2,
  "error_count": 0,
  "users": [
    {
      "id": "uuid-1",
      "keycloak_id": "user-1-uuid",
      "email": "user1@example.com",
      "action": "created",
      "message": "User successfully created"
    },
    {
      "id": "uuid-2", 
      "keycloak_id": "user-2-uuid",
      "email": "user2@example.com",
      "action": "updated",
      "message": "User successfully updated"
    }
  ]
}
```

### 3. Get User by Keycloak ID
**GET** `/api/provisioning/users/{keycloak_id}`

**Response:**
```json
{
  "id": "generated-uuid",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "engagemx_client_id": "client_medical_group_1",
  "is_active": true,
  "created_at": "2025-09-10T10:00:00Z",
  "updated_at": "2025-09-10T10:00:00Z"
}
```

### 4. Deactivate User
**PUT** `/api/provisioning/users/{keycloak_id}/deactivate`

**Response:**
```json
{
  "message": "User successfully deactivated",
  "action": "deactivated",
  "timestamp": "2025-09-10T10:30:00Z",
  "user": {
    "id": "generated-uuid",
    "email": "user@example.com",
    "is_active": false
  }
}
```

### 5. List Users by Client
**GET** `/api/provisioning/users?client_id=client_medical_group_1`

**Response:**
```json
{
  "client_id": "client_medical_group_1",
  "count": 2,
  "users": [
    {
      "id": "uuid-1",
      "email": "user1@example.com",
      "engagemx_client_id": "client_medical_group_1",
      "is_active": true
    },
    {
      "id": "uuid-2",
      "email": "user2@example.com", 
      "engagemx_client_id": "client_medical_group_1",
      "is_active": true
    }
  ]
}
```

## Django Implementation

### 1. Django Settings
```python
# settings.py
DEBT_COLLECTION_API_BASE_URL = "http://localhost:8080/api/provisioning"
DEBT_COLLECTION_API_KEY = "your-secure-api-key-here"
```

### 2. Django Service Class
Create a service class to handle API communication:

```python
# services/debt_collection_service.py
import requests
from django.conf import settings
from typing import Dict, List, Optional
import logging

logger = logging.getLogger(__name__)

class DebtCollectionService:
    def __init__(self):
        self.base_url = settings.DEBT_COLLECTION_API_BASE_URL
        self.api_key = settings.DEBT_COLLECTION_API_KEY
        self.headers = {
            'Authorization': f'Bearer {self.api_key}',
            'Content-Type': 'application/json'
        }
    
    def provision_user(self, user_data: Dict) -> Dict:
        """Provision a single user"""
        try:
            response = requests.post(
                f"{self.base_url}/users",
                json=user_data,
                headers=self.headers,
                timeout=30
            )
            response.raise_for_status()
            return response.json()
        except requests.RequestException as e:
            logger.error(f"Failed to provision user: {e}")
            raise
    
    def bulk_provision_users(self, users_data: List[Dict]) -> Dict:
        """Provision multiple users"""
        try:
            response = requests.post(
                f"{self.base_url}/users/bulk",
                json=users_data,
                headers=self.headers,
                timeout=60
            )
            response.raise_for_status()
            return response.json()
        except requests.RequestException as e:
            logger.error(f"Failed to bulk provision users: {e}")
            raise
    
    def deactivate_user(self, keycloak_id: str) -> Dict:
        """Deactivate a user"""
        try:
            response = requests.put(
                f"{self.base_url}/users/{keycloak_id}/deactivate",
                headers=self.headers,
                timeout=30
            )
            response.raise_for_status()
            return response.json()
        except requests.RequestException as e:
            logger.error(f"Failed to deactivate user: {e}")
            raise
    
    def get_user(self, keycloak_id: str) -> Optional[Dict]:
        """Get user by Keycloak ID"""
        try:
            response = requests.get(
                f"{self.base_url}/users/{keycloak_id}",
                headers=self.headers,
                timeout=30
            )
            if response.status_code == 404:
                return None
            response.raise_for_status()
            return response.json()
        except requests.RequestException as e:
            logger.error(f"Failed to get user: {e}")
            return None
    
    def list_users_by_client(self, client_id: str) -> List[Dict]:
        """List all users for a client"""
        try:
            response = requests.get(
                f"{self.base_url}/users",
                params={'client_id': client_id},
                headers=self.headers,
                timeout=30
            )
            response.raise_for_status()
            return response.json().get('users', [])
        except requests.RequestException as e:
            logger.error(f"Failed to list users: {e}")
            return []
```

### 3. Django Model Integration
```python
# models.py
from django.db import models
from django.contrib.auth.models import AbstractUser
from .services.debt_collection_service import DebtCollectionService

class Client(models.Model):
    client_id = models.CharField(max_length=255, unique=True)
    name = models.CharField(max_length=255)
    is_active = models.BooleanField(default=True)
    created_at = models.DateTimeField(auto_now_add=True)

class User(AbstractUser):
    client = models.ForeignKey(Client, on_delete=models.CASCADE, related_name='users')
    keycloak_id = models.CharField(max_length=255, unique=True, null=True, blank=True)
    is_provisioned_to_debt_collection = models.BooleanField(default=False)
    
    def provision_to_debt_collection(self):
        """Provision this user to the debt collection system"""
        service = DebtCollectionService()
        user_data = {
            'keycloak_id': self.keycloak_id,
            'email': self.email,
            'first_name': self.first_name,
            'last_name': self.last_name,
            'engagemx_client_id': self.client.client_id,
            'is_active': self.is_active
        }
        
        try:
            result = service.provision_user(user_data)
            self.is_provisioned_to_debt_collection = True
            self.save()
            return result
        except Exception as e:
            raise Exception(f"Failed to provision user: {e}")
```

### 4. Django Admin Integration
```python
# admin.py
from django.contrib import admin
from django.contrib import messages
from django.http import HttpResponseRedirect
from .models import User, Client
from .services.debt_collection_service import DebtCollectionService

@admin.register(Client)
class ClientAdmin(admin.ModelAdmin):
    list_display = ['client_id', 'name', 'is_active', 'created_at']
    list_filter = ['is_active', 'created_at']
    search_fields = ['client_id', 'name']
    actions = ['sync_users_to_debt_collection']
    
    def sync_users_to_debt_collection(self, request, queryset):
        """Sync all users for selected clients to debt collection system"""
        service = DebtCollectionService()
        total_synced = 0
        
        for client in queryset:
            users_data = []
            for user in client.users.filter(is_active=True):
                users_data.append({
                    'keycloak_id': user.keycloak_id,
                    'email': user.email,
                    'first_name': user.first_name,
                    'last_name': user.last_name,
                    'engagemx_client_id': client.client_id,
                    'is_active': user.is_active
                })
            
            if users_data:
                try:
                    result = service.bulk_provision_users(users_data)
                    total_synced += result.get('success_count', 0)
                    
                    # Update provisioning status
                    client.users.filter(is_active=True).update(
                        is_provisioned_to_debt_collection=True
                    )
                except Exception as e:
                    messages.error(request, f"Failed to sync users for {client.name}: {e}")
        
        messages.success(request, f"Successfully synced {total_synced} users to debt collection system")
    
    sync_users_to_debt_collection.short_description = "Sync users to debt collection system"

@admin.register(User)
class UserAdmin(admin.ModelAdmin):
    list_display = ['username', 'email', 'client', 'is_active', 'is_provisioned_to_debt_collection']
    list_filter = ['is_active', 'is_provisioned_to_debt_collection', 'client']
    search_fields = ['username', 'email', 'first_name', 'last_name']
    actions = ['provision_to_debt_collection', 'deactivate_in_debt_collection']
    
    def provision_to_debt_collection(self, request, queryset):
        """Provision selected users to debt collection system"""
        service = DebtCollectionService()
        users_data = []
        
        for user in queryset:
            if user.keycloak_id:
                users_data.append({
                    'keycloak_id': user.keycloak_id,
                    'email': user.email,
                    'first_name': user.first_name,
                    'last_name': user.last_name,
                    'engagemx_client_id': user.client.client_id,
                    'is_active': user.is_active
                })
        
        if users_data:
            try:
                result = service.bulk_provision_users(users_data)
                success_count = result.get('success_count', 0)
                
                # Update provisioning status for successful users
                queryset.update(is_provisioned_to_debt_collection=True)
                
                messages.success(request, f"Successfully provisioned {success_count} users")
                
                if result.get('errors'):
                    for error in result['errors']:
                        messages.warning(request, f"Error: {error}")
                        
            except Exception as e:
                messages.error(request, f"Failed to provision users: {e}")
        else:
            messages.warning(request, "No users with Keycloak IDs found")
    
    provision_to_debt_collection.short_description = "Provision users to debt collection"
    
    def deactivate_in_debt_collection(self, request, queryset):
        """Deactivate selected users in debt collection system"""
        service = DebtCollectionService()
        deactivated_count = 0
        
        for user in queryset:
            if user.keycloak_id:
                try:
                    service.deactivate_user(user.keycloak_id)
                    deactivated_count += 1
                except Exception as e:
                    messages.error(request, f"Failed to deactivate {user.username}: {e}")
        
        messages.success(request, f"Successfully deactivated {deactivated_count} users in debt collection system")
    
    deactivate_in_debt_collection.short_description = "Deactivate in debt collection system"
```

### 5. Django Management Command
Create a management command for bulk synchronization:

```python
# management/commands/sync_debt_collection_users.py
from django.core.management.base import BaseCommand
from myapp.models import User, Client
from myapp.services.debt_collection_service import DebtCollectionService

class Command(BaseCommand):
    help = 'Sync users to debt collection system'
    
    def add_arguments(self, parser):
        parser.add_argument('--client-id', help='Sync users for specific client only')
        parser.add_argument('--dry-run', action='store_true', help='Show what would be synced without making changes')
    
    def handle(self, *args, **options):
        service = DebtCollectionService()
        
        if options['client_id']:
            clients = Client.objects.filter(client_id=options['client_id'])
        else:
            clients = Client.objects.filter(is_active=True)
        
        total_users = 0
        total_synced = 0
        
        for client in clients:
            users = client.users.filter(is_active=True, keycloak_id__isnull=False)
            total_users += users.count()
            
            if options['dry_run']:
                self.stdout.write(f"Would sync {users.count()} users for client {client.client_id}")
                continue
            
            users_data = []
            for user in users:
                users_data.append({
                    'keycloak_id': user.keycloak_id,
                    'email': user.email,
                    'first_name': user.first_name,
                    'last_name': user.last_name,
                    'engagemx_client_id': client.client_id,
                    'is_active': user.is_active
                })
            
            if users_data:
                try:
                    result = service.bulk_provision_users(users_data)
                    success_count = result.get('success_count', 0)
                    total_synced += success_count
                    
                    # Update provisioning status
                    users.update(is_provisioned_to_debt_collection=True)
                    
                    self.stdout.write(
                        self.style.SUCCESS(f"Synced {success_count} users for client {client.client_id}")
                    )
                    
                    if result.get('errors'):
                        for error in result['errors']:
                            self.stdout.write(self.style.WARNING(f"Warning: {error}"))
                            
                except Exception as e:
                    self.stdout.write(
                        self.style.ERROR(f"Failed to sync users for client {client.client_id}: {e}")
                    )
        
        if options['dry_run']:
            self.stdout.write(f"Dry run complete. Would sync {total_users} total users")
        else:
            self.stdout.write(f"Sync complete. Successfully synced {total_synced} of {total_users} users")
```

## Usage Examples

### 1. Manual User Provisioning
```bash
# In Django shell
from myapp.models import User
from myapp.services.debt_collection_service import DebtCollectionService

user = User.objects.get(email='user@example.com')
result = user.provision_to_debt_collection()
print(result)
```

### 2. Bulk Sync via Management Command
```bash
# Sync all users
python manage.py sync_debt_collection_users

# Sync users for specific client
python manage.py sync_debt_collection_users --client-id="medical_group_1"

# Dry run to see what would be synced
python manage.py sync_debt_collection_users --dry-run
```

### 3. Django Admin Interface
1. Go to Django Admin
2. Select Users or Clients
3. Choose "Provision users to debt collection" action
4. Click "Go" to execute

## Security Considerations

1. **API Key Management**: Store the API key securely using Django's secret management
2. **HTTPS**: Always use HTTPS in production for API communication
3. **Rate Limiting**: Implement rate limiting for bulk operations
4. **Logging**: Log all provisioning operations for audit trails
5. **Error Handling**: Implement proper error handling and retry logic

## Monitoring and Maintenance

1. **Health Checks**: Regularly test API connectivity
2. **Sync Status**: Monitor provisioning status in Django Admin
3. **Error Alerts**: Set up alerts for provisioning failures
4. **Data Validation**: Validate user data before provisioning
5. **Backup Strategy**: Maintain sync logs for recovery purposes