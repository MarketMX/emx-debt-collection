import { useKeycloak } from '@react-keycloak/web';
import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Header } from '@/components/layout/header';
import { FileUpload } from '@/components/upload/file-upload';
import { AccountsTable } from '@/components/accounts/accounts-table';
import { RecentUploads } from '@/components/dashboard/recent-uploads';
import { Reports } from '@/components/reports/reports';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { api } from '@/lib/api';
import type { Upload } from '@/types';
import { FileText, MessageSquare, TrendingUp, BarChart3, Plus, Users, Clock } from 'lucide-react';

export function Dashboard() {
  const { keycloak, initialized } = useKeycloak();
  const [selectedUpload, setSelectedUpload] = useState<Upload | null>(null);
  const [showReports, setShowReports] = useState(false);
  const [showUploadDialog, setShowUploadDialog] = useState(false);

  const { data, isLoading } = useQuery({
    queryKey: ['uploads'],
    queryFn: async () => {
      try {
        const res = await api.uploads.list();
        console.log('API Response:', res.data);
        return res.data;
      } catch (error) {
        console.error('Failed to fetch uploads:', error);
        throw error;
      }
    },
    enabled: initialized && !!keycloak?.authenticated,
    retry: 1,
  });

  // Safely handle uploads data with fallbacks
  const uploads = Array.isArray(data?.uploads) ? data.uploads : [];
  const recentUploads = uploads.slice(0, 5);
  const totalUploads = uploads.length;
  const completedUploads = uploads.filter((u: Upload) => u.status === 'completed').length;
  const processingUploads = uploads.filter((u: Upload) => u.status === 'processing').length;

  // Get the latest completed upload to show accounts for
  const latestCompletedUpload = uploads.find((u: Upload) => u.status === 'completed') || uploads[0];
  
  // Auto-select the latest upload if none is selected and we have uploads
  if (!selectedUpload && latestCompletedUpload) {
    setSelectedUpload(latestCompletedUpload);
  }

  if (showReports) {
    return <Reports onBack={() => setShowReports(false)} />;
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-blue-50/30">
      <Header />
      
      <main className="max-w-7xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
        {/* Always show accounts front and center */}
        <div className="space-y-6">
          {/* Header */}
          <div className="bg-white rounded-2xl shadow-sm border border-slate-200/60 overflow-hidden">
            {/* Top Section - Title and Icon */}
            <div className="bg-gradient-to-r from-blue-50 to-cyan-50 p-6 border-b border-slate-200/60">
              <div className="flex items-center space-x-4">
                <div className="w-12 h-12 bg-gradient-to-br from-blue-500 to-cyan-600 rounded-xl flex items-center justify-center shadow-lg">
                  <Users className="h-6 w-6 text-white" />
                </div>
                <div>
                  <h1 className="text-2xl font-bold text-slate-800">Account Management</h1>
                  <p className="text-slate-600 text-sm mt-1">
                    {selectedUpload ? `Managing accounts from: ${selectedUpload.filename}` : 'Select a file to manage accounts'}
                  </p>
                </div>
              </div>
            </div>

            {/* Bottom Section - Controls */}
            <div className="p-6">
              <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between space-y-4 lg:space-y-0">
                {/* File Selection */}
                <div className="flex-1 max-w-md">
                  {uploads.length > 0 && (
                    <div>
                      <label className="block text-xs font-medium text-slate-700 mb-2 uppercase tracking-wide">
                        Select Data Source
                      </label>
                      <select
                        value={selectedUpload?.id || ''}
                        onChange={(e) => {
                          const upload = uploads.find((u: Upload) => u.id === e.target.value);
                          setSelectedUpload(upload || null);
                        }}
                        className="w-full px-4 py-3 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white shadow-sm"
                      >
                        <option value="">Choose a file...</option>
                        {uploads.map((upload: Upload) => (
                          <option key={upload.id} value={upload.id}>
                            {upload.filename} • {upload.status}
                          </option>
                        ))}
                      </select>
                    </div>
                  )}
                </div>

                {/* Action Buttons */}
                <div className="flex items-center space-x-3">
                  <Button 
                    onClick={() => setShowUploadDialog(true)}
                    className="bg-gradient-to-r from-blue-500 to-cyan-600 hover:from-blue-600 hover:to-cyan-700 px-6 py-3 shadow-lg"
                  >
                    <Plus className="h-4 w-4 mr-2" />
                    Upload File
                  </Button>

                  <Button 
                    variant="outline" 
                    onClick={() => setShowReports(true)}
                    className="border-slate-300 hover:bg-slate-50 px-6 py-3 shadow-sm"
                  >
                    <BarChart3 className="h-4 w-4 mr-2" />
                    View Analytics
                  </Button>
                </div>
              </div>
            </div>
          </div>

          {/* Quick Stats Overview */}
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
            <Card className="bg-white border-slate-200/60 shadow-sm hover:shadow-md transition-shadow duration-200">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
                <div>
                  <CardTitle className="text-sm font-medium text-slate-600">Total Files</CardTitle>
                  <div className="text-2xl font-bold text-slate-800 mt-2">{totalUploads}</div>
                </div>
                <div className="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center">
                  <FileText className="h-6 w-6 text-blue-600" />
                </div>
              </CardHeader>
            </Card>

            <Card className="bg-white border-slate-200/60 shadow-sm hover:shadow-md transition-shadow duration-200">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
                <div>
                  <CardTitle className="text-sm font-medium text-slate-600">Completed</CardTitle>
                  <div className="text-2xl font-bold text-emerald-600 mt-2">{completedUploads}</div>
                </div>
                <div className="w-12 h-12 bg-emerald-100 rounded-xl flex items-center justify-center">
                  <TrendingUp className="h-6 w-6 text-emerald-600" />
                </div>
              </CardHeader>
            </Card>

            <Card className="bg-white border-slate-200/60 shadow-sm hover:shadow-md transition-shadow duration-200">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
                <div>
                  <CardTitle className="text-sm font-medium text-slate-600">Processing</CardTitle>
                  <div className="text-2xl font-bold text-amber-600 mt-2">{processingUploads}</div>
                </div>
                <div className="w-12 h-12 bg-amber-100 rounded-xl flex items-center justify-center">
                  <Clock className="h-6 w-6 text-amber-600" />
                </div>
              </CardHeader>
            </Card>

            <Card className="bg-white border-slate-200/60 shadow-sm hover:shadow-md transition-shadow duration-200">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
                <div>
                  <CardTitle className="text-sm font-medium text-slate-600">Messages Sent</CardTitle>
                  <div className="text-2xl font-bold text-cyan-600 mt-2">-</div>
                </div>
                <div className="w-12 h-12 bg-cyan-100 rounded-xl flex items-center justify-center">
                  <MessageSquare className="h-6 w-6 text-cyan-600" />
                </div>
              </CardHeader>
            </Card>
          </div>

          {/* Main Content Area */}
          {selectedUpload && selectedUpload.status === 'completed' ? (
            /* Show Accounts Table - FRONT AND CENTER */
            <AccountsTable uploadId={selectedUpload.id} />
          ) : uploads.length === 0 ? (
            /* First Time User - Upload CTA */
            <Card className="bg-white border-slate-200/60 shadow-sm">
              <CardContent className="py-16">
                <div className="text-center">
                  <div className="w-20 h-20 bg-gradient-to-br from-blue-500 to-cyan-600 rounded-2xl mx-auto mb-6 flex items-center justify-center">
                    <Plus className="h-10 w-10 text-white" />
                  </div>
                  <h3 className="text-2xl font-bold text-slate-800 mb-3">Get Started</h3>
                  <p className="text-slate-600 mb-8 max-w-md mx-auto">
                    Upload your first age analysis Excel file to start managing customer accounts and sending collection messages.
                  </p>
                  <Button 
                    onClick={() => setShowUploadDialog(true)}
                    size="lg"
                    className="bg-gradient-to-r from-blue-500 to-cyan-600 hover:from-blue-600 hover:to-cyan-700"
                  >
                    <Plus className="h-5 w-5 mr-2" />
                    Upload Your First File
                  </Button>
                </div>
              </CardContent>
            </Card>
          ) : (
            /* File Selection State */
            <Card className="bg-white border-slate-200/60 shadow-sm">
              <CardContent className="py-12">
                <div className="text-center">
                  <div className="w-16 h-16 bg-slate-100 rounded-2xl mx-auto mb-4 flex items-center justify-center">
                    <FileText className="h-8 w-8 text-slate-400" />
                  </div>
                  <h3 className="text-xl font-semibold text-slate-800 mb-2">Select a File</h3>
                  <p className="text-slate-600 mb-6">
                    Choose a completed file from the dropdown above to view and manage its accounts.
                  </p>
                  {processingUploads > 0 && (
                    <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 max-w-md mx-auto">
                      <div className="flex items-center space-x-2 text-amber-700">
                        <Clock className="h-4 w-4" />
                        <span className="text-sm font-medium">{processingUploads} files are still processing...</span>
                      </div>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          )}

          {/* Upload Dialog Modal */}
          {showUploadDialog && (
            <Card className="bg-white border-slate-200/60 shadow-lg">
              <CardHeader className="pb-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-3">
                    <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-cyan-600 rounded-lg flex items-center justify-center">
                      <Plus className="h-5 w-5 text-white" />
                    </div>
                    <div>
                      <CardTitle className="text-lg text-slate-800">Upload Age Analysis</CardTitle>
                      <p className="text-sm text-slate-600">Drop your Excel file to start processing accounts</p>
                    </div>
                  </div>
                  <Button
                    variant="outline"
                    onClick={() => setShowUploadDialog(false)}
                    className="border-slate-200 hover:bg-slate-50"
                  >
                    Close
                  </Button>
                </div>
              </CardHeader>
              <CardContent>
                <FileUpload />
              </CardContent>
            </Card>
          )}

          {/* Recent Files - Collapsible */}
          {uploads.length > 0 && (
            <Card className="bg-white border-slate-200/60 shadow-sm">
              <CardHeader className="pb-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-3">
                    <div className="w-10 h-10 bg-slate-100 rounded-lg flex items-center justify-center">
                      <FileText className="h-5 w-5 text-slate-600" />
                    </div>
                    <div>
                      <CardTitle className="text-lg text-slate-800">Recent Files</CardTitle>
                      <p className="text-sm text-slate-600">Your uploaded analysis files</p>
                    </div>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <RecentUploads 
                  uploads={recentUploads} 
                  onSelectUpload={setSelectedUpload}
                  isLoading={isLoading}
                />
              </CardContent>
            </Card>
          )}
        </div>
      </main>
    </div>
  );
}