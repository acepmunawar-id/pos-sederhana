import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../data/transaction_repository.dart';
import '../models/transaction.dart';

enum TxPeriod { today, week, all }

class TxQuery {
  const TxQuery({this.search = '', this.period = TxPeriod.today});
  final String search;
  final TxPeriod period;

  @override
  bool operator ==(Object other) => other is TxQuery && other.search == search && other.period == period;

  @override
  int get hashCode => Object.hash(search, period);
}

class TxListState {
  const TxListState({required this.items, required this.total, required this.page, this.loadingMore = false});
  final List<Transaction> items;
  final int total;
  final int page;
  final bool loadingMore;
  bool get hasMore => items.length < total;
}

class TransactionListController extends FamilyAsyncNotifier<TxListState, TxQuery> {
  String? get _from {
    final now = DateTime.now();
    return switch (arg.period) {
      TxPeriod.today => DateFormat('yyyy-MM-dd').format(now),
      TxPeriod.week => DateFormat('yyyy-MM-dd').format(now.subtract(const Duration(days: 6))),
      TxPeriod.all => null,
    };
  }

  @override
  Future<TxListState> build(TxQuery query) async {
    final res = await ref.read(transactionRepositoryProvider).list(search: query.search, dateFrom: _from);
    return TxListState(items: res.items, total: res.meta.total, page: 1);
  }

  Future<void> loadMore() async {
    final s = state.valueOrNull;
    if (s == null || !s.hasMore || s.loadingMore) return;
    state = AsyncData(TxListState(items: s.items, total: s.total, page: s.page, loadingMore: true));
    try {
      final res = await ref.read(transactionRepositoryProvider).list(search: arg.search, dateFrom: _from, page: s.page + 1);
      state = AsyncData(TxListState(items: [...s.items, ...res.items], total: res.meta.total, page: s.page + 1));
    } catch (_) {
      state = AsyncData(s);
    }
  }
}

final transactionListProvider =
    AsyncNotifierProvider.family<TransactionListController, TxListState, TxQuery>(TransactionListController.new);

final transactionDetailProvider = FutureProvider.family<Transaction, int>((ref, id) => ref.watch(transactionRepositoryProvider).get(id));
