import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Header } from '@/components/layout/header';
import { FileUpload } from '@/components/upload/file-upload';
import { AccountsTable } from '@/components/accounts/accounts-table';
import { RecentUploads } from '@/components/dashboard/recent-uploads';
import { Reports } from '@/components/reports/reports';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { api } from '@/lib/api';
import type { Upload } from '@/types';
import { Upload as UploadIcon, FileText, MessageSquare, TrendingUp, BarChart3 } from 'lucide-react';

export function Dashboard() {
  const [selectedUpload, setSelectedUpload] = useState<Upload | null>(null);
  const [showReports, setShowReports] = useState(false);

  const { data: uploads, isLoading } = useQuery({
    queryKey: ['uploads'],
    queryFn: () => api.uploads.list().then(res => res.data),
  });

  const recentUploads = uploads?.slice(0, 5) || [];
  const totalUploads = uploads?.length || 0;
  const completedUploads = uploads?.filter((u: Upload) => u.status === 'completed').length || 0;
  const processingUploads = uploads?.filter((u: Upload) => u.status === 'processing').length || 0;

  if (showReports) {
    return <Reports onBack={() => setShowReports(false)} />;
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <Header />
      
      <main className="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8">
        {selectedUpload ? (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-2xl font-bold text-gray-900">
                  Upload Details: {selectedUpload.filename}
                </h2>
                <p className="text-gray-600">
                  Status: <span className="font-medium capitalize">{selectedUpload.status}</span>
                </p>
              </div>
              <button
                onClick={() => setSelectedUpload(null)}
                className="text-blue-600 hover:text-blue-800 font-medium"
              >
                ← Back to Dashboard
              </button>
            </div>
            
            <AccountsTable uploadId={selectedUpload.id} />
          </div>
        ) : (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <div>
                <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
                <p className="text-gray-600">
                  Upload age analysis reports and manage your debt collection campaigns
                </p>
              </div>
              
              <Button onClick={() => setShowReports(true)} variant="outline">
                <BarChart3 className="h-4 w-4 mr-2" />
                View Reports
              </Button>
            </div>

            {/* Stats Overview */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Total Uploads</CardTitle>
                  <FileText className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{totalUploads}</div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Completed</CardTitle>
                  <TrendingUp className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{completedUploads}</div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Processing</CardTitle>
                  <UploadIcon className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{processingUploads}</div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Messages Sent</CardTitle>
                  <MessageSquare className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">-</div>
                  <p className="text-xs text-muted-foreground">Coming soon</p>
                </CardContent>
              </Card>
            </div>

            {/* File Upload Section */}
            <Card>
              <CardHeader>
                <CardTitle>Upload Age Analysis Report</CardTitle>
                <CardDescription>
                  Upload an Excel (.xlsx) file containing your age analysis data
                </CardDescription>
              </CardHeader>
              <CardContent>
                <FileUpload />
              </CardContent>
            </Card>

            {/* Recent Uploads */}
            <Card>
              <CardHeader>
                <CardTitle>Recent Uploads</CardTitle>
                <CardDescription>
                  Your recently uploaded files and their processing status
                </CardDescription>
              </CardHeader>
              <CardContent>
                <RecentUploads 
                  uploads={recentUploads} 
                  onSelectUpload={setSelectedUpload}
                  isLoading={isLoading}
                />
              </CardContent>
            </Card>
          </div>
        )}
      </main>
    </div>
  );
}