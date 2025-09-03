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
  MessageSquare, 
  Search, 
  ArrowUpDown, 
  CheckSquare, 
  Square,
  Loader2,
  DollarSign
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

  const { data: accounts = [], isLoading } = useQuery({
    queryKey: ['accounts', uploadId],
    queryFn: () => api.accounts.getByUploadId(uploadId).then(res => res.data),
    enabled: !!uploadId,
  });

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
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
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

  return (
    <div className="space-y-6">
      {/* Summary Card */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Accounts</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{accounts.length}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Selected</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{selectedAccounts.size}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Selected Balance</CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatCurrency(totalBalance)}</div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="flex items-center justify-center h-full">
            <Button
              onClick={handleSendMessages}
              disabled={selectedAccounts.size === 0 || sendMessagesMutation.isPending}
              className="w-full"
            >
              {sendMessagesMutation.isPending ? (
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              ) : (
                <MessageSquare className="h-4 w-4 mr-2" />
              )}
              Send Messages ({selectedAccounts.size})
            </Button>
          </CardContent>
        </Card>
      </div>

      {/* Search and Controls */}
      <div className="flex items-center justify-between space-x-4">
        <div className="flex items-center space-x-2">
          <Search className="h-4 w-4 text-gray-400" />
          <Input
            placeholder="Search by name, account code, or phone..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-64"
          />
        </div>

        <Button
          variant="outline"
          onClick={handleSelectAll}
          className="flex items-center space-x-2"
        >
          {selectedAccounts.size === filteredAndSortedAccounts.length ? (
            <CheckSquare className="h-4 w-4" />
          ) : (
            <Square className="h-4 w-4" />
          )}
          <span>
            {selectedAccounts.size === filteredAndSortedAccounts.length 
              ? 'Deselect All' 
              : 'Select All'
            }
          </span>
        </Button>
      </div>

      {/* Accounts Table */}
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-12">
                  <Checkbox
                    checked={selectedAccounts.size === filteredAndSortedAccounts.length && filteredAndSortedAccounts.length > 0}
                    onCheckedChange={handleSelectAll}
                  />
                </TableHead>
                <TableHead 
                  className="cursor-pointer hover:bg-gray-50"
                  onClick={() => handleSort('account_code')}
                >
                  <div className="flex items-center space-x-1">
                    <span>Account Code</span>
                    <ArrowUpDown className="h-4 w-4" />
                  </div>
                </TableHead>
                <TableHead 
                  className="cursor-pointer hover:bg-gray-50"
                  onClick={() => handleSort('customer_name')}
                >
                  <div className="flex items-center space-x-1">
                    <span>Customer Name</span>
                    <ArrowUpDown className="h-4 w-4" />
                  </div>
                </TableHead>
                <TableHead>Contact</TableHead>
                <TableHead>Phone</TableHead>
                <TableHead 
                  className="text-right cursor-pointer hover:bg-gray-50"
                  onClick={() => handleSort('amount_current')}
                >
                  <div className="flex items-center justify-end space-x-1">
                    <span>Current</span>
                    <ArrowUpDown className="h-4 w-4" />
                  </div>
                </TableHead>
                <TableHead 
                  className="text-right cursor-pointer hover:bg-gray-50"
                  onClick={() => handleSort('amount_30d')}
                >
                  <div className="flex items-center justify-end space-x-1">
                    <span>30 Days</span>
                    <ArrowUpDown className="h-4 w-4" />
                  </div>
                </TableHead>
                <TableHead 
                  className="text-right cursor-pointer hover:bg-gray-50"
                  onClick={() => handleSort('amount_60d')}
                >
                  <div className="flex items-center justify-end space-x-1">
                    <span>60 Days</span>
                    <ArrowUpDown className="h-4 w-4" />
                  </div>
                </TableHead>
                <TableHead 
                  className="text-right cursor-pointer hover:bg-gray-50"
                  onClick={() => handleSort('amount_90d')}
                >
                  <div className="flex items-center justify-end space-x-1">
                    <span>90 Days</span>
                    <ArrowUpDown className="h-4 w-4" />
                  </div>
                </TableHead>
                <TableHead 
                  className="text-right cursor-pointer hover:bg-gray-50"
                  onClick={() => handleSort('amount_120d')}
                >
                  <div className="flex items-center justify-end space-x-1">
                    <span>120+ Days</span>
                    <ArrowUpDown className="h-4 w-4" />
                  </div>
                </TableHead>
                <TableHead 
                  className="text-right cursor-pointer hover:bg-gray-50"
                  onClick={() => handleSort('total_balance')}
                >
                  <div className="flex items-center justify-end space-x-1">
                    <span>Total Balance</span>
                    <ArrowUpDown className="h-4 w-4" />
                  </div>
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredAndSortedAccounts.map((account: Account) => (
                <TableRow 
                  key={account.id}
                  className={selectedAccounts.has(account.id) ? 'bg-blue-50' : ''}
                >
                  <TableCell>
                    <Checkbox
                      checked={selectedAccounts.has(account.id)}
                      onCheckedChange={() => handleSelectAccount(account.id)}
                    />
                  </TableCell>
                  <TableCell className="font-medium">{account.account_code}</TableCell>
                  <TableCell>{account.customer_name}</TableCell>
                  <TableCell>
                    {account.contact_person && (
                      <div className="text-sm text-gray-600">{account.contact_person}</div>
                    )}
                  </TableCell>
                  <TableCell>{account.telephone}</TableCell>
                  <TableCell className="text-right">
                    {formatCurrency(account.amount_current)}
                  </TableCell>
                  <TableCell className="text-right">
                    {account.amount_30d > 0 && (
                      <Badge variant="secondary">{formatCurrency(account.amount_30d)}</Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-right">
                    {account.amount_60d > 0 && (
                      <Badge variant="secondary">{formatCurrency(account.amount_60d)}</Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-right">
                    {account.amount_90d > 0 && (
                      <Badge variant="destructive">{formatCurrency(account.amount_90d)}</Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-right">
                    {account.amount_120d > 0 && (
                      <Badge variant="destructive">{formatCurrency(account.amount_120d)}</Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-right font-semibold">
                    {formatCurrency(account.total_balance)}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>

          {filteredAndSortedAccounts.length === 0 && (
            <div className="text-center py-8 text-gray-500">
              {searchTerm ? 'No accounts match your search criteria.' : 'No accounts found.'}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}