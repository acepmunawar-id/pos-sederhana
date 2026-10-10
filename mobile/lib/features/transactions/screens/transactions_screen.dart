import 'dart:async';

import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/network/api_exception.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/utils/format.dart';
import '../../../core/widgets/app_scaffold.dart';
import '../../../core/widgets/feedback.dart';
import '../../../core/widgets/inputs.dart';
import '../models/transaction.dart';
import '../providers/transaction_providers.dart';

class TransactionsScreen extends ConsumerStatefulWidget {
  const TransactionsScreen({super.key});

  @override
  ConsumerState<TransactionsScreen> createState() => _TransactionsScreenState();
}

class _TransactionsScreenState extends ConsumerState<TransactionsScreen> {
  TxQuery _query = const TxQuery();
  Timer? _debounce;

  @override
  void dispose() {
    _debounce?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final list = ref.watch(transactionListProvider(_query));

    return NotificationListener<ScrollNotification>(
      onNotification: (n) {
        if (n.metrics.extentAfter < 300) ref.read(transactionListProvider(_query).notifier).loadMore();
        return false;
      },
      child: LargeTitleScaffold(
        title: 'Transaksi',
        onRefresh: () async {
          ref.invalidate(transactionListProvider(_query));
          await ref.read(transactionListProvider(_query).future);
        },
        slivers: [
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(AppSpacing.lg, AppSpacing.sm, AppSpacing.lg, 0),
              child: Column(children: [
                CupertinoSearchTextField(
                  placeholder: 'Cari nomor invoice',
                  backgroundColor: c.fill,
                  style: AppText.body.copyWith(color: c.text),
                  onChanged: (v) {
                    _debounce?.cancel();
                    _debounce = Timer(const Duration(milliseconds: 300), () => setState(() => _query = TxQuery(search: v.trim(), period: _query.period)));
                  },
                ),
                const SizedBox(height: AppSpacing.md),
                AppSegmented<TxPeriod>(
                  value: _query.period,
                  segments: const {TxPeriod.today: 'Hari ini', TxPeriod.week: '7 hari', TxPeriod.all: 'Semua'},
                  onChanged: (p) => setState(() => _query = TxQuery(search: _query.search, period: p)),
                ),
              ]),
            ),
          ),
          ...list.when(
            loading: () => [
              SliverList.builder(
                itemCount: 6,
                itemBuilder: (_, __) => const Padding(
                  padding: EdgeInsets.fromLTRB(AppSpacing.lg, AppSpacing.md, AppSpacing.lg, 0),
                  child: Skeleton(height: 64, radius: AppRadius.sm),
                ),
              ),
            ],
            error: (e, _) => [
              SliverToBoxAdapter(child: ErrorState(message: errorMessage(e), onRetry: () => ref.invalidate(transactionListProvider(_query)))),
            ],
            data: (s) => s.items.isEmpty
                ? [
                    const SliverToBoxAdapter(
                      child: EmptyState(icon: CupertinoIcons.doc_text, title: 'Belum ada transaksi', message: 'Transaksi yang Anda buat akan tampil di sini.'),
                    ),
                  ]
                : [
                    SliverPadding(
                      padding: const EdgeInsets.fromLTRB(AppSpacing.lg, AppSpacing.lg, AppSpacing.lg, 0),
                      sliver: SliverToBoxAdapter(
                        child: Text(
                          '${formatNumber(s.total)} transaksi',
                          style: AppText.footnote.copyWith(color: c.textSecondary),
                        ),
                      ),
                    ),
                    SliverPadding(
                      padding: const EdgeInsets.fromLTRB(AppSpacing.lg, 6, AppSpacing.lg, 0),
                      sliver: SliverToBoxAdapter(
                        child: ClipRRect(
                          borderRadius: BorderRadius.circular(AppRadius.sm),
                          child: ColoredBox(
                            color: c.card,
                            child: Column(children: [
                              for (final (i, t) in s.items.indexed) ...[
                                if (i > 0) Padding(padding: const EdgeInsets.only(left: AppSpacing.lg), child: Container(height: 0.5, color: c.separator)),
                                _TxRow(tx: t, onTap: () => context.push('/transaksi/${t.id}')),
                              ],
                            ]),
                          ),
                        ),
                      ),
                    ),
                    if (s.loadingMore)
                      const SliverToBoxAdapter(child: Padding(padding: EdgeInsets.all(AppSpacing.lg), child: CupertinoActivityIndicator())),
                  ],
          ),
        ],
      ),
    );
  }
}

class _TxRow extends StatelessWidget {
  const _TxRow({required this.tx, required this.onTap});
  final Transaction tx;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: AppSpacing.lg, vertical: AppSpacing.md),
          child: Row(children: [
            Expanded(
              child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                Text(tx.invoiceNumber, style: AppText.headline.copyWith(color: c.text)),
                const SizedBox(height: 2),
                Text('${formatDateTime(tx.createdAt)} · ${tx.paymentLabel}', style: AppText.footnote.copyWith(color: c.textSecondary)),
              ]),
            ),
            Column(crossAxisAlignment: CrossAxisAlignment.end, children: [
              Text(
                formatRupiah(tx.total),
                style: AppText.headline.copyWith(
                  color: tx.isVoided ? c.gray : c.text,
                  decoration: tx.isVoided ? TextDecoration.lineThrough : null,
                ),
              ),
              if (tx.isVoided) Text('Dibatalkan', style: AppText.caption.copyWith(color: c.danger, fontWeight: FontWeight.w600)),
            ]),
            const SizedBox(width: 6),
            Icon(CupertinoIcons.chevron_forward, size: 18, color: c.gray.withValues(alpha: 0.6)),
          ]),
        ),
      ),
    );
  }
}
