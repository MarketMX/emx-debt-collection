#!/usr/bin/env python3
"""
Django Admin Setup Example for Debt Collection Integration

This script demonstrates how to set up Django admin integration
with the debt collection application for user provisioning.

Run this in your Django project directory after installing requirements.
"""

import os
import sys
import requests
from datetime import datetime
from typing import Dict, List

class DebtCollectionSetup:
    def __init__(self, api_base_url: str = "http://localhost:8080/api", api_key: str = None):
        self.api_base_url = api_base_url
        self.api_key = api_key or os.getenv('PROVISIONING_API_KEY')
        self.headers = {
            'Authorization': f'Bearer {self.api_key}',
            'Content-Type': 'application/json'
        }
    
    def test_connection(self) -> bool:
        """Test connection to debt collection API"""
        try:
            response = requests.get(f"{self.api_base_url}/webhooks/status", headers=self.headers, timeout=10)
            if response.status_code == 200:
                print("✅ Successfully connected to debt collection API")
                print(f"Webhook system status: {response.json()}")
                return True
            else:
                print(f"❌ Connection failed with status: {response.status_code}")
                return False
        except requests.RequestException as e:
            print(f"❌ Connection error: {e}")
            return False
    
    def provision_sample_users(self) -> bool:
        """Provision sample users for testing"""
        sample_users = [
            {
                "keycloak_id": "test-user-1-uuid",
                "email": "admin@medical-group-1.com",
                "first_name": "Dr. John",
                "last_name": "Doe",
                "engagemx_client_id": "medical_group_1",
                "is_active": True
            },
            {
                "keycloak_id": "test-user-2-uuid", 
                "email": "manager@hospital-2.com",
                "first_name": "Jane",
                "last_name": "Smith",
                "engagemx_client_id": "hospital_2",
                "is_active": True
            }
        ]
        
        try:
            response = requests.post(
                f"{self.api_base_url}/provisioning/users/bulk",
                json=sample_users,
                headers=self.headers,
                timeout=30
            )
            
            if response.status_code == 200:
                result = response.json()
                print(f"✅ Successfully provisioned {result['success_count']} sample users")
                if result.get('errors'):
                    print(f"⚠️  {result['error_count']} errors occurred")
                return True
            else:
                print(f"❌ Provisioning failed with status: {response.status_code}")
                print(response.text)
                return False
                
        except requests.RequestException as e:
            print(f"❌ Provisioning error: {e}")
            return False
    
    def send_webhook_test(self) -> bool:
        """Send a test webhook event"""
        test_event = {
            "event_type": "user.created",
            "event_id": f"test-{datetime.now().isoformat()}",
            "timestamp": datetime.now().isoformat(),
            "source": "django-admin-test",
            "data": {
                "keycloak_id": "webhook-test-user-uuid",
                "email": "webhook-test@example.com",
                "first_name": "Webhook",
                "last_name": "Test",
                "engagemx_client_id": "test_client",
                "is_active": True
            },
            "version": "1.0.0"
        }
        
        try:
            response = requests.post(
                f"{self.api_base_url}/webhooks/events",
                json=test_event,
                headers=self.headers,
                timeout=30
            )
            
            if response.status_code == 200:
                result = response.json()
                print("✅ Webhook test successful")
                print(f"Event ID: {result['event_id']}")
                return True
            else:
                print(f"❌ Webhook test failed with status: {response.status_code}")
                return False
                
        except requests.RequestException as e:
            print(f"❌ Webhook test error: {e}")
            return False
    
    def generate_django_models(self) -> str:
        """Generate Django model code for integration"""
        return """
# Add this to your Django models.py

from django.db import models
from django.contrib.auth.models import AbstractUser

class Client(models.Model):
    \"\"\"Represents a medical client/organization\"\"\"
    client_id = models.CharField(max_length=255, unique=True, help_text="Unique client identifier for debt collection system")
    name = models.CharField(max_length=255)
    contact_email = models.EmailField(blank=True)
    is_active = models.BooleanField(default=True)
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)
    
    class Meta:
        ordering = ['name']
    
    def __str__(self):
        return f"{self.name} ({self.client_id})"

class User(AbstractUser):
    \"\"\"Extended user model with client association\"\"\"
    client = models.ForeignKey(
        Client, 
        on_delete=models.CASCADE, 
        related_name='users',
        help_text="Client this user belongs to"
    )
    keycloak_id = models.CharField(
        max_length=255, 
        unique=True, 
        null=True, 
        blank=True,
        help_text="Keycloak user ID for authentication"
    )
    is_provisioned_to_debt_collection = models.BooleanField(
        default=False,
        help_text="Whether user has been synced to debt collection system"
    )
    provisioned_at = models.DateTimeField(null=True, blank=True)
    
    class Meta:
        ordering = ['client__name', 'username']
    
    def __str__(self):
        return f"{self.username} - {self.client.name}"
    
    def get_debt_collection_data(self):
        \"\"\"Get user data formatted for debt collection API\"\"\"
        return {
            'keycloak_id': self.keycloak_id,
            'email': self.email,
            'first_name': self.first_name,
            'last_name': self.last_name,
            'engagemx_client_id': self.client.client_id,
            'is_active': self.is_active
        }
"""

    def generate_django_admin(self) -> str:
        """Generate Django admin code for integration"""
        return """
# Add this to your Django admin.py

from django.contrib import admin
from django.contrib import messages
from django.utils import timezone
from .models import User, Client
from .services.debt_collection_service import DebtCollectionService

@admin.register(Client)
class ClientAdmin(admin.ModelAdmin):
    list_display = ['client_id', 'name', 'user_count', 'is_active', 'created_at']
    list_filter = ['is_active', 'created_at']
    search_fields = ['client_id', 'name', 'contact_email']
    readonly_fields = ['created_at', 'updated_at']
    actions = ['sync_all_users_to_debt_collection']
    
    def user_count(self, obj):
        return obj.users.count()
    user_count.short_description = 'Users'
    
    def sync_all_users_to_debt_collection(self, request, queryset):
        service = DebtCollectionService()
        total_synced = 0
        
        for client in queryset:
            users = client.users.filter(is_active=True, keycloak_id__isnull=False)
            if users.exists():
                users_data = [user.get_debt_collection_data() for user in users]
                
                try:
                    result = service.bulk_provision_users(users_data)
                    success_count = result.get('success_count', 0)
                    total_synced += success_count
                    
                    # Update provisioning status
                    users.update(
                        is_provisioned_to_debt_collection=True,
                        provisioned_at=timezone.now()
                    )
                    
                except Exception as e:
                    messages.error(request, f"Failed to sync users for {client.name}: {e}")
        
        messages.success(request, f"Successfully synced {total_synced} users to debt collection system")
    
    sync_all_users_to_debt_collection.short_description = "Sync all users to debt collection"

@admin.register(User)
class UserAdmin(admin.ModelAdmin):
    list_display = ['username', 'email', 'client', 'is_active', 'is_provisioned_to_debt_collection', 'provisioned_at']
    list_filter = ['is_active', 'is_provisioned_to_debt_collection', 'client', 'provisioned_at']
    search_fields = ['username', 'email', 'first_name', 'last_name']
    readonly_fields = ['provisioned_at']
    actions = ['provision_to_debt_collection', 'sync_updates_to_debt_collection']
    
    fieldsets = (
        (None, {
            'fields': ('username', 'password')
        }),
        ('Personal info', {
            'fields': ('first_name', 'last_name', 'email')
        }),
        ('Client & Permissions', {
            'fields': ('client', 'is_active', 'is_staff', 'is_superuser', 'groups', 'user_permissions')
        }),
        ('Authentication', {
            'fields': ('keycloak_id', 'last_login', 'date_joined')
        }),
        ('Debt Collection Integration', {
            'fields': ('is_provisioned_to_debt_collection', 'provisioned_at'),
            'classes': ('collapse',)
        }),
    )
    
    def provision_to_debt_collection(self, request, queryset):
        service = DebtCollectionService()
        users_data = []
        
        for user in queryset:
            if user.keycloak_id:
                users_data.append(user.get_debt_collection_data())
        
        if users_data:
            try:
                result = service.bulk_provision_users(users_data)
                success_count = result.get('success_count', 0)
                
                # Update provisioning status
                queryset.filter(keycloak_id__isnull=False).update(
                    is_provisioned_to_debt_collection=True,
                    provisioned_at=timezone.now()
                )
                
                messages.success(request, f"Successfully provisioned {success_count} users")
                
                if result.get('errors'):
                    messages.warning(request, f"{result['error_count']} errors occurred")
                    
            except Exception as e:
                messages.error(request, f"Failed to provision users: {e}")
        else:
            messages.warning(request, "No users with Keycloak IDs found")
    
    provision_to_debt_collection.short_description = "Provision to debt collection system"
"""

