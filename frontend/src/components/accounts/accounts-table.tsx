import { useState, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useToast } from '@/hooks/use-toast';
import { api } from '@/lib/api';
import type { Account } from '@/types';
import { 
  Search, 
  ArrowUpDown, 
  CheckSquare, 
  Square,
  Loader2,
  DollarSign,
  Users,
  Send,
  Filter,
  AlertTriangle,
  Phone,
  Mail
} from 'lucide-react';

interface AccountsTableProps {
  uploadId: string;
}

export function AccountsTable({ uploadId }: AccountsTableProps) {
  const [selectedAccounts, setSelectedAccounts] = useState<Set<string>>(new Set());
  const [searchTerm, setSearchTerm] = useState('');
  const [sortField, setSortField] = useState<keyof Account>('total_balance');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('desc');
  
  const queryClient = useQueryClient();
  const { toast } = useToast();

  const { data: accountsResponse, isLoading, error } = useQuery({
    queryKey: ['accounts', uploadId],
    queryFn: async () => {
      console.log('Fetching accounts for uploadId:', uploadId);
      const response = await api.accounts.getByUploadId(uploadId);
      console.log('Accounts API response:', response.data);
      return response.data;
    },
    enabled: !!uploadId,
  });

  const accounts = accountsResponse?.accounts || [];
  console.log('Processed accounts:', accounts);

  const sendMessagesMutation = useMutation({
    mutationFn: (accountIds: string[]) => api.messaging.send(accountIds),
    onSuccess: () => {
      const sentCount = selectedAccounts.size;
      setSelectedAccounts(new Set());
      queryClient.invalidateQueries({ queryKey: ['message-logs'] });
      
      toast({
        title: "Messages sent successfully!",
        description: `Sent reminders to ${sentCount} accounts. Check the reports page for details.`,
      });
    },
    onError: (error: { response?: { data?: { message?: string } } }) => {
      toast({
        variant: "destructive",
        title: "Failed to send messages",
        description: error.response?.data?.message || 'An error occurred while sending messages.',
      });
    },
  });

  // Filter and sort accounts
  const filteredAndSortedAccounts = useMemo(() => {
    const filtered = accounts.filter((account: Account) =>
      account.customer_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      account.account_code.toLowerCase().includes(searchTerm.toLowerCase()) ||
      account.telephone.includes(searchTerm)
    );

    return filtered.sort((a: Account, b: Account) => {
      const aValue = a[sortField];
      const bValue = b[sortField];
      
      if (typeof aValue === 'string') {
        const comparison = aValue.localeCompare(bValue as string);
        return sortDirection === 'asc' ? comparison : -comparison;
      }
      
      if (typeof aValue === 'number') {
        const comparison = aValue - (bValue as number);
        return sortDirection === 'asc' ? comparison : -comparison;
      }
      
      return 0;
    });
  }, [accounts, searchTerm, sortField, sortDirection]);

  const handleSort = (field: keyof Account) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  const handleSelectAccount = (accountId: string) => {
    const newSelected = new Set(selectedAccounts);
    if (newSelected.has(accountId)) {
      newSelected.delete(accountId);
    } else {
      newSelected.add(accountId);
    }
    setSelectedAccounts(newSelected);
  };

  const handleSelectAll = () => {
    if (selectedAccounts.size === filteredAndSortedAccounts.length) {
      setSelectedAccounts(new Set());
    } else {
      setSelectedAccounts(new Set(filteredAndSortedAccounts.map((account: Account) => account.id)));
    }
  };

  const handleSendMessages = () => {
    if (selectedAccounts.size > 0) {
      sendMessagesMutation.mutate(Array.from(selectedAccounts));
    }
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-ZA', {
      style: 'currency',
      currency: 'ZAR',
    }).format(amount);
  };

  const totalBalance = selectedAccounts.size > 0 
    ? filteredAndSortedAccounts
        .filter((account: Account) => selectedAccounts.has(account.id))
        .reduce((sum: number, account: Account) => sum + account.total_balance, 0)
    : 0;

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  if (error) {
    console.error('Error loading accounts:', error);
    return (
      <div className="flex items-center justify-center h-64 text-red-500">
        <div className="text-center">
          <p>Error loading accounts</p>
          <p className="text-sm text-gray-500">{error.message}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      {/* Action Header */}
      <div className="bg-gradient-to-r from-blue-500 to-cyan-600 rounded-2xl shadow-lg text-white p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <div className="w-12 h-12 bg-white/20 rounded-xl flex items-center justify-center">
              <Users className="h-6 w-6 text-white" />
            </div>
            <div>
              <h2 className="text-2xl font-bold">Account Management</h2>
              <p className="text-blue-100">Manage customer accounts and send collection messages</p>
            </div>
          </div>
          <Button
            onClick={handleSendMessages}
            disabled={selectedAccounts.size === 0 || sendMessagesMutation.isPending}
            className="bg-white/20 hover:bg-white/30 text-white border-white/20 backdrop-blur-sm"
            size="lg"
          >
            {sendMessagesMutation.isPending ? (
              <Loader2 className="h-5 w-5 mr-2 animate-spin" />
            ) : (
              <Send className="h-5 w-5 mr-2" />
            )}
            Send Messages ({selectedAccounts.size})
          </Button>
        </div>
      </div>

      {/* Stats Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <Card className="bg-white border-slate-200/60 shadow-sm hover:shadow-md transition-all duration-200">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
            <div>
              <CardTitle className="text-sm font-medium text-slate-600">Total Accounts</CardTitle>
              <div className="text-2xl font-bold text-slate-800 mt-2">{accounts.length}</div>
            </div>
            <div className="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center">
              <Users className="h-6 w-6 text-blue-600" />
            </div>
          </CardHeader>
        </Card>

        <Card className="bg-white border-slate-200/60 shadow-sm hover:shadow-md transition-all duration-200">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
            <div>
              <CardTitle className="text-sm font-medium text-slate-600">Selected</CardTitle>
              <div className="text-2xl font-bold text-cyan-600 mt-2">{selectedAccounts.size}</div>
            </div>
            <div className="w-12 h-12 bg-cyan-100 rounded-xl flex items-center justify-center">
              <CheckSquare className="h-6 w-6 text-cyan-600" />
            </div>
          </CardHeader>
        </Card>

        <Card className="bg-white border-slate-200/60 shadow-sm hover:shadow-md transition-all duration-200">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
            <div>
              <CardTitle className="text-sm font-medium text-slate-600">Selected Value</CardTitle>
              <div className="text-2xl font-bold text-emerald-600 mt-2">{formatCurrency(totalBalance)}</div>
            </div>
            <div className="w-12 h-12 bg-emerald-100 rounded-xl flex items-center justify-center">
              <DollarSign className="h-6 w-6 text-emerald-600" />
            </div>
          </CardHeader>
        </Card>

        <Card className="bg-white border-slate-200/60 shadow-sm hover:shadow-md transition-all duration-200">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
            <div>
              <CardTitle className="text-sm font-medium text-slate-600">Overdue Accounts</CardTitle>
              <div className="text-2xl font-bold text-orange-600 mt-2">
                {accounts.filter((acc: Account) => (acc.amount_30d + acc.amount_60d + acc.amount_90d + acc.amount_120d) > 0).length}
              </div>
            </div>
            <div className="w-12 h-12 bg-orange-100 rounded-xl flex items-center justify-center">
              <AlertTriangle className="h-6 w-6 text-orange-600" />
            </div>
          </CardHeader>
        </Card>
      </div>

      {/* Search and Controls */}
      <Card className="bg-white border-slate-200/60 shadow-sm">
        <CardHeader className="pb-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-slate-400" />
                <Input
                  placeholder="Search by name, account code, or phone..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10 w-80 border-slate-200 focus:border-blue-500 focus:ring-blue-500/20"
                />
              </div>
              <Badge variant="secondary" className="bg-slate-100 text-slate-600">
                {filteredAndSortedAccounts.length} accounts
              </Badge>
            </div>

            <div className="flex items-center space-x-3">
              <Button
                variant="outline"
                onClick={handleSelectAll}
                className="border-slate-200 hover:bg-slate-50"
                size="sm"
              >
                {selectedAccounts.size === filteredAndSortedAccounts.length ? (
                  <CheckSquare className="h-4 w-4 mr-2 text-cyan-600" />
                ) : (
                  <Square className="h-4 w-4 mr-2" />
                )}
                {selectedAccounts.size === filteredAndSortedAccounts.length 
                  ? 'Deselect All' 
                  : 'Select All'
                }
              </Button>
              <Button variant="outline" size="sm" className="border-slate-200 hover:bg-slate-50">
                <Filter className="h-4 w-4 mr-2" />
                Filter
              </Button>
            </div>
          </div>
        </CardHeader>
      </Card>

      {/* Accounts Table */}
      <Card className="bg-white border-slate-200/60 shadow-sm overflow-hidden">
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <Table>
            <TableHeader>
              <TableRow className="bg-slate-50/50 hover:bg-slate-50/50 border-b border-slate-200">
                <TableHead className="w-12 pl-6">
                  <Checkbox
                    checked={selectedAccounts.size === filteredAndSortedAccounts.length && filteredAndSortedAccounts.length > 0}
                    onCheckedChange={handleSelectAll}
                    className="border-slate-300"
                  />
                </TableHead>
                <TableHead 
                  className="cursor-pointer hover:bg-slate-100/50 transition-colors font-semibold text-slate-700"
                  onClick={() => handleSort('account_code')}
                >
                  <div className="flex items-center space-x-2">
                    <span>Account Code</span>
                    <ArrowUpDown className="h-4 w-4 text-slate-400" />
                  </div>
                </TableHead>
                <TableHead 
                  className="cursor-pointer hover:bg-slate-100/50 transition-colors font-semibold text-slate-700"
                  onClick={() => handleSort('customer_name')}
                >
                  <div className="flex items-center space-x-2">
                    <span>Customer Details</span>
                    <ArrowUpDown className="h-4 w-4 text-slate-400" />
                  </div>
                </TableHead>
                <TableHead className="font-semibold text-slate-700">Contact Info</TableHead>
                <TableHead 
                  className="text-right cursor-pointer hover:bg-slate-100/50 transition-colors font-semibold text-slate-700"
                  onClick={() => handleSort('amount_current')}
                >
                  <div className="flex items-center justify-end space-x-2">
                    <span>Current</span>
                    <ArrowUpDown className="h-4 w-4 text-slate-400" />
                  </div>
                </TableHead>
                <TableHead 
                  className="text-right cursor-pointer hover:bg-slate-100/50 transition-colors font-semibold text-slate-700"
                  onClick={() => handleSort('amount_30d')}
                >
                  <div className="flex items-center justify-end space-x-2">
                    <span>30 Days</span>
                    <ArrowUpDown className="h-4 w-4 text-slate-400" />
                  </div>
                </TableHead>
                <TableHead 
                  className="text-right cursor-pointer hover:bg-slate-100/50 transition-colors font-semibold text-slate-700"
                  onClick={() => handleSort('amount_60d')}
                >
                  <div className="flex items-center justify-end space-x-2">
                    <span>60 Days</span>
                    <ArrowUpDown className="h-4 w-4 text-slate-400" />
                  </div>
                </TableHead>
                <TableHead 
                  className="text-right cursor-pointer hover:bg-slate-100/50 transition-colors font-semibold text-slate-700"
                  onClick={() => handleSort('amount_90d')}
                >
                  <div className="flex items-center justify-end space-x-2">
                    <span>90 Days</span>
                    <ArrowUpDown className="h-4 w-4 text-slate-400" />
                  </div>
                </TableHead>
                <TableHead 
                  className="text-right cursor-pointer hover:bg-slate-100/50 transition-colors font-semibold text-slate-700"
                  onClick={() => handleSort('amount_120d')}
                >
                  <div className="flex items-center justify-end space-x-2">
                    <span>120+ Days</span>
                    <ArrowUpDown className="h-4 w-4 text-slate-400" />
                  </div>
                </TableHead>
                <TableHead 
                  className="text-right cursor-pointer hover:bg-slate-100/50 transition-colors font-semibold text-slate-700 pr-6"
                  onClick={() => handleSort('total_balance')}
                >
                  <div className="flex items-center justify-end space-x-2">
                    <span>Total Balance</span>
                    <ArrowUpDown className="h-4 w-4 text-slate-400" />
                  </div>
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredAndSortedAccounts.map((account: Account) => {
                const isOverdue = (account.amount_30d + account.amount_60d + account.amount_90d + account.amount_120d) > 0;
                const isSelected = selectedAccounts.has(account.id);
                
                return (
                  <TableRow 
                    key={account.id}
                    className={`border-b border-slate-100 hover:bg-slate-50/50 transition-colors ${
                      isSelected ? 'bg-blue-50/50 hover:bg-blue-50' : ''
                    }`}
                  >
                    <TableCell className="pl-6">
                      <Checkbox
                        checked={isSelected}
                        onCheckedChange={() => handleSelectAccount(account.id)}
                        className="border-slate-300"
                      />
                    </TableCell>
                    
                    <TableCell className="py-4">
                      <div className="flex items-center space-x-2">
                        <span className="font-semibold text-slate-800">{account.account_code}</span>
                        {isOverdue && (
                          <Badge className="bg-orange-100 text-orange-700 hover:bg-orange-100 text-xs">
                            Overdue
                          </Badge>
                        )}
                      </div>
                    </TableCell>
                    
                    <TableCell className="py-4">
                      <div>
                        <div className="font-medium text-slate-800">{account.customer_name}</div>
                        {account.contact_person && (
                          <div className="text-sm text-slate-500 flex items-center mt-1">
                            <Mail className="h-3 w-3 mr-1" />
                            {account.contact_person}
                          </div>
                        )}
                      </div>
                    </TableCell>
                    
                    <TableCell className="py-4">
                      <div className="flex items-center text-slate-600">
                        <Phone className="h-4 w-4 mr-2 text-slate-400" />
                        {account.telephone}
                      </div>
                    </TableCell>
                    
                    <TableCell className="text-right py-4">
                      <span className="font-medium text-slate-800">
                        {formatCurrency(account.amount_current)}
                      </span>
                    </TableCell>
                    
                    <TableCell className="text-right py-4">
                      {account.amount_30d > 0 ? (
                        <Badge className="bg-amber-100 text-amber-700 hover:bg-amber-100">
                          {formatCurrency(account.amount_30d)}
                        </Badge>
                      ) : (
                        <span className="text-slate-400">-</span>
                      )}
                    </TableCell>
                    
                    <TableCell className="text-right py-4">
                      {account.amount_60d > 0 ? (
                        <Badge className="bg-orange-100 text-orange-700 hover:bg-orange-100">
                          {formatCurrency(account.amount_60d)}
                        </Badge>
                      ) : (
                        <span className="text-slate-400">-</span>
                      )}
                    </TableCell>
                    
                    <TableCell className="text-right py-4">
                      {account.amount_90d > 0 ? (
                        <Badge className="bg-red-100 text-red-700 hover:bg-red-100">
                          {formatCurrency(account.amount_90d)}
                        </Badge>
                      ) : (
                        <span className="text-slate-400">-</span>
                      )}
                    </TableCell>
                    
                    <TableCell className="text-right py-4">
                      {account.amount_120d > 0 ? (
                        <Badge className="bg-red-200 text-red-800 hover:bg-red-200">
                          {formatCurrency(account.amount_120d)}
                        </Badge>
                      ) : (
                        <span className="text-slate-400">-</span>
                      )}
                    </TableCell>
                    
                    <TableCell className="text-right font-bold text-slate-900 pr-6 py-4">
                      {formatCurrency(account.total_balance)}
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
          
          {filteredAndSortedAccounts.length === 0 && (
            <div className="text-center py-12 text-slate-500">
              <div className="flex flex-col items-center space-y-3">
                <div className="w-16 h-16 bg-slate-100 rounded-2xl flex items-center justify-center">
                  <Users className="h-8 w-8 text-slate-400" />
                </div>
                <div>
                  <p className="font-medium">
                    {searchTerm ? 'No accounts match your search criteria' : 'No accounts found'}
                  </p>
                  <p className="text-sm text-slate-400 mt-1">
                    {searchTerm ? 'Try adjusting your search terms' : 'Upload a file to get started'}
                  </p>
                </div>
              </div>
            </div>
          )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}