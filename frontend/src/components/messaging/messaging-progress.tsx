import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Progress } from '@/components/ui/progress';
import { Badge } from '@/components/ui/badge';
import type { MessageJob } from '@/types';
import { 
  MessageSquare, 
  CheckCircle, 
  AlertCircle, 
  Clock, 
  Loader2,
  RefreshCw,
  X
} from 'lucide-react';

interface MessagingProgressProps {
  jobId?: string;
  onClose?: () => void;
}

export function MessagingProgress({ jobId, onClose }: MessagingProgressProps) {
  const [currentJob, setCurrentJob] = useState<MessageJob | null>(null);

  // Simulate real-time progress updates
  useEffect(() => {
    if (!jobId) return;

    // Mock data for demonstration since real-time API isn't implemented
    const mockJob: MessageJob = {
      id: jobId,
      user_id: 'user1',
      total_accounts: 25,
      successful_sends: 18,
      failed_sends: 2,
      status: 'in_progress',
      created_at: new Date().toISOString(),
    };

    setCurrentJob(mockJob);

    const interval = setInterval(() => {
      setCurrentJob((prev) => {
        if (!prev || prev.status === 'completed' || prev.status === 'failed') {
          return prev;
        }

        const newSuccessful = Math.min(prev.successful_sends + Math.floor(Math.random() * 3), prev.total_accounts);
        const newFailed = Math.min(prev.failed_sends + (Math.random() > 0.8 ? 1 : 0), prev.total_accounts - newSuccessful);
        const processed = newSuccessful + newFailed;
        
        return {
          ...prev,
          successful_sends: newSuccessful,
          failed_sends: newFailed,
          status: processed >= prev.total_accounts ? 'completed' : 'in_progress',
          completed_at: processed >= prev.total_accounts ? new Date().toISOString() : undefined,
        };
      });
    }, 2000);

    return () => clearInterval(interval);
  }, [jobId]);

  if (!currentJob) {
    return (
      <Card className="max-w-2xl mx-auto">
        <CardContent className="flex items-center justify-center py-8">
          <div className="text-center text-gray-500">
            <MessageSquare className="h-12 w-12 mx-auto mb-4 text-gray-400" />
            <p>No messaging job in progress</p>
          </div>
        </CardContent>
      </Card>
    );
  }

  const progress = currentJob.total_accounts > 0 
    ? ((currentJob.successful_sends + currentJob.failed_sends) / currentJob.total_accounts) * 100
    : 0;

  const successRate = currentJob.successful_sends + currentJob.failed_sends > 0
    ? (currentJob.successful_sends / (currentJob.successful_sends + currentJob.failed_sends)) * 100
    : 0;

  const StatusIcon = () => {
    switch (currentJob.status) {
      case 'pending':
        return <Clock className="h-5 w-5 text-yellow-500" />;
      case 'in_progress':
        return <Loader2 className="h-5 w-5 text-blue-500 animate-spin" />;
      case 'completed':
        return <CheckCircle className="h-5 w-5 text-green-500" />;
      case 'failed':
        return <AlertCircle className="h-5 w-5 text-red-500" />;
      default:
        return <MessageSquare className="h-5 w-5 text-gray-400" />;
    }
  };

  return (
    <Card className="max-w-2xl mx-auto">
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <StatusIcon />
            <div>
              <CardTitle>Messaging Progress</CardTitle>
              <CardDescription>
                Job started {new Date(currentJob.created_at).toLocaleString()}
              </CardDescription>
            </div>
          </div>
          {onClose && (
            <Button variant="outline" size="sm" onClick={onClose}>
              <X className="h-4 w-4" />
            </Button>
          )}
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Progress Bar */}
        <div className="space-y-2">
          <div className="flex items-center justify-between text-sm">
            <span className="font-medium">Overall Progress</span>
            <span>{Math.round(progress)}% Complete</span>
          </div>
          <Progress value={progress} className="h-3" />
          <div className="flex items-center justify-between text-xs text-gray-500">
            <span>{currentJob.successful_sends + currentJob.failed_sends} of {currentJob.total_accounts} processed</span>
            <span>
              {currentJob.status === 'completed' ? 'Completed' : 
               currentJob.status === 'in_progress' ? 'In Progress...' : 
               currentJob.status}
            </span>
          </div>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="text-center">
            <div className="text-2xl font-bold text-blue-600">{currentJob.total_accounts}</div>
            <div className="text-xs text-gray-500 uppercase tracking-wide">Total</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-green-600">{currentJob.successful_sends}</div>
            <div className="text-xs text-gray-500 uppercase tracking-wide">Sent</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-red-600">{currentJob.failed_sends}</div>
            <div className="text-xs text-gray-500 uppercase tracking-wide">Failed</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-gray-600">
              {Math.round(successRate)}%
            </div>
            <div className="text-xs text-gray-500 uppercase tracking-wide">Success Rate</div>
          </div>
        </div>

        {/* Status Badge */}
        <div className="flex items-center justify-center">
          <Badge 
            variant={
              currentJob.status === 'completed' ? 'default' :
              currentJob.status === 'failed' ? 'destructive' :
              'secondary'
            }
            className="px-4 py-2"
          >
            {currentJob.status === 'pending' && 'Pending'}
            {currentJob.status === 'in_progress' && 'Processing Messages...'}
            {currentJob.status === 'completed' && 'All Messages Sent'}
            {currentJob.status === 'failed' && 'Job Failed'}
          </Badge>
        </div>

        {/* Actions */}
        {currentJob.status === 'completed' && (
          <div className="flex items-center justify-center space-x-4 pt-4 border-t">
            <Button variant="outline" size="sm">
              <RefreshCw className="h-4 w-4 mr-2" />
              View Details
            </Button>
            <Button variant="outline" size="sm">
              Download Report
            </Button>
          </div>
        )}

        {/* Real-time updates indicator */}
        {currentJob.status === 'in_progress' && (
          <div className="text-center text-xs text-gray-500">
            <div className="flex items-center justify-center space-x-1">
              <div className="w-2 h-2 bg-blue-500 rounded-full animate-pulse"></div>
              <span>Live updates every 2 seconds</span>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}