import { useCallback, useState } from 'react';
import { useDropzone } from 'react-dropzone';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Progress } from '@/components/ui/progress';
import { useToast } from '@/hooks/use-toast';
import { api } from '@/lib/api';
import { Upload, FileSpreadsheet, CheckCircle, AlertCircle, X } from 'lucide-react';
import { cn } from '@/lib/utils';

interface FileUploadProps {
  onUploadComplete?: (uploadId: string) => void;
}

export function FileUpload({ onUploadComplete }: FileUploadProps) {
  const [uploadProgress, setUploadProgress] = useState(0);
  const [uploadStatus, setUploadStatus] = useState<'idle' | 'uploading' | 'success' | 'error'>('idle');
  const [errorMessage, setErrorMessage] = useState<string>('');
  
  const queryClient = useQueryClient();
  const { toast } = useToast();

  const uploadMutation = useMutation({
    mutationFn: (file: File) => api.uploads.create(file),
    onSuccess: (response) => {
      setUploadStatus('success');
      setUploadProgress(100);
      queryClient.invalidateQueries({ queryKey: ['uploads'] });
      
      toast({
        title: "Upload successful!",
        description: "Your file is being processed and will appear in recent uploads shortly.",
      });
      
      if (onUploadComplete) {
        onUploadComplete(response.data.id);
      }
      // Reset after 3 seconds
      setTimeout(() => {
        setUploadStatus('idle');
        setUploadProgress(0);
      }, 3000);
    },
    onError: (error: { response?: { data?: { message?: string } } }) => {
      setUploadStatus('error');
      setErrorMessage(error.response?.data?.message || 'Upload failed');
      setUploadProgress(0);
      
      toast({
        variant: "destructive",
        title: "Upload failed",
        description: error.response?.data?.message || 'An error occurred while uploading your file.',
      });
    },
  });

  const onDrop = useCallback((acceptedFiles: File[]) => {
    const file = acceptedFiles[0];
    if (file) {
      setUploadStatus('uploading');
      setUploadProgress(0);
      setErrorMessage('');
      
      // Simulate progress for better UX
      const progressInterval = setInterval(() => {
        setUploadProgress(prev => {
          if (prev >= 90) {
            clearInterval(progressInterval);
            return 90;
          }
          return prev + 10;
        });
      }, 200);

      uploadMutation.mutate(file);
    }
  }, [uploadMutation]);

  const { getRootProps, getInputProps, isDragActive, fileRejections } = useDropzone({
    onDrop,
    accept: {
      'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet': ['.xlsx']
    },
    maxFiles: 1,
    maxSize: 10 * 1024 * 1024, // 10MB
  });

  const resetUpload = () => {
    setUploadStatus('idle');
    setUploadProgress(0);
    setErrorMessage('');
  };

  return (
    <div className="space-y-4">
      <div
        {...getRootProps()}
        className={cn(
          "border-2 border-dashed rounded-lg p-6 text-center cursor-pointer transition-colors",
          isDragActive 
            ? "border-blue-400 bg-blue-50" 
            : "border-gray-300 hover:border-gray-400",
          uploadStatus === 'success' && "border-green-400 bg-green-50",
          uploadStatus === 'error' && "border-red-400 bg-red-50"
        )}
      >
        <input {...getInputProps()} />
        
        {uploadStatus === 'idle' && (
          <>
            <div className="mx-auto w-12 h-12 text-gray-400 mb-4">
              <FileSpreadsheet className="w-full h-full" />
            </div>
            <div>
              <p className="text-lg font-medium text-gray-900 mb-1">
                {isDragActive ? "Drop your file here" : "Upload age analysis file"}
              </p>
              <p className="text-sm text-gray-600">
                Drag and drop an Excel (.xlsx) file, or <span className="text-blue-600 font-medium">browse</span>
              </p>
              <p className="text-xs text-gray-500 mt-2">
                Maximum file size: 10MB
              </p>
            </div>
          </>
        )}

        {uploadStatus === 'uploading' && (
          <div className="space-y-4">
            <Upload className="w-8 h-8 text-blue-500 mx-auto animate-pulse" />
            <div>
              <p className="font-medium text-gray-900">Uploading...</p>
              <div className="mt-2 max-w-xs mx-auto">
                <Progress value={uploadProgress} className="h-2" />
                <p className="text-xs text-gray-500 mt-1">{uploadProgress}% complete</p>
              </div>
            </div>
          </div>
        )}

        {uploadStatus === 'success' && (
          <div className="space-y-2">
            <CheckCircle className="w-8 h-8 text-green-500 mx-auto" />
            <div>
              <p className="font-medium text-green-700">Upload successful!</p>
              <p className="text-sm text-green-600">
                Your file is being processed and will appear in recent uploads shortly.
              </p>
            </div>
          </div>
        )}

        {uploadStatus === 'error' && (
          <div className="space-y-2">
            <AlertCircle className="w-8 h-8 text-red-500 mx-auto" />
            <div>
              <p className="font-medium text-red-700">Upload failed</p>
              <p className="text-sm text-red-600">{errorMessage}</p>
            </div>
            <Button 
              variant="outline" 
              size="sm" 
              onClick={resetUpload}
              className="mt-2"
            >
              Try again
            </Button>
          </div>
        )}
      </div>

      {fileRejections.length > 0 && (
        <div className="rounded-md bg-red-50 p-4">
          <div className="flex items-start">
            <X className="h-5 w-5 text-red-400 mt-0.5" />
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">File rejected</h3>
              <div className="mt-2 text-sm text-red-700">
                <ul className="list-disc pl-5 space-y-1">
                  {fileRejections.map((rejection, index) => (
                    <li key={index}>
                      {rejection.errors.map(error => error.message).join(', ')}
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}