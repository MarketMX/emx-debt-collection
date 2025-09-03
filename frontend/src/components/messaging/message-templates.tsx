import { useState } from 'react';
import { Header } from '@/components/layout/header';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { useToast } from '@/hooks/use-toast';
import { Plus, Edit3, Trash2, MessageSquare } from 'lucide-react';

interface MessageTemplate {
  id: string;
  name: string;
  content: string;
  is_default: boolean;
  created_at: string;
}

interface MessageTemplatesProps {
  onSelectTemplate?: (template: MessageTemplate) => void;
  selectedTemplateId?: string;
}

export function MessageTemplates({ onSelectTemplate, selectedTemplateId }: MessageTemplatesProps) {
  const [isCreating, setIsCreating] = useState(false);
  const [editingTemplate, setEditingTemplate] = useState<MessageTemplate | null>(null);
  const [newTemplate, setNewTemplate] = useState({ name: '', content: '' });
  
  const { toast } = useToast();

  // For demo purposes, using mock data since the API endpoint isn't implemented yet
  const templates: MessageTemplate[] = [
    {
      id: '1',
      name: 'Standard Payment Reminder',
      content: 'Hi [Name], this is a reminder regarding account [Account] for an outstanding balance of [Total Balance]. Please contact us.',
      is_default: true,
      created_at: new Date().toISOString()
    },
    {
      id: '2',
      name: '60+ Days Overdue',
      content: 'Dear [Name], account [Account] is now 60+ days overdue with a balance of [Total Balance]. Immediate payment is required to avoid further action.',
      is_default: false,
      created_at: new Date().toISOString()
    }
  ];

  const handleCreateTemplate = () => {
    if (newTemplate.name.trim() && newTemplate.content.trim()) {
      toast({
        title: "Template created",
        description: `Template "${newTemplate.name}" has been created successfully.`,
      });
      setNewTemplate({ name: '', content: '' });
      setIsCreating(false);
    }
  };

  const handleSelectTemplate = (template: MessageTemplate) => {
    if (onSelectTemplate) {
      onSelectTemplate(template);
    }
  };

  const availableVariables = [
    '[Name]', '[Account]', '[Total Balance]', '[Current]', 
    '[30 Days]', '[60 Days]', '[90 Days]', '[120 Days]', '[Contact]'
  ];

  return (
    <div className="min-h-screen bg-gray-50">
      <Header />
      
      <main className="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8">
        <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Message Templates</h2>
          <p className="text-gray-600">
            Manage your message templates for different scenarios
          </p>
        </div>
        <Button onClick={() => setIsCreating(true)}>
          <Plus className="h-4 w-4 mr-2" />
          New Template
        </Button>
      </div>

      {/* Available Variables */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Available Variables</CardTitle>
          <CardDescription>
            Use these placeholders in your message templates - they will be replaced with actual account data
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-2">
            {availableVariables.map((variable) => (
              <Badge key={variable} variant="secondary" className="font-mono">
                {variable}
              </Badge>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Create/Edit Template Form */}
      {(isCreating || editingTemplate) && (
        <Card>
          <CardHeader>
            <CardTitle>
              {isCreating ? 'Create New Template' : 'Edit Template'}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Template Name
              </label>
              <Input
                value={editingTemplate?.name || newTemplate.name}
                onChange={(e) => {
                  if (editingTemplate) {
                    setEditingTemplate({ ...editingTemplate, name: e.target.value });
                  } else {
                    setNewTemplate({ ...newTemplate, name: e.target.value });
                  }
                }}
                placeholder="Enter template name..."
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Message Content
              </label>
              <textarea
                className="w-full min-h-32 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                value={editingTemplate?.content || newTemplate.content}
                onChange={(e) => {
                  if (editingTemplate) {
                    setEditingTemplate({ ...editingTemplate, content: e.target.value });
                  } else {
                    setNewTemplate({ ...newTemplate, content: e.target.value });
                  }
                }}
                placeholder="Enter message content using variables like [Name], [Account], [Total Balance]..."
              />
            </div>
            <div className="flex space-x-2">
              <Button onClick={handleCreateTemplate}>
                <MessageSquare className="h-4 w-4 mr-2" />
                {isCreating ? 'Create Template' : 'Save Changes'}
              </Button>
              <Button 
                variant="outline" 
                onClick={() => {
                  setIsCreating(false);
                  setEditingTemplate(null);
                  setNewTemplate({ name: '', content: '' });
                }}
              >
                Cancel
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Templates List */}
      <div className="grid gap-4">
        {templates.map((template) => (
          <Card 
            key={template.id}
            className={selectedTemplateId === template.id ? 'ring-2 ring-blue-500' : ''}
          >
            <CardHeader>
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <CardTitle className="text-lg">{template.name}</CardTitle>
                  {template.is_default && (
                    <Badge variant="default">Default</Badge>
                  )}
                </div>
                <div className="flex space-x-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setEditingTemplate(template)}
                  >
                    <Edit3 className="h-4 w-4" />
                  </Button>
                  {!template.is_default && (
                    <Button
                      variant="outline"
                      size="sm"
                      className="text-red-600 hover:text-red-700"
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  )}
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <p className="text-gray-700 mb-4 whitespace-pre-wrap">
                {template.content}
              </p>
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-500">
                  Created: {new Date(template.created_at).toLocaleDateString()}
                </span>
                <Button 
                  variant={selectedTemplateId === template.id ? "default" : "outline"}
                  onClick={() => handleSelectTemplate(template)}
                >
                  {selectedTemplateId === template.id ? 'Selected' : 'Select'}
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}
        </div>
        </div>
      </main>
    </div>
  );
}