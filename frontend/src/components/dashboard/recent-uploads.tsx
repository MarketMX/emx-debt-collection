import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import type { Upload } from '@/types';
import { FileText, Clock, CheckCircle, AlertCircle, Loader2 } from 'lucide-react';

interface RecentUploadsProps {
  uploads: Upload[];
  onSelectUpload: (upload: Upload) => void;
  isLoading: boolean;
}

const StatusIcon = ({ status }: { status: Upload['status'] }) => {
  switch (status) {
    case 'pending':
      return <Clock className="h-4 w-4 text-yellow-500" />;
    case 'processing':
      return <Loader2 className="h-4 w-4 text-blue-500 animate-spin" />;
    case 'completed':
      return <CheckCircle className="h-4 w-4 text-green-500" />;
    case 'failed':
      return <AlertCircle className="h-4 w-4 text-red-500" />;
    default:
      return <FileText className="h-4 w-4 text-gray-400" />;
  }
};

const StatusBadge = ({ status }: { status: Upload['status'] }) => {
  const variants = {
    pending: 'secondary' as const,
    processing: 'default' as const,
    completed: 'default' as const,
    failed: 'destructive' as const,
  };

  return (
    <Badge variant={variants[status]} className="capitalize">
      {status}
    </Badge>
  );
};

export function RecentUploads({ uploads, onSelectUpload, isLoading }: RecentUploadsProps) {
  if (isLoading) {
    return (
      <div className="space-y-4">
        {[...Array(3)].map((_, i) => (
          <div key={i} className="flex items-center space-x-4 p-4 border rounded-lg animate-pulse">
            <div className="w-10 h-10 bg-gray-200 rounded"></div>
            <div className="flex-1 space-y-2">
              <div className="h-4 bg-gray-200 rounded w-1/3"></div>
              <div className="h-3 bg-gray-200 rounded w-1/4"></div>
            </div>
            <div className="w-20 h-6 bg-gray-200 rounded"></div>
          </div>
        ))}
      </div>
    );
  }

  if (uploads.length === 0) {
    return (
      <div className="text-center py-8">
        <FileText className="h-12 w-12 text-gray-400 mx-auto mb-4" />
        <h3 className="text-lg font-medium text-gray-900 mb-2">No uploads yet</h3>
        <p className="text-gray-600">
          Upload your first age analysis file to get started
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {uploads.map((upload) => (
        <div
          key={upload.id}
          className="flex items-center justify-between p-4 border rounded-lg hover:bg-gray-50 transition-colors"
        >
          <div className="flex items-center space-x-4">
            <StatusIcon status={upload.status} />
            <div>
              <h4 className="font-medium text-gray-900">{upload.filename}</h4>
              <div className="flex items-center space-x-4 text-sm text-gray-600">
                <span>{new Date(upload.created_at).toLocaleString()}</span>
                {upload.processed_count && upload.total_count && (
                  <span>{upload.processed_count}/{upload.total_count} accounts</span>
                )}
              </div>
            </div>
          </div>
          
          <div className="flex items-center space-x-3">
            <StatusBadge status={upload.status} />
            {upload.status === 'completed' && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => onSelectUpload(upload)}
              >
                View Details
              </Button>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}