def main():
    print("🏥 Django Admin - Debt Collection Integration Setup")
    print("=" * 55)
    
    # Check for API key
    api_key = os.getenv('PROVISIONING_API_KEY')
    if not api_key:
        print("❌ PROVISIONING_API_KEY environment variable not set")
        print("Please set it with: export PROVISIONING_API_KEY='your-api-key'")
        return False
    
    # Initialize setup
    setup = DebtCollectionSetup(api_key=api_key)
    
    print("1. Testing API connection...")
    if not setup.test_connection():
        print("Please ensure the debt collection API is running and accessible")
        return False
    
    print("\\n2. Provisioning sample users...")
    if not setup.provision_sample_users():
        print("Sample user provisioning failed")
        return False
    
    print("\\n3. Testing webhook functionality...")
    if not setup.send_webhook_test():
        print("Webhook test failed")
        return False
    
    print("\\n4. Generating Django integration code...")
    
    # Save generated code to files
    with open('django_models_example.py', 'w') as f:
        f.write(setup.generate_django_models())
    
    with open('django_admin_example.py', 'w') as f:
        f.write(setup.generate_django_admin())
    
    print("✅ Setup complete!")
    print("\\nGenerated files:")
    print("- django_models_example.py (model definitions)")
    print("- django_admin_example.py (admin integration)")
    print("\\nNext steps:")
    print("1. Copy the model code to your Django models.py")
    print("2. Copy the admin code to your Django admin.py")
    print("3. Run Django migrations: python manage.py makemigrations && python manage.py migrate")
    print("4. Create a Django superuser: python manage.py createsuperuser")
    print("5. Install the DebtCollectionService from the documentation")
    print("6. Configure your PROVISIONING_API_KEY in Django settings")
    
    return True

if __name__ == '__main__':
    success = main()
    sys.exit(0 if success else 1)