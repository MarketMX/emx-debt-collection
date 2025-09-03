import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Header } from '@/components/layout/header';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { api } from '@/lib/api';
import type { MessageLog } from '@/types';
import { 
  Search, 
  MessageSquare, 
  CheckCircle, 
  AlertCircle,
  TrendingUp,
  Clock,
  ArrowLeft
} from 'lucide-react';

interface ReportsProps {
  onBack?: () => void;
}

export function Reports({ onBack }: ReportsProps) {
  const [searchTerm, setSearchTerm] = useState('');

  const { data: messageLogs = [], isLoading } = useQuery({
    queryKey: ['message-logs'],
    queryFn: () => api.messaging.getLogs().then(res => res.data),
  });

  const filteredLogs = messageLogs.filter((log: MessageLog) =>
    log.account?.customer_name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
    log.account?.account_code?.toLowerCase().includes(searchTerm.toLowerCase()) ||
    log.account?.telephone?.includes(searchTerm)
  );

  const stats = {
    total: messageLogs.length,
    successful: messageLogs.filter((log: MessageLog) => log.status === 'sent').length,
    failed: messageLogs.filter((log: MessageLog) => log.status === 'failed').length,
  };

  const successRate = stats.total > 0 ? (stats.successful / stats.total) * 100 : 0;

  return (
    <div className="min-h-screen bg-gray-50">
      <Header />
      
      <main className="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8">
        <div className="space-y-6">
          {/* Header */}
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">Reports & Analytics</h1>
              <p className="text-gray-600">
                Track messaging performance and view detailed logs
              </p>
            </div>
            {onBack && (
              <Button variant="outline" onClick={onBack}>
                <ArrowLeft className="h-4 w-4 mr-2" />
                Back to Dashboard
              </Button>
            )}
          </div>

          {/* Stats Overview */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Total Messages</CardTitle>
                <MessageSquare className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.total}</div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Successful</CardTitle>
                <CheckCircle className="h-4 w-4 text-green-500" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-green-600">{stats.successful}</div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Failed</CardTitle>
                <AlertCircle className="h-4 w-4 text-red-500" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-red-600">{stats.failed}</div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Success Rate</CardTitle>
                <TrendingUp className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{successRate.toFixed(1)}%</div>
              </CardContent>
            </Card>
          </div>

          {/* Search */}
          <div className="flex items-center space-x-2">
            <Search className="h-4 w-4 text-gray-400" />
            <Input
              placeholder="Search by customer name, account code, or phone..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-64"
            />
          </div>

          {/* Message Logs Table */}
          <Card>
            <CardHeader>
              <CardTitle>Message History</CardTitle>
              <CardDescription>
                Detailed log of all messaging activities
              </CardDescription>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <div className="flex items-center justify-center py-8">
                  <Clock className="h-8 w-8 animate-spin text-gray-400" />
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Date/Time</TableHead>
                      <TableHead>Account Code</TableHead>
                      <TableHead>Customer Name</TableHead>
                      <TableHead>Phone</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Response</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredLogs.length > 0 ? (
                      filteredLogs.map((log: MessageLog) => (
                        <TableRow key={log.id}>
                          <TableCell>
                            {new Date(log.sent_at).toLocaleString()}
                          </TableCell>
                          <TableCell className="font-medium">
                            {log.account?.account_code || 'N/A'}
                          </TableCell>
                          <TableCell>
                            {log.account?.customer_name || 'N/A'}
                          </TableCell>
                          <TableCell>
                            {log.account?.telephone || 'N/A'}
                          </TableCell>
                          <TableCell>
                            <Badge
                              variant={log.status === 'sent' ? 'default' : 'destructive'}
                            >
                              {log.status}
                            </Badge>
                          </TableCell>
                          <TableCell className="max-w-xs truncate">
                            {log.response_from_service || '-'}
                          </TableCell>
                        </TableRow>
                      ))
                    ) : (
                      <TableRow>
                        <TableCell colSpan={6} className="text-center py-8">
                          {searchTerm ? 'No messages match your search criteria.' : 'No messages sent yet.'}
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </div>
      </main>
    </div>
  );
}