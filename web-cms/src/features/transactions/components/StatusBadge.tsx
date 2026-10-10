import { Badge } from '@/components/ui';
import type { TransactionStatus } from '../types';

export function TransactionStatusBadge({ status }: { status: TransactionStatus }) {
  return status === 'COMPLETED' ? <Badge tone="success">Selesai</Badge> : <Badge tone="danger">Dibatalkan</Badge>;
}
